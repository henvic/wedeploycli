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
	ContainerPath string
	Error         error
}

// Deploy holds the information of a POD to be packed or deployed
type Deploy struct {
	Container     *containers.Container
	ContainerPath string
	PackageSize   uint64
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
	var request, file, err = d.setupPackage(src)

	switch {
	case err != nil:
		return err
	default:
		return d.deployUpload(request, file, &writeCounter{
			progress: d.progress,
			Size:     d.PackageSize,
		})
	}
}

func (d *Deploy) deployUpload(
	request *launchpad.Launchpad, rc io.ReadCloser, wc io.Writer) error {
	var r, w = io.Pipe()
	var mpw = multipart.NewWriter(w)
	var errMultipartChan = make(chan error, 1)

	go deployWriter(errMultipartChan, mpw, w, rc)
	request.Body(io.TeeReader(r, wc))
	request.Headers.Set("Content-Type", mpw.FormDataContentType())

	return d.deployFeedback(
		apihelper.Validate(request, request.Post()),
		<-errMultipartChan)
}

// HooksAndOnly run the hooks and Only method
func (d *Deploy) HooksAndOnly(df *Flags) error {
	var wdir, err = os.Getwd()

	if err != nil {
		return err
	}

	if err = d.runBeforeHook(df, wdir); err != nil {
		return err
	}

	if err = d.Only(); err != nil {
		return err
	}

	return d.runAfterHook(df, wdir)
}

// Only PODify a container and deploys it to WeDeploy
func (d *Deploy) Only() error {
	if config.Stores["global"].Get("local") == "true" {
		return nil
	}

	return d.only()
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

func chdir(dir string) {
	if ech := os.Chdir(dir); ech != nil {
		panic(ech)
	}
}

func deployWriter(
	emw chan error, mpw *multipart.Writer, w io.Closer, file io.ReadCloser) {
	emw <- multipartWriter(mpw, w, file)
	close(emw)
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

func reportDeployMultipleError(err, errMultipart error) {
	if err != nil && errMultipart != nil {
		verbose.Debug("Error both in multipart and error")
		verbose.Debug("Error:")
		verbose.Debug(err.Error())
		verbose.Debug("Multipart error:")
		verbose.Debug(errMultipart.Error())
	}
}

func (d *Deploy) createDeployRequest(rs io.ReadSeeker) (
	*launchpad.Launchpad, error) {
	var projectID = config.Stores["project"].Get("id")
	var request *launchpad.Launchpad
	var hash, err = getPackageSHA1(rs)

	if err == nil {
		request = apihelper.URL(path.Join("push", projectID, d.Container.ID))
		apihelper.Auth(request)
		request.Header("Package-Size", fmt.Sprintf("%d", d.PackageSize))
		request.Header("Package-SHA1", hash)
	}

	return request, err
}

func (d *Deploy) deployFeedback(err, errMultipart error) error {
	if err != nil || errMultipart != nil {
		d.setProgressFailure()
	}

	reportDeployMultipleError(err, errMultipart)

	switch {
	case errMultipart != nil:
		return errMultipart
	case err != nil:
		return err
	default:
		d.setProgressComplete(d.PackageSize)
		return err
	}
}

func (d *Deploy) getPackageFD(src string) (*os.File, uint64, error) {
	var fileInfo os.FileInfo
	var size uint64
	var file, err = os.Open(src)

	if err != nil {
		return nil, 0, err
	}

	fileInfo, err = file.Stat()

	if err == nil {
		size = uint64(fileInfo.Size())
	}

	return file, size, err
}

func (d *Deploy) only() error {
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

func (d *Deploy) runHook(df *Flags, wdir, path string) error {
	var ch = d.Container.Hooks

	if df.Hooks && ch != nil && path != "" {
		chdir(d.ContainerPath)
		var err = hooks.Run(path)
		chdir(wdir)

		return err
	}

	return nil
}

func (d *Deploy) runBeforeHook(df *Flags, wdir string) error {
	var hooks = d.Container.Hooks
	return d.runHook(df, wdir, hooks.BeforeDeploy)
}

func (d *Deploy) runAfterHook(df *Flags, wdir string) error {
	var hooks = d.Container.Hooks
	return d.runHook(df, wdir, hooks.AfterDeploy)
}

func (d *Deploy) setProgressComplete(size uint64) {
	d.progress.Append = fmt.Sprintf(
		"%s (Complete)",
		humanize.Bytes(size))
	d.progress.Set(progress.Total)
}

func (d *Deploy) setProgressFailure() {
	d.progress.Append = "(Failure)"
	d.progress.Fail()
}

func (d *Deploy) setProgressUploading() {
	d.progress.Reset("Uploading", "")
}

func (d *Deploy) setupPackage(src string) (
	*launchpad.Launchpad, *os.File, error) {
	var file, size, errFD = d.getPackageFD(src)
	d.PackageSize = size

	if errFD != nil {
		return nil, nil, errFD
	}

	d.setProgressUploading()
	var request, err = d.createDeployRequest(file)
	return request, file, err
}
