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

// Deploy holds the information of a POD to be packed or deployed
type Deploy struct {
	Project       *projects.Project
	Container     *containers.Container
	ContainerPath string
	PackageSize   uint64
	progress      *deployProgress
}

// Flags modifiers
type Flags struct {
	Quiet bool
	Hooks bool
}

// Pack packages a POD to a .pod package
func Pack(dest, cpath string) error {
	var deploy, err = New(cpath)

	if err == nil {
		err = deploy.Pack(dest)
	}

	return err
}

// New Deploy instance
func New(cpath string) (*Deploy, error) {
	var deploy = &Deploy{
		ContainerPath: path.Join(config.Context.ProjectRoot, cpath),
		progress:      &deployProgress{progress.New(cpath)},
	}

	p, err := projects.Read(filepath.Join(deploy.ContainerPath, ".."))
	deploy.Project = p

	if err != nil {
		return nil, err
	}

	c, err := containers.Read(deploy.ContainerPath)
	deploy.Container = c

	if err != nil {
		return nil, err
	}

	return deploy, err
}

// Deploy POD to WeDeploy
func (d *Deploy) Deploy(src string) error {
	var request, file, err = d.setupPackage(src)

	switch {
	case err != nil:
		return err
	default:
		return d.deployUpload(request, file, &writeCounter{
			progress: d.progress.bar,
			Size:     d.PackageSize,
		})
	}
}

type deploySubmission struct {
	emc chan error
	mpw *multipart.Writer
	pw  io.Closer
	rc  io.ReadCloser
}

func (ds *deploySubmission) Writer() {
	ds.emc <- multipartWriter(ds.mpw, ds.pw, ds.rc)
	close(ds.emc)
}

func (ds *deploySubmission) Setup(rc io.ReadCloser) *io.PipeReader {
	var pr, pw = io.Pipe()
	ds.rc = rc
	ds.pw = pw
	ds.emc = make(chan error, 1)
	ds.mpw = multipart.NewWriter(pw)
	return pr
}

func (d *Deploy) deployUpload(
	request *launchpad.Launchpad, rc io.ReadCloser, wc io.Writer) error {
	var ds = &deploySubmission{}
	var pr = ds.Setup(rc)

	go ds.Writer()
	request.Body(io.TeeReader(pr, wc))
	request.Headers.Set("Content-Type", ds.mpw.FormDataContentType())

	return d.deployFeedback(
		apihelper.Validate(request, request.Post()),
		<-ds.emc)
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
	if config.Global.Local {
		return nil
	}

	return d.only()
}

// Pack packages a POD to a .pod package
func (d *Deploy) Pack(dest string) (err error) {
	d.progress.setPacking()
	dest, _ = filepath.Abs(dest)

	var ignorePatterns = append(d.Container.DeployIgnore, pod.CommonIgnorePatterns...)

	_, err = pod.Pack(pod.PackParams{
		RelDest:        dest,
		RelSource:      d.ContainerPath,
		IgnorePatterns: ignorePatterns,
	}, d.progress.bar)

	if err == nil {
		d.progress.bar.Set(progress.Total)
	}

	return err
}

func chdir(dir string) {
	if ech := os.Chdir(dir); ech != nil {
		panic(ech)
	}
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
	var request *launchpad.Launchpad
	var hash, err = getPackageSHA1(rs)

	if err == nil {
		request = apihelper.URL(path.Join("push", d.Project.ID, d.Container.ID))
		apihelper.Auth(request)
		request.Header("Package-Size", fmt.Sprintf("%d", d.PackageSize))
		request.Header("Package-SHA1", hash)
	}

	return request, err
}

func (d *Deploy) deployFeedback(err, errMultipart error) error {
	if err != nil || errMultipart != nil {
		d.progress.setFailure()
	}

	reportDeployMultipleError(err, errMultipart)

	switch {
	case errMultipart != nil:
		return errMultipart
	case err != nil:
		return err
	default:
		d.progress.setComplete(d.PackageSize)
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

	switch hooks {
	case nil:
		return nil
	default:
		return d.runHook(df, wdir, hooks.BeforeDeploy)
	}
}

func (d *Deploy) runAfterHook(df *Flags, wdir string) error {
	var hooks = d.Container.Hooks

	switch hooks {
	case nil:
		return nil
	default:
		return d.runHook(df, wdir, hooks.AfterDeploy)
	}
}

func (d *Deploy) setupPackage(src string) (
	*launchpad.Launchpad, *os.File, error) {
	var file, size, errFD = d.getPackageFD(src)
	d.PackageSize = size

	if errFD != nil {
		return nil, nil, errFD
	}

	d.progress.setUploading()
	var request, err = d.createDeployRequest(file)
	return request, file, err
}

type deployProgress struct {
	bar *progress.Bar
}

func (dp *deployProgress) setComplete(size uint64) {
	dp.bar.Append = fmt.Sprintf(
		"%s (Complete)",
		humanize.Bytes(size))
	dp.bar.Set(progress.Total)
}

func (dp *deployProgress) setFailure() {
	dp.bar.Append = "(Failure)"
	dp.bar.Fail()
}

func (dp *deployProgress) setUploading() {
	dp.bar.Reset("Uploading", "")
}

func (dp *deployProgress) setPacking() {
	dp.bar.Reset("Packing", "")
}
