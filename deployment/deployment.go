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
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/hashicorp/errwrap"
	"github.com/henvic/uilive"
	"github.com/wedeploy/cli/apihelper"
	"github.com/wedeploy/cli/color"
	"github.com/wedeploy/cli/config"
	"github.com/wedeploy/cli/envs"
	"github.com/wedeploy/cli/errorhandler"
	"github.com/wedeploy/cli/fancy"
	"github.com/wedeploy/cli/services"
	"github.com/wedeploy/cli/timehelper"
	"github.com/wedeploy/cli/verbose"
	"github.com/wedeploy/cli/waitlivemsg"
)

var (
	outStream io.Writer = os.Stdout
	errStream io.Writer = os.Stderr
)

// Deploy project
type Deploy struct {
	ConfigContext config.Context
	ProjectID     string
	ServiceID     string
	Path          string
	Services      services.ServiceInfoList

	CopyPackage string
	Quiet       bool

	gitVersion    string
	groupUID      string
	pushStartTime time.Time
	pushEndTime   time.Time

	sActivities   servicesMap
	wlm           waitlivemsg.WaitLiveMsg
	stepMessage   *waitlivemsg.Message
	uploadMessage *waitlivemsg.Message

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
	bin, err := getPreparedServicePackage(s.ServiceID, s.Location)
	if err != nil {
		return err
	}

	return d.overwriteServicePackage(bin, filepath.Join(filepath.Base(s.Location),
		"wedeploy.json"))
}

func getPreparedServicePackage(serviceID string, path string) ([]byte, error) {
	// this smells a little bad because wedeploy.json is the responsibility of the services package
	// and I shouldn't be accessing it directly from here
	var spMap = map[string]interface{}{}
	wedeployJSON, err := ioutil.ReadFile(filepath.Join(path, "wedeploy.json"))
	switch {
	case err == nil:
		if err = json.Unmarshal(wedeployJSON, &spMap); err != nil {
			return nil, errwrap.Wrapf("error parsing wedeploy.json on "+path+": {{err}}", err)
		}
	case os.IsNotExist(err):
		spMap = map[string]interface{}{}
		err = nil
	default:
		return nil, err
	}

	spMap["id"] = strings.ToLower(serviceID)
	delete(spMap, "projectId")

	return json.MarshalIndent(&spMap, "", "    ")
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

func (d *Deploy) listenCleanupOnCancel() {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigs
		d.ctxCancel()
	}()
}

func (d *Deploy) prepareQuiet() {
	p := &bytes.Buffer{}

	p.WriteString(d.getDeployingMessage())
	p.WriteString("\n")

	if len(d.Services) > 0 {
		p.WriteString(fmt.Sprintf("\nList of services:\n"))
	}

	for _, s := range d.Services {
		p.WriteString(d.coloredServiceAddress(s.ServiceID))
		p.WriteString("\n")
	}

	fmt.Print(p)
}

func (d *Deploy) prepareNoisy() {
	var us = uilive.New()
	d.wlm.SetStream(us)
	go d.wlm.Wait()
}

func (d *Deploy) prepare() {
	if d.Quiet {
		d.prepareQuiet()
		return
	}

	d.prepareNoisy()
}

// Do deployment
func (d *Deploy) Do(ctx context.Context) error {
	d.ctx, d.ctxCancel = context.WithCancel(ctx)
	d.stepMessage = &waitlivemsg.Message{}
	d.wlm = waitlivemsg.WaitLiveMsg{}
	d.prepare()

	var err = d.do()

	if d.Quiet {
		d.notifyDeploymentOnQuiet(err)
		return err
	}

	if err != nil {
		d.updateDeploymentEndStep(err)
		d.notifyFailedUpload()
		d.wlm.Stop()
		return err
	}

	d.watchDeployment()

	states, err := d.verifyFinalState()

	d.updateDeploymentEndStep(err)

	d.wlm.Stop()

	var askLogs = (len(states.BuildFailed) != 0 || len(states.DeployFailed) != 0)

	if askLogs {
		errorhandler.SetAfterError(func() {
			d.maybeOpenLogs(states)
		})
	}

	return err
}

func (d *Deploy) do() (err error) {
	d.createStatusMessages()
	d.createServicesActivitiesMap()

	d.listenCleanupOnCancel()

	defer func() {
		if ec := d.Cleanup(); ec != nil {
			if err == nil {
				err = ec
				return
			}

			verbose.Debug(err)
		}

		signal.Reset(syscall.SIGINT, syscall.SIGTERM)
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

	d.stepMessage.StopText(d.getDeployingMessage())

	if err = d.uploadPackage(); err != nil {
		return err
	}

	signal.Reset(syscall.SIGINT, syscall.SIGTERM)
	return nil
}

func (d *Deploy) preparePackage() (err error) {
	d.stepMessage.StopText(
		fmt.Sprintf("Preparing deployment for project %v in %v...",
			color.Format(color.FgBlue, d.ProjectID),
			color.Format(d.ConfigContext.Remote())),
	)

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
	defer d.uploadMessage.End()
	if d.groupUID, err = d.Push(); err != nil {
		d.uploadMessage.StopText(fancy.Error("Upload failed"))
		if _, ok := err.(*exec.ExitError); ok {
			return errwrap.Wrapf("deployment push failed", err)
		}

		// don't wrap: expect apihelper.APIFault
		return err
	}

	d.uploadMessage.StopText(fmt.Sprintf("Upload completed in %v",
		timehelper.RoundDuration(d.UploadDuration(), time.Second)))
	return nil
}

// Cleanup directory
func (d *Deploy) Cleanup() error {
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
