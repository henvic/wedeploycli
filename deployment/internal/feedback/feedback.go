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
	"sync"
	"time"

	humanize "github.com/dustin/go-humanize"
	"github.com/hashicorp/errwrap"
	"github.com/henvic/ctxsignal"
	"github.com/henvic/uilive"
	"github.com/wedeploy/cli/activities"
	"github.com/wedeploy/cli/apihelper"
	"github.com/wedeploy/cli/color"
	"github.com/wedeploy/cli/config"
	"github.com/wedeploy/cli/exiterror"
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

const uncancelableExitCode = 64

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

	states            map[string]*swatch
	onlyBuildServices map[string]struct{}

	f final
}

// Start watching deployment feedback.
func (w *Watch) Start(ctx context.Context) {
	w.ctx = ctx
	w.start = time.Now()

	w.stepMessage = &waitlivemsg.Message{}
	w.header = &waitlivemsg.Message{}
	w.wlm = waitlivemsg.WaitLiveMsg{}

	w.states = map[string]*swatch{}
	w.onlyBuildServices = map[string]struct{}{}

	if w.Quiet && w.IsUpload {
		w.prepareQuiet()
		fmt.Println("Uploading.")
	}

	if !w.SkipProgress {
		w.prepareProgress()
	}

	w.createServicesActivitiesMap()
}

