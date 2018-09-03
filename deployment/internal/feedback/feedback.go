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

	OnlyBuild    bool
	SkipProgress bool
	Quiet        bool

	ctx context.Context

	IsUpload        bool
	uploadCompleted time.Duration

	start time.Time

	wlm         waitlivemsg.WaitLiveMsg
	header      *waitlivemsg.Message
	stepMessage *waitlivemsg.Message

	transitions map[string][]string
	activities  servicesMap

	states finalStates
}

// Start watching deployment feedback.
func (w *Watch) Start(ctx context.Context) {
	w.ctx = ctx
	w.start = time.Now()

	w.stepMessage = &waitlivemsg.Message{}
	w.header = &waitlivemsg.Message{}
	w.wlm = waitlivemsg.WaitLiveMsg{}

	w.setValidTransitions()

	if w.Quiet && w.IsUpload {
		w.prepareQuiet()
		fmt.Println("Uploading.")
	}

	if !w.SkipProgress {
		w.prepareProgress()
	}

	w.createServicesActivitiesMap()
}

func (w *Watch) setValidTransitions() {
	if w.OnlyBuild {
		w.transitions = ValidBuildTransitions
		return
	}

	w.transitions = ValidDeployTransitions
}

func (w *Watch) notifySucceeded() {
	var templateMsg = "%s Deployment succeeded in %s"

	if w.OnlyBuild {
		templateMsg = "%s Build succeeded in %s"
	}

	var timeElapsed = timehelper.RoundDuration(time.Since(w.start), time.Second)

	w.header.StopText(figures.Tick + " " + w.getActionMessage())

	msg := fmt.Sprintf(templateMsg, figures.Tick, timeElapsed)
	msgWithUploadTime := msg

	if w.IsUpload && w.uploadCompleted != 0 {
		msgWithUploadTime = fmt.Sprintf("%s. Upload completed in %v.",
			msg,
			timehelper.RoundDuration(w.uploadCompleted, time.Second))
	}

	w.stepMessage.StopText(msgWithUploadTime)

	if w.Quiet {
		fmt.Printf("\n%s\n", msg)
	}
}

func (w *Watch) notifyFailed() {
	var templateMsg = "%s Deployment failed in %s"

	if w.OnlyBuild {
		templateMsg = "%s Build failed in %s"
	}

	var timeElapsed = timehelper.RoundDuration(time.Since(w.start), time.Second)

	w.header.StopText(figures.Cross + " " + w.getActionMessage())

	msg := fmt.Sprintf(templateMsg, figures.Cross, timeElapsed)

	if w.IsUpload && w.uploadCompleted != 0 {
		msg = fmt.Sprintf("%s. Upload completed in %v.",
			msg,
			timehelper.RoundDuration(w.uploadCompleted, time.Second))
	}

	w.stepMessage.StopText(msg)

	if w.Quiet {
		_, _ = fmt.Fprintf(os.Stderr, "\n%s\n", msg)
	}
}

// PrintSkipProgress prints a message telling that the deployment
// is happening on background when the deployment is not interactive.
func (w *Watch) PrintSkipProgress() {
	fmt.Printf("Deployment %v is in progress on remote %v\n",
		color.Format(color.FgMagenta, color.Bold, w.GroupUID),
		color.Format(color.FgMagenta, color.Bold, w.ConfigContext.InfrastructureDomain()))
}

// StopFailedUpload stops the deployment messages due to a failed upload error.
func (w *Watch) StopFailedUpload() {
	if !w.Quiet {
		return
	}

	w.notifyFailed()

	for serviceID, s := range w.activities.states {
		s.msgWLM.PlayText(fancy.Error(w.makeServiceStatusMessage(serviceID, "Upload failed")))
	}

	w.wlm.Stop()
}

// NotifyPacking notifies that a package is being prepared for deployment.
func (w *Watch) NotifyPacking() {
	w.header.PlayText(fmt.Sprintf("Preparing deployment for project %v in %v...",
		color.Format(color.FgMagenta, color.Bold, w.ProjectID),
		color.Format(w.ConfigContext.Remote())))
}

// NotifyStart notifies that the deployment started.
func (w *Watch) NotifyStart() {
	w.header.PlayText(w.getActionMessage())
}

