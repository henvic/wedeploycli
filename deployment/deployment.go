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
	"regexp"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/hashicorp/errwrap"
	"github.com/henvic/uilive"
	"github.com/pkg/browser"
	"github.com/wedeploy/cli/activities"
	"github.com/wedeploy/cli/apihelper"
	"github.com/wedeploy/cli/color"
	"github.com/wedeploy/cli/services"
	"github.com/wedeploy/cli/defaults"
	"github.com/wedeploy/cli/errorhandling"
	"github.com/wedeploy/cli/prompt"
	"github.com/wedeploy/cli/timehelper"
	"github.com/wedeploy/cli/userhome"
	"github.com/wedeploy/cli/verbose"
	"github.com/wedeploy/cli/waitlivemsg"
	"golang.org/x/time/rate"
)

var (
	outStream io.Writer = os.Stdout
	errStream io.Writer = os.Stderr
)

// Deploy project
type Deploy struct {
	Context              context.Context
	AuthorEmail          string
	ProjectID            string
	ServiceID            string
	ChangedServiceID     bool
	Path                 string
	Remote               string
	InfrastructureDomain string
	ServiceDomain        string
	Token                string
	GitRemoteAddress string
	Services         []string
	Quiet            bool
	groupUID         string
	pushStartTime    time.Time
	pushEndTime      time.Time
	sActivities      servicesMap
	wlm              waitlivemsg.WaitLiveMsg
	stepMessage      *waitlivemsg.Message
	uploadMessage    *waitlivemsg.Message
}

func (d *Deploy) getGitPath() string {
	return filepath.Join(userhome.GetHomeDir(), ".wedeploy", "tmp", "repos", d.Path)
}

func (d *Deploy) getGitRemote() string {
	var remote = d.Remote

	// always add a "wedeploy-" prefix to all deployment remote endpoints, but "wedeploy"
	if d.Remote != "wedeploy" {
		remote = "wedeploy" + "-" + d.Remote
	}

	return remote
}

// Cleanup directory
func (d *Deploy) Cleanup() error {
	return os.RemoveAll(d.getGitPath())
}

// CreateGitDirectory creates the git directory for the deployment
func (d *Deploy) CreateGitDirectory() error {
	return os.MkdirAll(d.getGitPath(), 0700)
}

// InitializeRepository as a git repo
func (d *Deploy) InitializeRepository() error {
	var params = []string{"init"}
	verbose.Debug(fmt.Sprintf("Running git %v", strings.Join(params, " ")))
	var cmd = exec.CommandContext(d.Context, "git", params...)
	cmd.Env = append(cmd.Env, "GIT_DIR="+d.getGitPath(), "GIT_WORK_TREE="+d.Path)
	cmd.Dir = d.Path
	cmd.Stderr = errStream

	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
}

// GetCurrentBranch gets the current branch
func (d *Deploy) GetCurrentBranch() (branch string, err error) {
	var params = []string{"symbolic-ref", "HEAD"}
	verbose.Debug(fmt.Sprintf("Running git %v", strings.Join(params, " ")))
	var cmd = exec.CommandContext(d.Context, "git", params...)
	cmd.Env = append(cmd.Env, "GIT_DIR="+d.getGitPath(), "GIT_WORK_TREE="+d.Path)
	var buf bytes.Buffer
	cmd.Dir = d.Path
	cmd.Stderr = errStream
	cmd.Stdout = &buf

	err = cmd.Run()

	if err != nil {
		return "", errwrap.Wrapf("Can not get current branch: {{err}}", err)
	}

	branch = strings.TrimPrefix(strings.TrimSpace(buf.String()), "refs/heads/")
	return branch, nil
}

func (d *Deploy) stageAllFiles() (err error) {
	var params = []string{"add", "."}
	verbose.Debug(fmt.Sprintf("Running git %v", strings.Join(params, " ")))
	var cmd = exec.CommandContext(d.Context, "git", params...)
	cmd.Env = append(cmd.Env, "GIT_DIR="+d.getGitPath(), "GIT_WORK_TREE="+d.Path)
	cmd.Dir = d.Path
	cmd.Stderr = errStream
	cmd.Stdout = outStream

	err = cmd.Run()

	if err != nil || !d.ChangedServiceID {
		return err
	}

	if err = d.stageChangedServiceFile(); err != nil {
		return errwrap.Wrapf("can't stage custom wedeploy.json to replace service ID: {{err}}", err)
	}

	return err
}

