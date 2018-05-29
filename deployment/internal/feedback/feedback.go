package feedback

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/errwrap"
	"github.com/henvic/browser"
	"github.com/henvic/uilive"
	"github.com/wedeploy/cli/activities"
	"github.com/wedeploy/cli/apihelper"
	"github.com/wedeploy/cli/color"
	"github.com/wedeploy/cli/config"
	"github.com/wedeploy/cli/defaults"
	"github.com/wedeploy/cli/errorhandler"
	"github.com/wedeploy/cli/fancy"
	"github.com/wedeploy/cli/figures"
	"github.com/wedeploy/cli/projects"
	"github.com/wedeploy/cli/services"
	"github.com/wedeploy/cli/timehelper"
	"github.com/wedeploy/cli/verbose"
	"github.com/wedeploy/cli/verbosereq"
	"github.com/wedeploy/cli/waitlivemsg"
	"golang.org/x/time/rate"
)

// Watch deployment feedback.
type Watch struct {
	ConfigContext config.Context

	ProjectID string
	GroupUID  string

	Services services.ServiceInfoList

	Quiet bool

	ctx context.Context

	IsUpload        bool
	uploadCompleted time.Duration

	wlm         waitlivemsg.WaitLiveMsg
	header      *waitlivemsg.Message
	stepMessage *waitlivemsg.Message

	sActivities servicesMap

	states finalStates
}

// Start watching deployment feedback.
func (w *Watch) Start(ctx context.Context) {
	w.ctx = ctx

	w.stepMessage = &waitlivemsg.Message{}
	w.header = &waitlivemsg.Message{}
	w.wlm = waitlivemsg.WaitLiveMsg{}

	if w.Quiet && w.IsUpload {
		w.prepareQuiet()
		fmt.Println("Uploading.")
	}

	if !w.Quiet {
		w.prepareNoisy()
		w.header.PlayText(w.getDeployingMessage())
		w.stepMessage.StopText("")
		w.wlm.AddMessage(w.header)
		w.wlm.AddMessage(w.stepMessage)
	}

	w.createServicesActivitiesMap()
}

func (w *Watch) notifyDeploymentSucceeded() {
	var timeElapsed = timehelper.RoundDuration(w.wlm.Duration(), time.Second)

	w.header.StopText(figures.Tick + " " + w.getDeployingMessage())

	msg := fmt.Sprintf("%s Deployment succeeded in %s", figures.Tick, timeElapsed)

	if w.IsUpload && w.uploadCompleted != 0 {
		msg = fmt.Sprintf("%s. Upload completed in %v.",
			msg,
			timehelper.RoundDuration(w.uploadCompleted, time.Second))
	}

	w.stepMessage.StopText(msg)
}

func (w *Watch) notifyDeploymentFailed() {
	var timeElapsed = timehelper.RoundDuration(w.wlm.Duration(), time.Second)

	w.header.StopText(figures.Cross + " " + w.getDeployingMessage())

	msg := fmt.Sprintf("%s Deployment failed in %s", figures.Cross, timeElapsed)

	if w.IsUpload && w.uploadCompleted != 0 {
		msg = fmt.Sprintf("%s. Upload completed in %v.",
			msg,
			timehelper.RoundDuration(w.uploadCompleted, time.Second))
	}

	w.stepMessage.StopText(msg)
}

// PrintQuietDeployment prints a message telling that the deployment
// is happening on background when the deployment is not interactive.
func (w *Watch) PrintQuietDeployment() {
	fmt.Printf("Deployment %v is in progress on remote %v\n",
		color.Format(color.FgMagenta, color.Bold, w.GroupUID),
		color.Format(color.FgMagenta, color.Bold, w.ConfigContext.InfrastructureDomain()))
}

// StopFailedUpload stops the deployment messages due to a failed upload error.
func (w *Watch) StopFailedUpload() {
	if !w.Quiet {
		return
	}

	w.notifyDeploymentFailed()

	for serviceID, s := range w.sActivities {
		s.msgWLM.PlayText(fancy.Error(w.makeServiceStatusMessage(serviceID, "Upload failed")))
	}

	w.wlm.Stop()
}

