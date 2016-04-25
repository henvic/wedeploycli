package deploy

import (
	"crypto/sha1"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"

	"github.com/dustin/go-humanize"
	"github.com/launchpad-project/api.go"
	"github.com/launchpad-project/cli/apihelper"
	"github.com/launchpad-project/cli/config"
	"github.com/launchpad-project/cli/containers"
	"github.com/launchpad-project/cli/hooks"
	"github.com/launchpad-project/cli/pod"
	"github.com/launchpad-project/cli/progress"
	"github.com/launchpad-project/cli/projects"
	"github.com/launchpad-project/cli/verbose"
)

// ContainerError struct
type ContainerError struct {
	Container string
	Error     error
}

// Deploy holds the information of a POD to be packed or deployed
type Deploy struct {
	Container     *containers.Container
	ContainerPath string
	PackageSize   int64
	progress      *progress.Bar
}

// Errors list
type Errors struct {
	List []ContainerError
}

// Flags modifiers
type Flags struct {
	Quiet bool
	Hooks bool
}

var (
	outStream io.Writer = os.Stdout
)

func (de Errors) Error() string {
	var msgs = []string{}

	for _, e := range de.List {
		msgs = append(msgs, fmt.Sprintf("%v: %v", e.Container, e.Error.Error()))
	}

	return fmt.Sprintf("List of errors (format is container: error)\n%v",
		strings.Join(msgs, "\n"))
}

// All deploys a list of containers on the given context
func All(list []string, df *Flags) (err error) {
	var wg sync.WaitGroup
	var mutex sync.Mutex
	var de = &Errors{
		List: []ContainerError{},
	}

	wg.Add(len(list))

	for _, container := range list {
		go func(container string) {
			var err = Only(container, df)

			if err != nil {
				mutex.Lock()
				de.List = append(de.List, ContainerError{
					Container: container,
					Error:     err,
				})
				mutex.Unlock()
			}

			wg.Done()
		}(container)
	}

	wg.Wait()

	if len(de.List) != 0 {
		err = de
	}

	return err
}

// Only PODify a container and deploys it to Launchpad
func Only(container string, df *Flags) error {
	var deploy, err = New(container)

	if err != nil {
		return err
	}

	var projectID = config.Stores["project"].Get("id")

	created, err := projects.ValidateOrCreate(
		projectID,
		config.Stores["project"].Get("name"))

	if created {
		fmt.Fprintf(outStream, "New project %v created\n", projectID)
	}

	if err != nil {
		return err
	}

	created, err = containers.ValidateOrCreate(projectID, deploy.Container)

	if created {
		fmt.Fprintf(outStream, "New container %v created\n", deploy.Container.ID)
	}

	if err != nil {
		return err
	}

	return deploy.HooksAndOnly(df)
}

// New Deploy instance
func New(cpath string) (*Deploy, error) {
	var deploy = &Deploy{
		ContainerPath: path.Join(config.Context.ProjectRoot, cpath),
		progress:      progress.New(cpath),
	}

	var c containers.Container
	var err = containers.GetConfig(deploy.ContainerPath, &c)
	deploy.Container = &c

	if err != nil {
		return nil, err
	}

	return deploy, err
}

// Pack packages a POD to a .pod package
func Pack(dest, cpath string) error {
	var deploy, err = New(cpath)

	if err == nil {
		err = deploy.Pack(dest)
	}

	return err
}

// Deploy POD to Launchpad
func (d *Deploy) Deploy(src string) error {
	var projectID = config.Stores["project"].Get("id")
	var u = path.Join("api/push", projectID, d.Container.ID)
	var req = apihelper.URL(u)
	var fileInfo os.FileInfo

	apihelper.Auth(req)

	var file, err = os.Open(src)

	if err != nil {
		return err
	}

	fileInfo, err = file.Stat()

	if err != nil {
		return err
	}

	var packageSize = fileInfo.Size()

	var wc = &writeCounter{
		progress: d.progress,
		Size:     uint64(packageSize),
	}

	d.progress.Reset("Uploading", "")

	var hash, esha1 = getPackageSHA1(file)

	if esha1 != nil {
		return esha1
	}

	req.Header("Launchpad-Package-Size", fmt.Sprintf("%d", packageSize))
	req.Header("Launchpad-Package-SHA1", hash)

	var r, w = io.Pipe()
	var mpw = multipart.NewWriter(w)

	var errMultipartChan = make(chan error, 1)

	go func() {
		errMultipartChan <- multipartWriter(mpw, w, file, wc)
		close(errMultipartChan)
	}()

	req.Body(r)
	req.Headers.Set("Content-Type", mpw.FormDataContentType())

	err = apihelper.Validate(req, req.Post())

	if errMultipart := <-errMultipartChan; errMultipart != nil {
		return errMultipart
	}

	if err == nil || err == launchpad.ErrUnexpectedResponse {
		d.progress.Append = fmt.Sprintf(
			"%s (Complete)",
			humanize.Bytes(uint64(packageSize)))
		d.progress.Set(progress.Total)
	}

	if err == nil {
		fmt.Fprintf(outStream, "Ready! %v.%v.liferay.io\n", d.Container.ID, projectID)
	}

	return err
}

// HooksAndOnly run the hooks and Only method
func (d *Deploy) HooksAndOnly(df *Flags) (err error) {
	var ch = d.Container.Hooks

	if df.Hooks && ch != nil && ch.BeforeDeploy != "" {
		err = hooks.Run(ch.BeforeDeploy)
	}

	if err == nil {
		err = d.Only()
	}

	if err == nil && df.Hooks && ch != nil && ch.AfterDeploy != "" {
		err = hooks.Run(ch.AfterDeploy)
	}

	return err
}

// Only PODify a container and deploys it to Launchpad
func (d *Deploy) Only() error {
	var tmp, err = ioutil.TempFile(os.TempDir(), "launchpad-cli")

	err = d.Pack(tmp.Name())

	if err == nil {
		err = d.Deploy(tmp.Name())
	}

	if tmp != nil {
		os.Remove(tmp.Name())
	}

	return err
}

// Pack packages a POD to a .pod package
func (d *Deploy) Pack(dest string) (err error) {
	d.progress.Reset("Packing", "")
	dest, _ = filepath.Abs(dest)

	var ignorePatterns = append(d.Container.DeployIgnore, pod.CommonIgnorePatterns...)

	_, err = pod.Pack(dest, d.ContainerPath, ignorePatterns, d.progress)

	if err == nil {
		d.progress.Set(progress.Total)
	}

	verbose.Debug("Saving container to", dest)

	return err
}

func getPackageSHA1(file *os.File) (string, error) {
	var hash = sha1.New()
	var _, err = io.Copy(hash, file)

	if err != nil {
		return "", err
	}

	_, err = file.Seek(0, 0)

	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", hash.Sum(nil)), err
}

func multipartWriter(
	mpw *multipart.Writer,
	w io.WriteCloser,
	file *os.File,
	wc *writeCounter) (err error) {
	var part io.Writer
	defer w.Close()
	defer file.Close()

	if part, err = mpw.CreateFormFile("pod", "container.pod"); err != nil {
		return err
	}

	part = io.MultiWriter(part, wc)
	if _, err = io.Copy(part, file); err != nil {
		return err
	}

	return mpw.Close()
}