// NotifyUploadComplete notifies that the upload has been completed.
func (w *Watch) NotifyUploadComplete(t time.Duration) {
	if !w.IsUpload {
		return
	}

	w.uploadCompleted = t

	uploadCompletedFeedback := fmt.Sprintf("%s Upload completed in %v.",
		figures.Tick,
		timehelper.RoundDuration(t, time.Second))

	w.stepMessage.StopText(uploadCompletedFeedback)

	if w.Quiet {
		_, _ = fmt.Fprintf(os.Stderr, "%s\n\n", uploadCompletedFeedback)
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
	fmt.Println(w.getActionMessage())

	if len(w.Services) > 0 {
		fmt.Print("\nList of services:\n")
	}

	for _, s := range w.Services {
		fmt.Print(w.coloredServiceAddress(s.ServiceID))
		fmt.Println()
	}

	fmt.Println()
}

func (w *Watch) prepareProgress() {
	var us = uilive.New()
	w.wlm.SetStream(us)

	if !w.Quiet {
		go w.wlm.Wait()
		w.header.PlayText(w.getActionMessage())
		w.stepMessage.StopText("")
		w.wlm.AddMessage(w.header)
		w.wlm.AddMessage(w.stepMessage)
	}
}

func (w *Watch) createServicesActivitiesMap() {
	w.activities = servicesMap{
		states: map[string]*serviceWatch{},

		buildOnly: w.OnlyBuild,
	}

	for _, s := range w.Services {
		var m = &waitlivemsg.Message{}
		msg := "Waiting"

		if w.IsUpload {
			msg = "Uploading"
		}

		m.PlayText(w.makeServiceStatusMessage(s.ServiceID, msg))

		w.activities.states[s.ServiceID] = &serviceWatch{
			msgWLM:  m,
			visited: map[string]bool{},
		}
		w.wlm.AddMessage(m)
	}
}

func (w *Watch) reorderDeployments() {
	projectsClient := projects.New(w.ConfigContext)
	order, _ := projectsClient.GetDeploymentOrder(w.ctx, w.ProjectID, w.GroupUID)

	for _, do := range order {
		if a, ok := w.activities.states[do]; ok {
			w.wlm.RemoveMessage(a.msgWLM)
		}
	}

	for _, do := range order {
		if a, ok := w.activities.states[do]; ok {
			w.wlm.AddMessage(a.msgWLM)
		}
	}
}

func (w *Watch) getActionMessage() string {
	templateMsg := "Deploying services to project %v on %v"

	if w.OnlyBuild {
		templateMsg = "Building services to project %v on %v"
	}

	return fmt.Sprintf(templateMsg,
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
	var states = w.activities.getFinalActivitiesStates()
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
	// stop processing if service ID is not available or state transition is invalid
	var serviceID, ok = a.Metadata["serviceId"].(string)

	if !ok || !w.isValidState(a.Type) {
		return
	}

	if _, exists := w.activities.states[serviceID]; !exists {
		// skip activity
		// we want to avoid problems (read: nil pointer panics)
		// if the server sends back a response with an ID we don't have already locally
		return
	}

	w.markTransition(serviceID, a.Type)
	w.updateActivitiesStateMessage(serviceID, a.Type)
	var sa = w.activities.states[serviceID]
	sa.visited[a.Type] = true
}

func (w *Watch) getDecoratedFriendlyActivity(t string) string {
	friendly, ok := activities.Friendly[t]

	if !ok {
		return t
	}

	switch t {
	case activities.BuildFailed,
		activities.DeployFailed,
		activities.DeployCanceled,
		activities.DeployTimeout,
		activities.DeployRollback:
		return figures.Cross + " " + friendly
	}

	switch {
	case w.OnlyBuild && t == activities.BuildSucceeded,
		!w.OnlyBuild && t == activities.DeploySucceeded:
		return figures.Tick + " " + friendly
	}

	return friendly
}

func (w *Watch) updateActivitiesStateMessage(serviceID, t string) {
	var friendly = w.getDecoratedFriendlyActivity(t)

	var ssm = w.makeServiceStatusMessage(serviceID, friendly)

	if w.OnlyBuild {
		w.updateBuildActivitiesStateMessage(serviceID, t, ssm)
		return
	}

	w.updateDeployActivitiesStateMessage(serviceID, t, ssm)
}

func (w *Watch) updateBuildActivitiesStateMessage(serviceID, t, ssm string) {
	var sa = w.activities.states[serviceID]
	var m = sa.msgWLM

	switch t {
	case activities.BuildStarted,
		activities.BuildPushed:
		m.PlayText(ssm)
	case
		activities.BuildFailed:
		ssm = fancy.Error(ssm)
		m.StopText(ssm)
	default:
		m.StopText(ssm)
	}

	w.maybePrintQuietActivity(sa, t, ssm)
}

func (w *Watch) updateDeployActivitiesStateMessage(serviceID, t, ssm string) {
	var sa = w.activities.states[serviceID]
	var m = sa.msgWLM

	switch t {
	case activities.BuildStarted,
		activities.BuildPushed,
		activities.BuildSucceeded,
		activities.DeployCreated,
		activities.DeployPending,
		activities.DeployStarted:
		m.PlayText(ssm)
	case
		activities.BuildFailed,
		activities.DeployFailed,
		activities.DeployCanceled,
		activities.DeployTimeout,
		activities.DeployRollback:
		ssm = fancy.Error(ssm)
		m.StopText(ssm)
	default:
		m.StopText(ssm)
	}

	w.maybePrintQuietActivity(sa, t, ssm)
}

func (w *Watch) maybePrintQuietActivity(sw *serviceWatch, aType, text string) {
	if w.Quiet && !sw.visited[aType] {
		fmt.Println(text)
	}
}

func (w *Watch) isValidState(activityType string) bool {
	if w.OnlyBuild {
		return w.isValidBuildState(activityType)
	}

	return w.isValidDeployState(activityType)
}

func (w *Watch) isValidBuildState(activityType string) bool {
	switch activityType {
	case
		activities.BuildStarted,
		activities.BuildFailed,
		activities.BuildPushed,
		activities.BuildSucceeded:
		return true
	}

	return false
}

func (w *Watch) isValidDeployState(activityType string) bool {
	if w.isValidBuildState(activityType) {
		return true
	}

	switch activityType {
	case
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

func (w *Watch) markTransition(serviceID, activityType string) {
	var old = w.activities.states[serviceID].state

	var validFrom, ok = w.transitions[old]

	if !ok {
		w.activities.states[serviceID].state = activityType

		if old != "" {
			verbose.Debug("Unexpected Transition: from:", old, "to:", activityType)
		}

		return
	}

	for _, eachValid := range validFrom {
		if activityType == eachValid {
			w.activities.states[serviceID].state = activityType
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

	for id := range w.activities.states {
		if !w.activities.isFinalState(id) {
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

	if err := w.wait(); err != nil {
		return err
	}

	err := w.setFinalStates()

	switch err {
	case nil:
		w.notifySucceeded()
	default:
		w.notifyFailed()
	}

	if !w.Quiet {
		w.wlm.Stop()
	}

	return err
}

func (w *Watch) wait() error {
	rate := rate.NewLimiter(rate.Every(time.Second), 1)

	for {
		if er := rate.Wait(w.ctx); er != nil {
			verbose.Debug(er)
		}

		var end, err = w.updateActivitiesStates()
		var stepText = w.header.GetText()

		if err != nil {
			umerr := updateMessageErrorStringCounter(stepText)

			w.header.PlayText(umerr)

			if w.Quiet {
				_, _ = fmt.Fprintln(os.Stderr, umerr)
			}

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
		_, _ = fmt.Fprintf(os.Stderr, "%v", err)
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
	state   string
	visited map[string]bool // deployment transitions that already happened

	msgWLM *waitlivemsg.Message
}

type servicesMap struct {
	states map[string]*serviceWatch

	buildOnly bool
}

func (s *servicesMap) getFinalActivitiesStates() finalStates {
	var states = finalStates{}

	for k, service := range s.states {
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

func (s *servicesMap) isFinalState(key string) bool {
	if s.states == nil || s.states[key] == nil {
		return false
	}

	if s.buildOnly {
		return s.isBuildFinalState(key)
	}

	return s.isDeployFinalState(key)
}

func (s *servicesMap) isBuildFinalState(key string) bool {
	var state = s.states[key].state

	if state == activities.BuildFailed || state == activities.BuildSucceeded {
		return true
	}

	return false
}

func (s *servicesMap) isDeployFinalState(key string) bool {
	var state = s.states[key].state

	switch state {
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
