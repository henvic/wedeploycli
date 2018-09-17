package deployment

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/hashicorp/errwrap"
	"github.com/wedeploy/cli/apihelper"
	"github.com/wedeploy/cli/config"
	"github.com/wedeploy/cli/deployment/internal/feedback"
	"github.com/wedeploy/cli/envs"
	"github.com/wedeploy/cli/services"
	"github.com/wedeploy/cli/verbose"
)

var (
	outStream io.Writer = os.Stdout
	errStream io.Writer = os.Stderr
)

// Deploy project
type Deploy struct {
	ConfigContext config.Context

	ProjectID string
	ServiceID string

	Image string

	Path     string
	Services services.ServiceInfoList

	CopyPackage string

	OnlyBuild    bool
	SkipProgress bool
	Quiet        bool

	gitVersion    string
	groupUID      string
	pushStartTime time.Time
	pushEndTime   time.Time

	watch *feedback.Watch

	ctx       context.Context
	ctxCancel context.CancelFunc

	tmpWorkDir     string
	tmpWorkDirLock sync.RWMutex

	ignoreList map[string]bool

	gitEnvCache []string
}

func (d *Deploy) prepareAndModifyServicePackage(s services.ServiceInfo) error {
	// ignore service package contents because it is strict (see note below)
	var _, err = services.Read(s.Location)

	switch err {
	case nil:
	case services.ErrServiceNotFound:
		verbose.Debug("Service not found. Creating service package on git repo only.")
		err = nil
	default:
		return err
	}

	// It is necessary to use a map instead of relying on the structure we have
	// to avoid compatibility issues due to lack of a synchronization channel
	// between the CLI team and the other teams in maintaining wedeploy.json structure
	// synced.

	c := changes{
		ServiceID: s.ServiceID,
		Image:     d.Image,
	}

	bin, err := getPreparedServicePackage(c, s.Location)
	if err != nil {
		return err
	}

	// see Windows os.PathSeparator related issue #449
	return d.overwriteServicePackage(bin, fmt.Sprintf("%s/wedeploy.json", filepath.Base(s.Location)))
}

type changes struct {
	ServiceID string
	Image     string
}

func getPreparedServicePackage(c changes, path string) ([]byte, error) {
	// this smells a little bad because wedeploy.json is the responsibility of the services package
	// and I shouldn't be accessing it directly from here
	var sp = map[string]interface{}{}
	wedeployJSON, err := ioutil.ReadFile(filepath.Join(path, "wedeploy.json"))
	switch {
	case err == nil:
		if err = json.Unmarshal(wedeployJSON, &sp); err != nil {
			return nil, errwrap.Wrapf("error parsing wedeploy.json on "+path+": {{err}}", err)
		}
	case os.IsNotExist(err):
		sp = map[string]interface{}{}
		err = nil
	default:
		return nil, err
	}

	sp["id"] = strings.ToLower(c.ServiceID)

	if c.Image != "" {
		sp["image"] = c.Image
	}

	delete(sp, "projectId")

	return json.MarshalIndent(&sp, "", "    ")
}

func copyErrStreamAndVerbose(cmd *exec.Cmd) *bytes.Buffer {
	var bufErr bytes.Buffer
	cmd.Stderr = &bufErr

	switch {
	case verbose.Enabled && verbose.IsUnsafeMode():
		cmd.Stderr = io.MultiWriter(&bufErr, os.Stderr)
	case verbose.Enabled:
		verbose.Debug(fmt.Sprintf(
			"Use %v=true to override security protection (see wedeploy/cli #327)",
			envs.UnsafeVerbose))
	}

	return &bufErr
}

func getGitErrors(s string) error {
	var parts = strings.Split(s, "\n")
	var list = []string{}
	for _, p := range parts {
		if strings.Contains(p, "error: ") {
			list = append(list, p)
		}
	}

	if len(list) == 0 {
		return nil
	}

	return fmt.Errorf("push: %v", strings.Join(list, "\n"))
}

var (
	gitRemoteDeployPrefix      = []byte("remote: wedeploy=")
	gitRemoteDeployErrorPrefix = []byte("remote: wedeployError=")
)