// NotifyPacking notifies that a package is being prepared for deployment.
func (w *Watch) NotifyPacking() {
	w.header.PlayText(
		fmt.Sprintf("Preparing deployment for project %v in %v...",
			color.Format(color.FgMagenta, color.Bold, w.ProjectID),
			color.Format(w.ConfigContext.Remote())),
	)
}

// NotifyDeploying notifies that the deployment started.
func (w *Watch) NotifyDeploying() {
	w.header.PlayText(w.getDeployingMessage())
}

// NotifyUploadComplete notifies that the upload has been completed.
func (w *Watch) NotifyUploadComplete(t time.Duration) {
	if !w.IsUpload {
		return
	}

	w.uploadCompleted = t

	uploadCompletedFeedback := fmt.Sprintf("Upload completed in %v.",
		timehelper.RoundDuration(t, time.Second))

	if w.Quiet {
		fmt.Fprintln(os.Stderr, uploadCompletedFeedback)
	} else {
		w.stepMessage.PlayText(uploadCompletedFeedback)
	}
}

type finalStates struct {
	BuildFailed    []string
	DeployFailed   []string
	DeployCanceled []string
	DeployTimeout  []string
	DeployRollback []string
}

func (w *Watch) prepareQuiet() {
	p := &bytes.Buffer{}

	p.WriteString(w.getDeployingMessage())
	p.WriteString("\n")

	if len(w.Services) > 0 {
		p.WriteString(fmt.Sprintf("\nList of services:\n"))
	}

	for _, s := range w.Services {
		p.WriteString(w.coloredServiceAddress(s.ServiceID))
		p.WriteString("\n")
	}

	fmt.Print(p)
}

func (w *Watch) prepareNoisy() {
	var us = uilive.New()
	w.wlm.SetStream(us)
	go w.wlm.Wait()
}

func (w *Watch) createServicesActivitiesMap() {
	w.sActivities = servicesMap{}
	for _, s := range w.Services {
		var m = &waitlivemsg.Message{}
		msg := "Waiting"

		if w.IsUpload {
			msg = "Uploading"
		}

		m.PlayText(w.makeServiceStatusMessage(s.ServiceID, msg))

		w.sActivities[s.ServiceID] = &serviceWatch{
			msgWLM: m,
		}
		w.wlm.AddMessage(m)
	}
}

func (w *Watch) reorderDeployments() {
	projectsClient := projects.New(w.ConfigContext)
	order, _ := projectsClient.GetDeploymentOrder(w.ctx, w.ProjectID, w.GroupUID)

	for _, do := range order {
		if a, ok := w.sActivities[do]; ok {
			w.wlm.RemoveMessage(a.msgWLM)
		}
	}

	for _, do := range order {
		if a, ok := w.sActivities[do]; ok {
			w.wlm.AddMessage(a.msgWLM)
		}
	}
}

func (w *Watch) getDeployingMessage() string {
	return fmt.Sprintf("Deploying services to project %v on %v",
		color.Format(color.FgMagenta, color.Bold, w.ProjectID),
		color.Format(color.FgMagenta, color.Bold, w.ConfigContext.InfrastructureDomain()),
	)
}

func (w *Watch) printServiceAddress(service string) string {
	var address = w.ProjectID + "." + w.ConfigContext.ServiceDomain()

	if service != "" {
		address = service + "-" + address
	}

	return address
}

func (w *Watch) coloredServiceAddress(serviceID string) string {
	return color.Format(
		color.Bold,
		w.printServiceAddress(serviceID))
}

func (w *Watch) makeServiceStatusMessage(serviceID, pre string) string {
	var buff bytes.Buffer

	if pre != "" {
		buff.WriteString(pre)
		buff.WriteString(" ")
	}

	buff.WriteString(w.coloredServiceAddress(serviceID))

	return buff.String()
}