func (d *Deploy) stageChangedServiceFile() error {
	var cp, err = services.Read(d.Path)

	if err != nil {
		return err
	}

	cp.ID = d.ServiceID

	var tmpDirPath = d.getGitPath()

	bin, err := json.MarshalIndent(cp, "", "    ")
	if err != nil {
		return err
	}

	if err = ioutil.WriteFile(filepath.Join(tmpDirPath, "wedeploy.json"), bin, 0644); err != nil {
		return err
	}

	defer func() {
		if er := os.Remove(filepath.Join(tmpDirPath, "wedeploy.json")); er != nil {
			verbose.Debug(er)
		}
	}()

	var params = []string{"add", "wedeploy.json"}
	verbose.Debug(fmt.Sprintf("Running git %v", strings.Join(params, " ")))
	var cmd = exec.CommandContext(d.Context, "git", params...)
	cmd.Env = append(cmd.Env, "GIT_DIR="+d.getGitPath(), "GIT_WORK_TREE="+tmpDirPath)
	cmd.Dir = d.Path
	cmd.Stderr = errStream
	cmd.Stdout = outStream

	return cmd.Run()
}

func (d *Deploy) unstageProjectJSON() (err error) {
	var params = []string{"reset", "--", "project.json"}
	verbose.Debug(fmt.Sprintf("Running git %v", strings.Join(params, " ")))
	var cmd = exec.CommandContext(d.Context, "git", params...)
	cmd.Env = append(cmd.Env, "GIT_DIR="+d.getGitPath(), "GIT_WORK_TREE="+d.Path)
	cmd.Dir = d.Path
	cmd.Stderr = errStream

	return cmd.Run()
}

func (d *Deploy) getLastCommit() (commit string, err error) {
	var params = []string{"rev-parse", "HEAD"}
	verbose.Debug(fmt.Sprintf("Running git %v", strings.Join(params, " ")))
	var cmd = exec.CommandContext(d.Context, "git", params...)
	cmd.Env = append(cmd.Env, "GIT_DIR="+d.getGitPath(), "GIT_WORK_TREE="+d.Path)
	var buf bytes.Buffer
	cmd.Dir = d.Path
	cmd.Stderr = errStream
	cmd.Stdout = &buf

	err = cmd.Run()

	if err != nil {
		return "", errwrap.Wrapf("Can not get last commit: {{err}}", err)
	}

	commit = strings.TrimSpace(buf.String())
	return commit, nil
}

// Commit adds all files and commits
func (d *Deploy) Commit() (commit string, err error) {
	if err = d.stageAllFiles(); err != nil {
		return "", errwrap.Wrapf("Trying to stage all files: {{err}}", err)
	}

	if err = d.unstageProjectJSON(); err != nil {
		return "", errwrap.Wrapf("Trying to unstage project.json: {{err}}", err)
	}

	var msg = fmt.Sprintf("Deployment at %v", time.Now().Format(time.RubyDate))

	var params = []string{
		"commit",
		"--allow-empty",
		"--message",
		msg,
	}

	verbose.Debug(fmt.Sprintf("Running git %v", strings.Join(params, " ")))
	var cmd = exec.CommandContext(d.Context, "git", params...)
	cmd.Env = append(cmd.Env, "GIT_DIR="+d.getGitPath(), "GIT_WORK_TREE="+d.Path)
	cmd.Env = append(cmd.Env, fmt.Sprintf("GIT_AUTHOR_EMAIL=%v", d.AuthorEmail))
	cmd.Dir = d.Path

	if verbose.Enabled {
		cmd.Stderr = errStream
	}

	err = cmd.Run()

	if err != nil {
		return "", errwrap.Wrapf("Can not commit: {{err}}", err)
	}

	commit, err = d.getLastCommit()

	if err != nil {
		return "", err
	}

	verbose.Debug("commit", commit)
	return commit, nil
}

func (d *Deploy) verboseOnPush() {
	if !verbose.Enabled {
		return
	}

	verbose.Debug(color.Format(color.FgBlue, "Push Authorization") +
		color.Format(color.FgRed, ": ") +
		verbose.SafeEscape(d.Token))
}

