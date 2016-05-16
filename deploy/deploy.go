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

	"github.com/dustin/go-humanize"
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
	ContainerPath string
	Error         error
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
		msgs = append(msgs, fmt.Sprintf("%v: %v", e.ContainerPath, e.Error.Error()))
	}

	return fmt.Sprintf("List of errors (format is container path: error)\n%v",
		strings.Join(msgs, "\n"))
}

// All deploys a list of containers on the given context
func All(list []string, df *Flags) (success []string, err error) {
	var projectID = config.Stores["project"].Get("id")

	var dm = &Machine{
		ProjectID: projectID,
		Flags:     df,
	}

	created, err := projects.ValidateOrCreate(
		filepath.Join(config.Context.ProjectRoot, "/project.json"))

	if created {
		dm.Success = append(dm.Success, "New project "+projectID+" created")
	}

	if err != nil {
		return success, err
	}

	err = dm.Run(list)

	success = dm.Success

	return success, err
}

// Only PODify a single container and deploys it to WeDeploy
func Only(container string, df *Flags) error {
	var deploy, err = New(container)

	if err != nil {
		return err
	}

	var projectID = config.Stores["project"].Get("id")

	created, err := projects.ValidateOrCreate(
		filepath.Join(config.Context.ProjectRoot, "/project.json"))

	if created {
		fmt.Fprintf(outStream, "New project %v created\n", projectID)
	}

	if err != nil {
		return err
	}

	err = containers.InstallFromDefinition(projectID,
		deploy.ContainerPath,
		deploy.Container)

	if err != nil {
		return err
	}

	verbose.Debug(deploy.Container.ID, "container definition installed")

	return deploy.HooksAndOnly(df)
}

// Pack packages a POD to a .pod package
func Pack(dest, cpath string) error {
	var deploy, err = New(cpath)

	if err == nil {
		err = deploy.Pack(dest)
	}

	return err
}

// Deploy POD to WeDeploy
func (d *Deploy) Deploy(src string) error {
	var projectID = config.Stores["project"].Get("id")
	var u = path.Join("push", projectID, d.Container.ID)
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

	req.Header("WeDeploy-Package-Size", fmt.Sprintf("%d", packageSize))
	req.Header("WeDeploy-Package-SHA1", hash)

	var r, w = io.Pipe()
	var mpw = multipart.NewWriter(w)

	var errMultipartChan = make(chan error, 1)

	go func() {
		errMultipartChan <- multipartWriter(mpw, w, file)
		close(errMultipartChan)
	}()

	req.Body(io.TeeReader(r, wc))

	req.Headers.Set("Content-Type", mpw.FormDataContentType())

	err = apihelper.Validate(req, req.Post())

	var errMultipart = <-errMultipartChan

	if err != nil || errMultipart != nil {
		d.progress.Append = "(Failure)"
		d.progress.Fail()
	}

	if err != nil && errMultipart != nil {
		verbose.Debug("Error both in multipart and error")
		verbose.Debug("Error:")
		verbose.Debug(err.Error())
		verbose.Debug("Multipart error:")
		verbose.Debug(errMultipart.Error())
	}

	if errMultipart != nil {
		return errMultipart
	}

	if err != nil {
		return err
	}

	d.progress.Append = fmt.Sprintf(
		"%s (Complete)",
		humanize.Bytes(uint64(packageSize)))
	d.progress.Set(progress.Total)

	return err
}

// HooksAndOnly run the hooks and Only method
func (d *Deploy) HooksAndOnly(df *Flags) (err error) {
	var ch = d.Container.Hooks
	var workingDir string

	workingDir, err = os.Getwd()

	if err != nil {
		return err
	}

	err = os.Chdir(d.ContainerPath)

	if err != nil {
		return err
	}

	if df.Hooks && ch != nil && ch.BeforeDeploy != "" {
		err = hooks.Run(ch.BeforeDeploy)
	}

	err = os.Chdir(workingDir)

	if err != nil {
		return err
	}

	if err == nil {
		err = d.Only()
	}

	err = os.Chdir(d.ContainerPath)

	if err != nil {
		return err
	}

	if err == nil && df.Hooks && ch != nil && ch.AfterDeploy != "" {
		err = hooks.Run(ch.AfterDeploy)
	}

	err = os.Chdir(workingDir)

	if err != nil {
		return err
	}

	return err
}

// Only PODify a container and deploys it to WeDeploy
func (d *Deploy) Only() error {
	if config.Stores["global"].Get("local") == "true" {
		return nil
	}

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

func getPackageSHA1(file io.ReadSeeker) (string, error) {
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
	w io.Closer,
	file io.ReadCloser) (err error) {
	var part io.Writer
	defer w.Close()
	defer file.Close()

	if part, err = mpw.CreateFormFile("pod", "container.pod"); err != nil {
		return err
	}

	if _, err = io.Copy(part, file); err != nil {
		return err
	}

	return mpw.Close()
}