func (w *Watch) setFinalStates() (err error) {
	var states = w.sActivities.getFinalActivitiesStates()
	w.states = states

	if len(states.BuildFailed) == 0 &&
		len(states.DeployFailed) == 0 &&
		len(states.DeployCanceled) == 0 &&
		len(states.DeployTimeout) == 0 &&
		len(states.DeployRollback) == 0 {
		return nil
	}

	var emsgs []string

	for _, s := range states.BuildFailed {
		emsgs = append(emsgs, fmt.Sprintf(`error building service "%s"`, s))
	}

	for _, s := range states.DeployFailed {
		emsgs = append(emsgs, fmt.Sprintf(`error deploying service "%s"`, s))
	}

	for _, s := range states.DeployCanceled {
		emsgs = append(emsgs, fmt.Sprintf(`canceled deployment of service "%s"`, s))
	}

	for _, s := range states.DeployTimeout {
		emsgs = append(emsgs, fmt.Sprintf(`timed out deploying "%s"`, s))
	}

	for _, s := range states.DeployRollback {
		emsgs = append(emsgs, fmt.Sprintf(`rolling back service "%s"`, s))
	}

	return errors.New(strings.Join(emsgs, "\n"))
}

func (w *Watch) updateActivityState(a activities.Activity) {
	var serviceID, ok = a.Metadata["serviceId"].(string)

	// stop processing if service is not any of the watched deployment cycle types,
	// or if service ID is somehow not available
	if !ok || !isActitityTypeDeploymentRelated(a.Type) {
		return
	}

	if _, exists := w.sActivities[serviceID]; !exists {
		// skip activity
		// we want to avoid problems (read: nil pointer panics)
		// if the server sends back a response with an ID we don't have already locally
		return
	}

	w.markDeploymentTransition(serviceID, a.Type)
	var m = w.sActivities[serviceID].msgWLM
	var pre string

	if pre, ok = activities.FriendlyActivities[a.Type]; !ok {
		pre = a.Type
	}

	switch a.Type {
	case activities.BuildStarted,
		activities.BuildPushed,
		activities.BuildSucceeded,
		activities.DeployCreated,
		activities.DeployPending,
		activities.DeployStarted:
		m.PlayText(w.makeServiceStatusMessage(serviceID, pre))
	case
		activities.BuildFailed,
		activities.DeployFailed,
		activities.DeployCanceled,
		activities.DeployTimeout,
		activities.DeployRollback:
		m.StopText(fancy.Error(w.makeServiceStatusMessage(serviceID, pre)))
	case
		activities.DeploySucceeded:
		m.StopText(w.makeServiceStatusMessage(serviceID, pre))
	default:
		m.StopText(w.makeServiceStatusMessage(serviceID, pre))
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
		activities.DeployCanceled,
		activities.DeployTimeout,
		activities.DeployRollback,
		activities.DeploySucceeded:
		return true
	}

	return false
}

// ValidTransitions for deployment states (from -> to)
var ValidTransitions = map[string][]string{
	activities.BuildFailed: []string{
		activities.BuildSucceeded,
		activities.DeployFailed,
		activities.DeployCanceled,
		activities.DeployTimeout,
		activities.DeployRollback,
		activities.DeploySucceeded,
	},
	activities.BuildSucceeded: []string{
		activities.DeployFailed,
		activities.DeployCanceled,
		activities.DeployTimeout,
		activities.DeployRollback,
		activities.DeploySucceeded,
	},
	activities.DeployFailed: []string{
		activities.DeploySucceeded,
	},
	activities.DeployCanceled: []string{
		activities.DeploySucceeded,
	},
	activities.DeployTimeout: []string{
		activities.DeploySucceeded,
	},
	activities.DeployRollback: []string{
		activities.DeploySucceeded,
	},
	activities.DeploySucceeded: []string{},
}

func (w *Watch) markDeploymentTransition(serviceID, activityType string) {
	var old = w.sActivities[serviceID].state

	var validFrom, ok = ValidTransitions[old]

	if !ok {
		w.sActivities[serviceID].state = activityType

		if old != "" {
			verbose.Debug("Unexpected Transition: from:", old, "to:", activityType)
		}

		return
	}

	for _, eachValid := range validFrom {
		if activityType == eachValid {
			w.sActivities[serviceID].state = activityType
			return
		}
	}
}