func tryGetPushGroupUID(buff bytes.Buffer) (groupUID string, err error) {
	for {
		line, err := buff.ReadBytes('\n')

		if bytes.HasPrefix(line, gitRemoteDeployPrefix) {
			// \x1b[K is showing up at the end of "remote: wedeploy=" on at least git 1.9
			if filterX1bk := []byte("\x1b[K\n"); bytes.HasSuffix(line, filterX1bk) {
				line = append(line[:len(line)-len(filterX1bk)], byte('\n'))
			}

			return extractGroupUIDFromBuild(bytes.TrimPrefix(line, gitRemoteDeployPrefix))
		}

		if bytes.HasPrefix(line, gitRemoteDeployErrorPrefix) {
			return "", extractErrorFromBuild(bytes.TrimPrefix(line, gitRemoteDeployErrorPrefix))
		}

		if err == io.EOF {
			return "", errors.New("can't find deployment group UID response")
		}
	}
}

func extractErrorFromBuild(e []byte) error {
	var af apihelper.APIFault
	if errJSON := json.Unmarshal(e, &af); errJSON != nil {
		return fmt.Errorf(`can't process error message: "%s"`, bytes.TrimSpace(e))
	}

	return af
}

type buildDeploymentOnGitServer struct {
	GroupUID string `json:"groupUid"`
}

func extractGroupUIDFromBuild(e []byte) (groupUID string, err error) {
	var bds []buildDeploymentOnGitServer

	if errJSON := json.Unmarshal(e, &bds); errJSON != nil {
		return "", errwrap.Wrapf("deployment response is invalid: {{err}}", errJSON)
	}

	if len(bds) == 0 {
		return "", errors.New("found no build during deployment")
	}

	return bds[0].GroupUID, nil
}

// Do deployment
func (d *Deploy) Do(ctx context.Context) error {
	d.ctx, d.ctxCancel = context.WithCancel(ctx)

	d.watch = &feedback.Watch{
		ConfigContext: d.ConfigContext,

		ProjectID: d.ProjectID,

		Services: d.Services,

		OnlyBuild:    d.OnlyBuild,
		SkipProgress: d.SkipProgress,
		Quiet:        d.Quiet,

		IsUpload: true,
	}

	d.watch.Start(d.ctx)

	var err = d.do()

	d.watch.GroupUID = d.GetGroupUID()

	if d.SkipProgress && err != nil {
		return err
	}

	if d.SkipProgress {
		d.watch.PrintSkipProgress()
		return nil
	}

	if err != nil {
		d.watch.StopFailedUpload()
		return err
	}

	return d.watch.Wait()
}

func (d *Deploy) do() (err error) {
	defer func() {
		if ec := d.CleanupPackage(); ec != nil {
			if err == nil {
				err = ec
				return
			}

			_, _ = fmt.Fprintf(os.Stderr, "%+v\n", err)
		}
	}()

	tmpWorkDir, err := ioutil.TempDir("", "wedeploy")

	if err != nil {
		return err
	}

	d.setTmpWorkDir(tmpWorkDir)

	if err = d.createGitDirectory(); err != nil {
		return err
	}

	if err = d.preparePackage(); err != nil {
		return err
	}

	d.watch.NotifyStart()

	return d.uploadPackage()
}

func (d *Deploy) preparePackage() (err error) {
	d.watch.NotifyPacking()

	if hasGit := existsDependency("git"); !hasGit {
		return errors.New("git was not found on your system: please visit https://git-scm.com/")
	}

	if err = d.initializeRepository(); err != nil {
		return err
	}

	if _, err = d.Commit(); err != nil {
		return err
	}

	if err = d.AddRemote(); err != nil {
		return err
	}

	return d.addCredentialHelper()
}

func existsDependency(cmd string) bool {
	_, err := exec.LookPath(cmd)
	return err == nil
}

func getWeExecutable() (string, error) {
	var exec, err = os.Executable()

	if err != nil {
		verbose.Debug(fmt.Sprintf("%v; falling back to os.Args[0]", err))
		return filepath.Abs(os.Args[0])
	}

	return exec, nil
}

func (d *Deploy) uploadPackage() (err error) {
	if d.groupUID, err = d.Push(); err != nil {
		if _, ok := err.(*exec.ExitError); ok {
			return errwrap.Wrapf("upload failed: {{err}}", err)
		}

		// don't wrap: expect apihelper.APIFault
		return err
	}

	d.watch.NotifyUploadComplete(d.UploadDuration())
	return nil
}

// CleanupPackage directory
func (d *Deploy) CleanupPackage() error {
	var tmpWorkDir = d.getTmpWorkDir()

	if tmpWorkDir != "" {
		return os.RemoveAll(d.getTmpWorkDir())
	}

	return nil
}

// GetGroupUID gets the deployment group UID
func (d *Deploy) GetGroupUID() string {
	return d.groupUID
}

// UploadDuration for deployment (only correct after it finishes)
func (d *Deploy) UploadDuration() time.Duration {
	return d.pushEndTime.Sub(d.pushStartTime)
}
