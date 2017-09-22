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
	"github.com/henvic/browser"
	"github.com/henvic/uilive"
	"github.com/wedeploy/cli/activities"
	"github.com/wedeploy/cli/apihelper"
	"github.com/wedeploy/cli/color"
	"github.com/wedeploy/cli/config"
	"github.com/wedeploy/cli/defaults"
	"github.com/wedeploy/cli/errorhandling"
	"github.com/wedeploy/cli/fancy"
	"github.com/wedeploy/cli/services"
	"github.com/wedeploy/cli/timehelper"
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
	Context          context.Context
	ConfigContext    *config.ContextType
	ProjectID        string
	ServiceID        string
	LocationRemap    []string
	Path             string
	GitRemoteAddress string
	Services         services.ServiceInfoList
	Quiet            bool
	groupUID         string
	pushStartTime    time.Time
	pushEndTime      time.Time
	sActivities      servicesMap
	wlm              waitlivemsg.WaitLiveMsg
	stepMessage      *waitlivemsg.Message
	uploadMessage    *waitlivemsg.Message
	gitEnvCache      []string
}

func (d *Deploy) renameServiceID(s services.ServiceInfo) error {
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

	var rel, errRel = filepath.Rel(d.Path, s.Location)

	if errRel != nil {
		return err
	}

	// It is necessary to use a map instead of relying on the structure we have
	// to avoid compatibility issues due to lack of a synchronization channel
	// between the CLI team and the other teams in maintaining wedeploy.json structure
	// synced.
	bin, err := replaceServicePackageToInterfaceOnRenaming(s.ServiceID, s.Location)
	if err != nil {
		return err
	}

	return d.gitRenameServiceID(bin, filepath.Join(rel, "wedeploy.json"))
}

func replaceServicePackageToInterfaceOnRenaming(serviceID string, path string) ([]byte, error) {
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

	spMap["id"] = serviceID
	return json.MarshalIndent(&spMap, "", "    ")
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

func getGitErrors(s string) error {
	var parts = strings.Split(s, "\n")
	var list = []string{}
	for _, p := range parts {
		if strings.Contains(p, "error: ") {
			fmt.Println(p)
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

func (d *Deploy) createStatusMessages() {
	d.stepMessage.PlayText("Initializing deployment process")
	d.wlm.AddMessage(d.stepMessage)

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
		var m = &waitlivemsg.Message{}
		m.StopText(d.makeServiceStatusMessage(s.ServiceID, "⠂"))

		d.sActivities[s.ServiceID] = &serviceWatch{
			msgWLM: m,
		}
		d.wlm.AddMessage(m)
	}
}

func (d *Deploy) updateDeploymentEndStep(err error) {
	var timeElapsed = timehelper.RoundDuration(d.wlm.Duration(), time.Second)

	switch err {
	case nil:
		d.stepMessage.StopText(d.getDeployingMessage() + "\n" +
			fancy.Success(fmt.Sprintf("Deployment successful in %s", timeElapsed)))
	default:
		d.stepMessage.StopText(d.getDeployingMessage() + "\n" +
			fancy.Error(fmt.Sprintf("Deployment failed in %s", timeElapsed)))
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

func (d *Deploy) notifyDeploymentOnQuiet(err error) {
	if err != nil {
		return
	}

	fmt.Printf("Deployment %v is in progress on remote %v\n",
		color.Format(color.FgBlue, d.GetGroupUID()),
		color.Format(color.FgBlue, d.ConfigContext.InfrastructureDomain))
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
		s.msgWLM.PlayText(fancy.Error(d.makeServiceStatusMessage(serviceID, "Upload failed")))
	}
}

func (d *Deploy) getDeployingMessage() string {
	return fmt.Sprintf("Deploying services on project %v in %v...",
		color.Format(color.FgBlue, d.ProjectID),
		color.Format(color.FgBlue, d.ConfigContext.InfrastructureDomain),
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
			color.Format(d.ConfigContext.Remote)),
	)

	if hasGit := existsDependency("git"); !hasGit {
		return errors.New("git was not found on your system: please visit https://git-scm.com/")
	}

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

	d.uploadMessage.StopText(
		fancy.Success(fmt.Sprintf("Upload completed in %v",
			timehelper.RoundDuration(d.UploadDuration(), time.Second))))
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
	var m = d.sActivities[serviceID].msgWLM
	var pre string

	var prefixes = map[string]string{
		activities.BuildFailed:     "Build failed",
		activities.BuildStarted:    "Building",
		activities.BuildPushed:     "Build push",
		activities.BuildSucceeded:  "Build successful",
		activities.DeployFailed:    "Deploy failed",
		activities.DeployCreated:   "Deploy created",
		activities.DeployPending:   "Deploy pending",
		activities.DeploySucceeded: "Deployed",
		activities.DeployStarted:   "Deploying",
	}

	if pre, ok = prefixes[a.Type]; !ok {
		pre = a.Type
	}

	switch a.Type {
	case activities.BuildStarted,
		activities.BuildPushed,
		activities.BuildSucceeded,
		activities.DeployCreated,
		activities.DeployPending,
		activities.DeployStarted:
		m.PlayText(d.makeServiceStatusMessage(serviceID, pre))
	case
		activities.BuildFailed,
		activities.DeployFailed:
		m.StopText(fancy.Error(d.makeServiceStatusMessage(serviceID, pre)))
	case
		activities.DeploySucceeded:
		m.StopText(fancy.Success(d.makeServiceStatusMessage(serviceID, pre)))
	default:
		m.StopText(d.makeServiceStatusMessage(serviceID, pre))
	}
}

func isActitityTypeDeploymentRelated(activityType string) bool {
	switch activityType {
	case
		activities.BuildStarted,
		activities.BuildFailed,
		activities.BuildPushed,
		activities.BuildSucceeded,
		activities.DeployCreated,
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
			d.stepMessage.StopText(updateMessageErrorStringCounter(stepText))
			verbose.Debug(err)
			continue
		}

		if strings.Contains(stepText, "retrying to get status #") {
			d.stepMessage.StopText(clearMessageErrorStringCounter(stepText))
		}

		if end {
			return
		}
	}
}

func (d *Deploy) printServiceAddress(service string) string {
	var address = d.ProjectID + "." + d.ConfigContext.ServiceDomain

	if service != "" {
		address = service + "-" + address
	}

	return address
}

func (d *Deploy) coloredServiceAddress(serviceID string) string {
	return color.Format(
		color.Bold,
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

		feedback += " " + color.Format(color.Bold, strings.Join(failedBuilds, ", "))
	}

	if len(failedBuilds) != 0 && len(failedDeploys) != 0 {
		feedback += ", and"
	}

	if len(failedDeploys) != 0 {
		feedback += " deploying service"

		if len(failedDeploys) != 1 {
			feedback += "s"
		}

		feedback += " " + color.Format(color.Bold, strings.Join(failedDeploys, ", "))
	}

	return failedBuilds, failedDeploys, errors.New(feedback)
}

func (d *Deploy) maybeOpenLogs(failedBuilds, failedDeploys []string) {
	var options = fancy.Options{}
	options.Add("Y", "Open Browser")
	options.Add("N", "Cancel")
	choice, err := options.Ask("Do you want to check the logs?")

	if err != nil {
		fmt.Fprintf(os.Stderr, "%v", err)
	}

	if choice == "N" {
		return
	}

	var logsURL = fmt.Sprintf("https://%v%v/projects/%v/logs",
		defaults.DashboardAddressPrefix,
		d.ConfigContext.InfrastructureDomain,
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