func (w *Watch) updateActivitiesStates() (end bool, err error) {
	var as activities.Activities
	var ctx, cancel = context.WithTimeout(w.ctx, 5*time.Second)
	defer cancel()

	activitiesClient := activities.New(w.ConfigContext)

	as, err = activitiesClient.List(verbosereq.ContextNoVerbose(ctx), w.ProjectID, activities.Filter{
		GroupUID: w.GroupUID,
	})
	cancel()
	as = as.Reverse()

	if err != nil {
		return isNotFoundError(err), err
	}

	for _, a := range as {
		w.updateActivityState(a)
	}

	for id := range w.sActivities {
		if !w.sActivities.isFinalState(id) {
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

func isNotFoundError(err error) bool {
	var aft = errwrap.GetType(err, &apihelper.APIFault{})

	if aft == nil {
		return false
	}

	af := aft.(*apihelper.APIFault)

	return af != nil && af.Status == 404
}

// Wait for deployment to finish.
func (w *Watch) Wait() error {
	defer w.maybeSetOpenLogsFunc()
	w.reorderDeployments()

	if err := w.waitFor(); err != nil {
		return err
	}

	switch err := w.setFinalStates(); err {
	case nil:
		w.notifyDeploymentSucceeded()
	default:
		w.notifyDeploymentFailed()
	}

	w.wlm.Stop()
	return nil
}

func (w *Watch) waitFor() error {
	rate := rate.NewLimiter(rate.Every(time.Second), 1)

	for {
		if er := rate.Wait(w.ctx); er != nil {
			verbose.Debug(er)
		}

		var end, err = w.updateActivitiesStates()
		var stepText = w.header.GetText()

		if err != nil {
			w.header.PlayText(updateMessageErrorStringCounter(stepText))
			verbose.Debug(err)
		}

		if err != nil && end {
			return errors.New("deployment not found")
		}

		if err == nil && strings.Contains(stepText, "retrying to get status #") {
			w.header.PlayText(clearMessageErrorStringCounter(stepText))
		}

		if end {
			return nil
		}
	}
}

func (w *Watch) maybeSetOpenLogsFunc() {
	states := w.states
	askLogs := (len(states.BuildFailed) != 0 || len(states.DeployFailed) != 0)

	if askLogs {
		errorhandler.SetAfterError(func() {
			w.maybeOpenLogs()
		})
	}
}

func (w *Watch) maybeOpenLogs() {
	var states = w.states
	var failedBuilds = states.BuildFailed
	var failedDeploys = states.DeployFailed

	switch yes, err := fancy.Boolean("Open browser to check the logs?"); {
	case err != nil:
		fmt.Fprintf(os.Stderr, "%v", err)
		fallthrough
	case !yes:
		return
	}

	var logsURL = fmt.Sprintf("https://%v%v/projects/%v/logs",
		defaults.DashboardAddressPrefix,
		w.ConfigContext.InfrastructureDomain(),
		w.ProjectID)

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

type serviceWatch struct {
	state  string
	msgWLM *waitlivemsg.Message
}

type servicesMap map[string]*serviceWatch

func (s servicesMap) getFinalActivitiesStates() finalStates {
	var states = finalStates{}

	for k, service := range s {
		switch service.state {
		case activities.BuildFailed:
			states.BuildFailed = append(states.BuildFailed, k)
		case activities.DeployFailed:
			states.DeployFailed = append(states.DeployFailed, k)
		case activities.DeployCanceled:
			states.DeployCanceled = append(states.DeployCanceled, k)
		case activities.DeployTimeout:
			states.DeployTimeout = append(states.DeployTimeout, k)
		case activities.DeployRollback:
			states.DeployRollback = append(states.DeployRollback, k)
		}
	}

	return states
}

func (s servicesMap) isFinalState(key string) bool {
	if s == nil || s[key] == nil {
		return false
	}

	switch s[key].state {
	case activities.BuildFailed,
		activities.DeployFailed,
		activities.DeployCanceled,
		activities.DeployTimeout,
		activities.DeployRollback,
		activities.DeploySucceeded:
		return true
	}

	return false
}