func copyErrStreamAndVerbose(cmd *exec.Cmd) *bytes.Buffer {
	var bufErr bytes.Buffer
	if verbose.Enabled {
		cmd.Stderr = io.MultiWriter(&bufErr, os.Stderr)
	} else {
		cmd.Stderr = &bufErr
	}

	return &bufErr
}

// Push deployment to the WeDeploy remote
func (d *Deploy) Push() (groupUID string, err error) {
	d.pushStartTime = time.Now()
	defer func() {
		d.pushEndTime = time.Now()
	}()

	var params = []string{"push", d.getGitRemote(), "master", "--force"}

	if verbose.Enabled {
		params = append(params, "--verbose")
	}

	verbose.Debug(fmt.Sprintf("Running git %v", strings.Join(params, " ")))
	var cmd = exec.CommandContext(d.Context, "git", params...)
	cmd.Env = append(os.Environ(),
		"GIT_DIR="+d.getGitPath(), "GIT_WORK_TREE="+d.Path,
		"GIT_TERMINAL_PROMPT=0",
		GitCredentialEnvRemoteToken+"="+d.Token,
	)
	cmd.Dir = d.Path

	var bufErr = copyErrStreamAndVerbose(cmd)
	err = cmd.Run()

	if err != nil && strings.Contains(bufErr.String(), "could not read Username") {
		return "", errors.New("Invalid credentials")
	}

	if err != nil {
		return "", err
	}

	return tryGetPushGroupUID(*bufErr)
}

var (
	gitRemoteDeployPrefix      = []byte("remote: wedeploy=")
	gitRemoteDeployErrorPrefix = []byte("remote: wedeployError=")
)