// PrintPackageSize related to the created temporary git repo.
func (w *Watch) PrintPackageSize(b uint64) {
	m := &waitlivemsg.Message{}
	msg := "Package size: " + humanize.Bytes(b)

	m.StopText(figures.Tick + " " + msg)
	w.wlm.AddMessage(m)

	if w.Quiet {
		_, _ = fmt.Fprintf(os.Stderr, "%s\n", msg)
	}
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
	w.notifyFailed()

	for serviceID, sw := range w.states {
		sw.msgWLMMutex.RLock()
		sw.msgWLM.PlayText(fancy.Error(w.makeServiceStatusMessage(serviceID, "Upload failed")))
		sw.msgWLMMutex.RUnlock()
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

type final struct {
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
	for _, s := range w.Services {
		var m = &waitlivemsg.Message{}
		msg := "Waiting"

		if w.IsUpload {
			msg = "Uploading"
		}

		m.PlayText(w.makeServiceStatusMessage(s.ServiceID, msg))

		w.states[s.ServiceID] = &swatch{
			ServiceID: s.ServiceID,
			msgWLM:    m,
			visited:   map[string]bool{},
		}

		w.wlm.AddMessage(m)
	}
}

func (w *Watch) markBuildOnlyServices() {
	projectsClient := projects.New(w.ConfigContext)
	bs, err := projectsClient.GetBuilds(w.ctx, w.ProjectID, w.GroupUID)

	if err != nil {
		fmt.Fprintf(os.Stderr, "can't get builds: %v\n", err)
		return
	}

	for _, b := range bs {
		if w.OnlyBuild || b.SkippedDeploy() {
			w.onlyBuildServices[b.ServiceID] = struct{}{}
		}
	}
}

func (w *Watch) reorderDeployments() {
	projectsClient := projects.New(w.ConfigContext)
	order, _ := projectsClient.GetDeploymentOrder(w.ctx, w.ProjectID, w.GroupUID)

	for _, do := range order {
		if a, ok := w.states[do]; ok {
			a.msgWLMMutex.RLock()
			w.wlm.RemoveMessage(a.msgWLM)
			a.msgWLMMutex.RUnlock()
		}
	}

	for _, do := range order {
		if a, ok := w.states[do]; ok {
			a.msgWLMMutex.RLock()
			w.wlm.AddMessage(a.msgWLM)
			a.msgWLMMutex.RUnlock()
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
	w.f = w.getFinalActivitiesStates()

	if len(w.f.BuildFailed) == 0 &&
		len(w.f.DeployFailed) == 0 &&
		len(w.f.DeployCanceled) == 0 &&
		len(w.f.DeployTimeout) == 0 &&
		len(w.f.DeployRollback) == 0 {
		return nil
	}

	var emsgs []string

	for _, s := range w.f.BuildFailed {
		emsgs = append(emsgs, fmt.Sprintf(`error building service "%s"`, s))
	}

	for _, s := range w.f.DeployFailed {
		emsgs = append(emsgs, fmt.Sprintf(`error deploying service "%s"`, s))
	}

	for _, s := range w.f.DeployCanceled {
		emsgs = append(emsgs, fmt.Sprintf(`canceled deployment of service "%s"`, s))
	}

	for _, s := range w.f.DeployTimeout {
		emsgs = append(emsgs, fmt.Sprintf(`timed out deploying "%s"`, s))
	}

	for _, s := range w.f.DeployRollback {
		emsgs = append(emsgs, fmt.Sprintf(`rolling back service "%s"`, s))
	}

	return errors.New(strings.Join(emsgs, "\n"))
}

func (w *Watch) getDecoratedFriendlyActivity(t, serviceID string) string {
	friendly, ok := activities.Friendly[t]

	if !ok {
		return t
	}

	if w.Quiet {
		return friendly
	}

	switch t {
	case activities.BuildFailed,
		activities.DeployFailed,
		activities.DeployCanceled,
		activities.DeployTimeout,
		activities.DeployRollback:
		return figures.Cross + " " + friendly
	}

	var onlyBuild = w.skippedDeploy(serviceID)

	switch {
	case onlyBuild && t == activities.BuildSucceeded,
		!onlyBuild && t == activities.DeploySucceeded:
		return figures.Tick + " " + friendly
	}

	return friendly
}

func (w *Watch) updateActivitiesStateMessage(serviceID, t string) {
	var skipDeploy = w.skippedDeploy(serviceID)

	var friendly = w.getDecoratedFriendlyActivity(t, serviceID)

	var ssm = w.makeServiceStatusMessage(serviceID, friendly)

	if skipDeploy {
		w.updateBuildActivitiesStateMessage(serviceID, t, ssm)
		return
	}

	w.updateDeployActivitiesStateMessage(serviceID, t, ssm)
}

func (w *Watch) updateBuildActivitiesStateMessage(serviceID, t, ssm string) {
	var sw = w.states[serviceID]
	sw.msgWLMMutex.RLock()
	var m = sw.msgWLM
	sw.msgWLMMutex.RUnlock()

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

	w.maybePrintQuietActivity(sw, t, ssm)
}

func (w *Watch) updateDeployActivitiesStateMessage(serviceID, t, ssm string) {
	var sw = w.states[serviceID]
	sw.msgWLMMutex.RLock()
	var m = sw.msgWLM
	sw.msgWLMMutex.RUnlock()

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

	w.maybePrintQuietActivity(sw, t, ssm)
}

func (w *Watch) maybePrintQuietActivity(sw *swatch, aType, text string) {
	if w.Quiet && !sw.visited[aType] {
		fmt.Println(text)
	}
}

func (w *Watch) isValidState(serviceID string, activityType string) bool {
	if w.skippedDeploy(serviceID) {
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

func (w *Watch) updateActivitiesStates() (end bool, err error) {
	var as []activities.Activity
	var ctx, cancel = context.WithTimeout(w.ctx, 5*time.Second)
	defer cancel()

	activitiesClient := activities.New(w.ConfigContext)

	as, err = activitiesClient.List(verbosereq.ContextNoVerbose(ctx), w.ProjectID, activities.Filter{
		GroupUID: w.GroupUID,
	})
	cancel()

	if err != nil {
		return isNotFoundError(err), err
	}

	w.sync(as)

	for serviceID := range w.states {
		if !w.isFinalState(serviceID) {
			return false, nil
		}
	}

	return true, nil
}

func (w *Watch) sync(as []activities.Activity) {
	for _, a := range as {
		// stop processing if service ID is not available or state transition is invalid
		var serviceID, ok = a.Metadata["serviceId"].(string)

		if _, exists := w.states[serviceID]; !exists {
			continue
		}

		if !ok || !w.isValidState(serviceID, a.Type) {
			continue
		}

		var sw = w.states[serviceID]

		if _, ok := sw.visited[a.Type]; ok {
			continue
		}

		w.updateActivitiesStateMessage(serviceID, a.Type)
		sw.current = a.Type
		sw.visited[a.Type] = true
	}
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
	var aft = errwrap.GetType(err, apihelper.APIFault{})

	if aft == nil {
		return false
	}

	af, ok := aft.(apihelper.APIFault)
	return ok && af.Status == 404
}

// Wait for deployment to finish.
func (w *Watch) Wait() error {
	defer w.maybeSetOpenLogsFunc()

	w.markBuildOnlyServices()
	w.reorderDeployments()

	if err := w.wait(); err != nil {
		return w.handleWaitError(err)
	}

	err := w.setFinalStates()

	switch err {
	case nil:
		w.notifySucceeded()
	default:
		w.notifyFailed()
	}

	w.wlm.Stop()

	return err
}

func (w *Watch) wait() error {
	rate := rate.NewLimiter(rate.Every(time.Second), 1)

	for {
		if er := rate.Wait(w.ctx); er != nil {
			verbose.Debug(er)
		}

		if err := w.ctx.Err(); err != nil {
			return err
		}

		var end, err = w.updateActivitiesStates()
		var stepText = w.header.GetText()

		if err != nil {
			umerr := updateMessageErrorStringCounter(stepText)

			w.header.PlayText(umerr)

			if _, errCtx := ctxsignal.Closed(w.ctx); errCtx != nil && w.Quiet {
				_, _ = fmt.Fprintln(os.Stderr, umerr)
			}

			verbose.Debug(err)
		}

		if err != nil && end {
			return err
		}

		if err == nil && strings.Contains(stepText, "retrying to get status #") {
			w.header.PlayText(clearMessageErrorStringCounter(stepText))
		}

		if end {
			return nil
		}
	}
}

func (w *Watch) handleWaitError(err error) error {
	w.finishError()

	if err != w.ctx.Err() {
		return err
	}

	s, err := ctxsignal.Closed(w.ctx)

	if err != nil {
		return err
	}

	fmt.Println()
	verbose.Debug(s)

	if w.OnlyBuild {
		return exiterror.New(
			"Too late to cancel build: please check your project logs and activities",
			uncancelableExitCode)
	}

	return exiterror.New(
		"Too late to cancel deployment: please check your project logs and activities",
		uncancelableExitCode)
}

func (w *Watch) finishError() {
	w.header.StopText(figures.Cross + " " + w.getActionMessage())

	if w.IsUpload && w.uploadCompleted != 0 {
		var msg = fmt.Sprintf(figures.Tick+" Upload completed in %v.",
			timehelper.RoundDuration(w.uploadCompleted, time.Second))

		w.stepMessage.StopText(msg)

		if w.Quiet {
			_, _ = fmt.Fprintf(os.Stderr, "\n%s\n", msg)
		}
	}

	for _, sw := range w.states {
		if !w.isFinalState(sw.ServiceID) {
			sw.msgWLMMutex.RLock()
			sw.msgWLM.StopText("? " + sw.msgWLM.GetText())
			sw.msgWLMMutex.RUnlock()
		}
	}

	w.wlm.Stop()
}

type swatch struct {
	ServiceID string

	current string
	visited map[string]bool // deployment transitions that already happened

	msgWLM      *waitlivemsg.Message
	msgWLMMutex sync.RWMutex
}

func (w *Watch) getFinalActivitiesStates() final {
	var f = final{}

	for k, service := range w.states {
		switch service.current {
		case activities.BuildFailed:
			f.BuildFailed = append(f.BuildFailed, k)
		case activities.DeployFailed:
			f.DeployFailed = append(f.DeployFailed, k)
		case activities.DeployCanceled:
			f.DeployCanceled = append(f.DeployCanceled, k)
		case activities.DeployTimeout:
			f.DeployTimeout = append(f.DeployTimeout, k)
		case activities.DeployRollback:
			f.DeployRollback = append(f.DeployRollback, k)
		}
	}

	return f
}

func (w *Watch) isFinalState(key string) bool {
	if w.states == nil || w.states[key] == nil {
		return false
	}

	if w.skippedDeploy(key) {
		return w.isBuildFinalState(key)
	}

	return w.isDeployFinalState(key)
}

func (w *Watch) isBuildFinalState(key string) bool {
	var state = w.states[key].current

	if state == activities.BuildFailed || state == activities.BuildSucceeded {
		return true
	}

	return false
}

func (w *Watch) isDeployFinalState(key string) bool {
	var sw = w.states[key]

	switch sw.current {
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

func (w *Watch) skippedDeploy(serviceID string) bool {
	skip := w.OnlyBuild

	if !skip {
		_, skip = w.onlyBuildServices[serviceID]
	}

	return skip
}