func tryGetPushGroupUID(buff bytes.Buffer) (groupUID string, err error) {
	for {
		line, err := buff.ReadBytes('\n')

		if bytes.HasPrefix(line, gitRemoteDeployPrefix) {
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

// AddRemote on project
func (d *Deploy) AddRemote() error {
	var params = []string{"remote", "add", d.getGitRemote(), d.GitRemoteAddress}
	verbose.Debug(fmt.Sprintf("Running git %v", strings.Join(params, " ")))
	var cmd = exec.CommandContext(d.Context, "git", params...)
	cmd.Env = append(cmd.Env, "GIT_DIR="+d.getGitPath(), "GIT_WORK_TREE="+d.Path)
	cmd.Dir = d.Path
	cmd.Stderr = errStream
	cmd.Stdout = outStream
	return cmd.Run()
}

func (d *Deploy) cleanupAfter() {
	if err := d.Cleanup(); err != nil {
		verbose.Debug(
			errwrap.Wrapf("Error trying to clean up directory after deployment: {{err}}", err))
	}
}

func (d *Deploy) listenCleanupOnCancel() {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigs
		_ = d.Cleanup()
	}()
}

func (d *Deploy) printFailureStep(s string) {
	d.stepMessage.SetText(waitlivemsg.FailureSymbol() + " " + s)
}

func (d *Deploy) printDeploymentFailed() {
	d.printFailureStep("Deployment failed")
}

func (d *Deploy) createStatusMessages() {
	d.stepMessage.SetText("Initializing deployment process")
	d.stepMessage.NoSymbol()
	d.wlm.AddMessage(d.stepMessage)
	d.wlm.AddMessage(waitlivemsg.EmptyLine())

	const udpm = "Uploading deployment package..."

	if d.Quiet {
		fmt.Println("\n" + udpm)
	}

	d.uploadMessage = waitlivemsg.NewMessage(udpm)
	d.wlm.AddMessage(d.uploadMessage)
}

func (d *Deploy) createServicesActivitiesMap() {
	d.sActivities = servicesMap{}
	for _, s := range d.Services {
		var m = waitlivemsg.NewMessage(d.makeServiceStatusMessage(s, "waiting for deployment"))

		d.sActivities[s] = &serviceWatch{
			msgWLM: m,
		}
		d.wlm.AddMessage(m)
	}
}

func (d *Deploy) updateDeploymentEndStep(err error) {
	var timeElapsed = color.Format(color.FgBlue, "%v",
		timehelper.RoundDuration(d.wlm.Duration(), time.Second))

	switch err {
	case nil:
		d.stepMessage.SetText(fmt.Sprintf("Deployment successful in " + timeElapsed))
	default:
		d.stepMessage.SetText(color.Format(color.FgRed, "Deployment failure in ") + timeElapsed)
	}
}

func (d *Deploy) prepareQuiet() {
	p := &bytes.Buffer{}

	p.WriteString(d.getDeployingMessage())
	p.WriteString("\n")

	if len(d.Services) > 0 {
		p.WriteString(fmt.Sprintf("\nList of services:\n"))
	}

	for _, s := range d.Services {
		p.WriteString(d.coloredServiceAddress(s))
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

func (d *Deploy) notifyDeploymentOnQuiet(err error) {
	if err != nil {
		return
	}

	fmt.Printf("Deployment %v is in progress on remote %v\n",
		color.Format(color.FgBlue, d.GetGroupUID()),
		color.Format(color.FgBlue, d.InfrastructureDomain))
}

// Do deployment
func (d *Deploy) Do() error {
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

	d.checkActivitiesLoop()

	var fb, fd []string
	var askLogs = false
	if err == nil {
		fb, fd, err = d.checkDeployment()
		askLogs = (err != nil)
	}

	d.updateDeploymentEndStep(err)

	d.wlm.Stop()

	if askLogs {
		errorhandling.SetAfterError(func() {
			d.maybeOpenLogs(fb, fd)
		})
	}

	return err
}

func (d *Deploy) notifyFailedUpload() {
	d.wlm.RemoveMessage(d.uploadMessage)
	for serviceID, s := range d.sActivities {
		s.msgWLM.SetText(d.makeServiceStatusMessage(serviceID, "Upload failed"))
		s.msgWLM.SetSymbolEnd(waitlivemsg.FailureSymbol())
	}
}

func (d *Deploy) getDeployingMessage() string {
	return fmt.Sprintf("Deploying services on project %v in %v...",
		color.Format(color.FgBlue, d.ProjectID),
		color.Format(color.FgBlue, d.InfrastructureDomain),
	)
}

func (d *Deploy) do() (err error) {
	d.createStatusMessages()
	d.createServicesActivitiesMap()

	if err = d.Cleanup(); err != nil {
		return errwrap.Wrapf("Can not clean up directory for deployment: {{err}}", err)
	}

	d.listenCleanupOnCancel()
	defer signal.Reset(syscall.SIGINT, syscall.SIGTERM)

	if err = d.CreateGitDirectory(); err != nil {
		return errwrap.Wrapf("Can not create temporary directory for deployment: {{err}}", err)
	}

	defer d.cleanupAfter()

	if err = d.preparePackage(); err != nil {
		return err
	}

	d.stepMessage.SetText(d.getDeployingMessage())

	if err = d.uploadPackage(); err != nil {
		return err
	}

	signal.Reset(syscall.SIGINT, syscall.SIGTERM)
	return nil
}

func (d *Deploy) preparePackage() (err error) {
	d.stepMessage.SetText(
		fmt.Sprintf("Preparing deployment for project %v in %v...",
			color.Format(color.FgBlue, d.ProjectID),
			color.Format(d.Remote)),
	)

	if err = d.InitializeRepository(); err != nil {
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

// GitCredentialEnvRemoteToken is the environment variable used for git credential-helper
const GitCredentialEnvRemoteToken = "WEDEPLOY_REMOTE_TOKEN"

func (d *Deploy) addCredentialHelper() (err error) {
	var params = []string{"config", "--add", "credential.helper", os.Args[0] + " git-credential-helper"}
	verbose.Debug(fmt.Sprintf("Running git %v", strings.Join(params, " ")))
	var cmd = exec.CommandContext(d.Context, "git", params...)
	cmd.Env = append(cmd.Env,
		"GIT_DIR="+d.getGitPath(),
		"GIT_WORK_TREE="+d.Path,
	)
	cmd.Dir = d.Path
	cmd.Stderr = errStream
	cmd.Stdout = outStream
	return cmd.Run()
}

func (d *Deploy) uploadPackage() (err error) {
	defer d.uploadMessage.End()
	if d.groupUID, err = d.Push(); err != nil {
		d.uploadMessage.SetText("Upload failed")
		d.uploadMessage.SetSymbolEnd(waitlivemsg.FailureSymbol())
		if _, ok := err.(*exec.ExitError); ok {
			return errwrap.Wrapf("deployment push failed", err)
		}

		return errwrap.Wrapf("deployment failed: {{err}}", err)
	}

	d.uploadMessage.SetText(
		fmt.Sprintf("Upload completed in %v",
			color.Format(color.FgBlue, timehelper.RoundDuration(d.UploadDuration(), time.Second))))
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

type servicesMap map[string]*serviceWatch

func (s servicesMap) GetServicesByState(state string) (keys []string) {
	for k, c := range s {
		if c.state == state {
			keys = append(keys, k)
		}
	}

	return keys
}

type serviceWatch struct {
	state  string
	msgWLM *waitlivemsg.Message
}

func (s servicesMap) isFinalState(key string) bool {
	if s == nil || s[key] == nil {
		return false
	}

	switch s[key].state {
	case activities.BuildFailed,
		activities.DeployFailed,
		activities.DeploySucceeded:
		return true
	}

	return false
}

func (d *Deploy) updateActivityState(a activities.Activity) {
	var serviceID, ok = a.Metadata["serviceId"]

	// stop processing if service is not any of the watched deployment cycle types,
	// or if service ID is somehow not available
	if !ok || !isActitityTypeDeploymentRelated(a.Type) {
		return
	}

	if _, exists := d.sActivities[serviceID]; !exists {
		// skip activity
		// we want to avoid problems (read: nil pointer panics)
		// if the server sends back a response with an ID we don't have already locally
		return
	}

	d.markActivityState(serviceID, a.Type)
	var wlm = d.sActivities[serviceID].msgWLM
	var pre string

	var prefixes = map[string]string{
		activities.BuildFailed:     "build failed",
		activities.BuildPending:    "build pending",
		activities.BuildStarted:    "building",
		activities.BuildSucceeded:  "build successful",
		activities.DeployFailed:    "deploy failed",
		activities.DeployPending:   "deploy pending",
		activities.DeploySucceeded: "deployed",
		activities.DeployStarted:   "deploying",
	}

	if pre, ok = prefixes[a.Type]; !ok {
		pre = a.Type
	}

	wlm.SetText(d.makeServiceStatusMessage(serviceID, pre))

	switch a.Type {
	case
		activities.BuildFailed,
		activities.DeployFailed:
		wlm.SetSymbolEnd(waitlivemsg.FailureSymbol())
		wlm.End()
	case
		activities.DeploySucceeded:
		wlm.End()
	}
}

func isActitityTypeDeploymentRelated(activityType string) bool {
	switch activityType {
	case
		activities.BuildPending,
		activities.BuildStarted,
		activities.BuildFailed,
		activities.BuildSucceeded,
		activities.DeployPending,
		activities.DeployStarted,
		activities.DeployFailed,
		activities.DeploySucceeded:
		return true
	}

	return false
}

func (d *Deploy) markActivityState(serviceID, activityType string) {
	switch activityType {
	case
		activities.BuildSucceeded,
		activities.BuildFailed,
		activities.DeployFailed,
		activities.DeploySucceeded:
		d.sActivities[serviceID].state = activityType
	}
}

func (d *Deploy) checkActivities() (end bool, err error) {
	var as activities.Activities
	var ctx, cancel = context.WithTimeout(d.Context, 5*time.Second)
	defer cancel()
	as, err = activities.List(ctx, d.ProjectID, activities.Filter{
		GroupUID: d.groupUID,
	})
	cancel()
	as = as.Reverse()

	if err != nil {
		return false, err
	}

	for _, a := range as {
		d.updateActivityState(a)
	}

	for id := range d.sActivities {
		if !d.sActivities.isFinalState(id) {
			return false, nil
		}
	}

	return true, nil
}

func updateMessageErrorStringCounter(input string) (output string) {
	var r = regexp.MustCompile(`\(retrying to get status #([0-9]+)\)`)

	if input == "" {
		return "(retrying to get status #1)"
	}

	if r.FindString(input) == "" {
		return input + " (retrying to get status #1)"
	}

	return string(r.ReplaceAllStringFunc(input, func(n string) string {
		const prefix = "(retrying to get status #"
		const suffix = ")"

		if len(n) <= len(prefix)+len(suffix) {
			return n
		}

		var num, _ = strconv.Atoi(n[len(prefix) : len(n)-1])
		num++
		return fmt.Sprintf("(retrying to get status #%v)", num)
	}))
}

func clearMessageErrorStringCounter(input string) (output string) {
	var r = regexp.MustCompile(`\s?\(retrying to get status #([0-9]+)\)`)
	return r.ReplaceAllString(input, "")
}

func (d *Deploy) checkActivitiesLoop() {
	rate := rate.NewLimiter(rate.Every(time.Second), 1)

	for {
		if er := rate.Wait(d.Context); er != nil {
			verbose.Debug(er)
		}

		var end, err = d.checkActivities()
		var stepText = d.stepMessage.GetText()

		if err != nil {
			d.stepMessage.SetText(updateMessageErrorStringCounter(stepText))
			verbose.Debug(err)
			continue
		}

		if strings.Contains(stepText, "retrying to get status #") {
			// it is not cleaning always
			d.stepMessage.SetText(clearMessageErrorStringCounter(stepText))
		}

		if end {
			return
		}
	}
}

func (d *Deploy) printServiceAddress(service string) string {
	var address = d.ProjectID + "." + d.ServiceDomain

	if service != "" {
		address = service + "-" + address
	}

	return address
}

func (d *Deploy) coloredServiceAddress(serviceID string) string {
	return color.Format(
		color.FgBlue,
		d.printServiceAddress(serviceID))
}

func (d *Deploy) makeServiceStatusMessage(serviceID, pre string) string {
	var buff bytes.Buffer

	if pre != "" {
		buff.WriteString(pre)
		buff.WriteString(" ")
	}

	buff.WriteString(d.coloredServiceAddress(serviceID))

	return buff.String()
}

func (d *Deploy) checkDeployment() (failedBuilds []string, failedDeploys []string, err error) {
	var feedback string
	failedBuilds = d.sActivities.GetServicesByState(activities.BuildFailed)
	failedDeploys = d.sActivities.GetServicesByState(activities.DeployFailed)

	if len(failedBuilds) == 0 && len(failedDeploys) == 0 {
		return failedBuilds, failedDeploys, nil
	}

	switch len(d.sActivities) {
	case len(failedBuilds) + len(failedDeploys):
		feedback = "deployment failed while"
	default:
		feedback = "deployment failed partially while"
	}

	if len(failedBuilds) != 0 {
		feedback += " building service"

		if len(failedBuilds) != 1 {
			feedback += "s"
		}

		feedback += " " + color.Format(color.FgBlue, strings.Join(failedBuilds, ", "))
	}

	if len(failedBuilds) != 0 && len(failedDeploys) != 0 {
		feedback += ", and"
	}

	if len(failedDeploys) != 0 {
		feedback += " deploying service"

		if len(failedDeploys) != 1 {
			feedback += "s"
		}

		feedback += " " + color.Format(color.FgBlue, strings.Join(failedDeploys, ", "))
	}

	return failedBuilds, failedDeploys, errors.New(feedback)
}

func (d *Deploy) maybeOpenLogs(failedBuilds, failedDeploys []string) {
shouldOpenPrompt:
	fmt.Println("Do you want to check the logs (yes/no)? [no]: ")
	var p, err = prompt.Prompt()

	if err != nil {
		verbose.Debug(err)
		return
	}

	switch strings.ToLower(p) {
	case "yes", "y":
		break
	case "no", "n", "":
		return
	default:
		goto shouldOpenPrompt
	}

	var logsURL = fmt.Sprintf("https://%v%v/projects/%v/logs",
		defaults.DashboardAddressPrefix,
		d.InfrastructureDomain,
		d.ProjectID)

	switch {
	case len(failedBuilds) == 1 && len(failedDeploys) == 0:
		logsURL += "?label=buildUid&logServiceId=" + failedBuilds[0]
	case len(failedBuilds) == 0 && len(failedDeploys) == 1:
		logsURL += "?logServiceId=" + failedDeploys[0]
	case len(failedDeploys) == 0:
		logsURL += "?label=buildUid"
	}

	if err := browser.OpenURL(logsURL); err != nil {
		fmt.Println("Open URL: (can't open automatically)", logsURL)
		verbose.Debug(err)
	}
}
