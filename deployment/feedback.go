package deployment

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

	"github.com/henvic/browser"
	"github.com/wedeploy/cli/activities"
	"github.com/wedeploy/cli/color"
	"github.com/wedeploy/cli/defaults"
	"github.com/wedeploy/cli/fancy"
	"github.com/wedeploy/cli/projects"
	"github.com/wedeploy/cli/timehelper"
	"github.com/wedeploy/cli/verbose"
	"github.com/wedeploy/cli/waitlivemsg"
	"golang.org/x/time/rate"
)

type finalActivitiesStates struct {
	BuildFailed    []string
	DeployFailed   []string
	DeployCanceled []string
	DeployTimeout  []string
	DeployRollback []string
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

func (d *Deploy) reorderDeployments() {
	projectsClient := projects.New(d.ConfigContext)
	order, _ := projectsClient.GetDeploymentOrder(d.ctx, d.ProjectID, d.groupUID)

	for _, do := range order {
		if a, ok := d.sActivities[do]; ok {
			d.wlm.RemoveMessage(a.msgWLM)
		}
	}

	for _, do := range order {
		if a, ok := d.sActivities[do]; ok {
			d.wlm.AddMessage(a.msgWLM)
		}
	}
}

func (d *Deploy) updateDeploymentEndStep(err error) {
	var timeElapsed = timehelper.RoundDuration(d.wlm.Duration(), time.Second)

	switch err {
	case nil:
		d.stepMessage.StopText(d.getDeployingMessage() +
			fmt.Sprintf("\nDeployment succeeded in %s", timeElapsed))
	default:
		d.stepMessage.StopText(d.getDeployingMessage() +
			fmt.Sprintf("\nDeployment failed in %s", timeElapsed))
	}
}

func (d *Deploy) notifyDeploymentOnQuiet(err error) {
	if err != nil {
		return
	}

	fmt.Printf("Deployment %v is in progress on remote %v\n",
		color.Format(color.FgBlue, d.GetGroupUID()),
		color.Format(color.FgBlue, d.ConfigContext.InfrastructureDomain()))
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
		color.Format(color.FgBlue, d.ConfigContext.InfrastructureDomain()),
	)
}

func (d *Deploy) printServiceAddress(service string) string {
	var address = d.ProjectID + "." + d.ConfigContext.ServiceDomain()

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

func (d *Deploy) verifyFinalState() (states finalActivitiesStates, err error) {
	states = d.sActivities.getFinalActivitiesStates()

	if len(states.BuildFailed) == 0 &&
		len(states.DeployFailed) == 0 &&
		len(states.DeployCanceled) == 0 &&
		len(states.DeployTimeout) == 0 &&
		len(states.DeployRollback) == 0 {
		return states, nil
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

	return states, errors.New(strings.Join(emsgs, "\n"))
}

func (d *Deploy) updateActivityState(a activities.Activity) {
	var serviceID, ok = a.Metadata["serviceId"].(string)

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
		activities.BuildStarted:    "Build started",
		activities.BuildPushed:     "Build pushed",
		activities.BuildSucceeded:  "Build succeeded",
		activities.DeployFailed:    "Deployment failed",
		activities.DeployCanceled:  "Deployment canceled",
		activities.DeployTimeout:   "Deployment timed out",
		activities.DeployRollback:  "Deployment rollback",
		activities.DeployCreated:   "Deployment created",
		activities.DeployPending:   "Deployment pending",
		activities.DeploySucceeded: "Deployment succeeded",
		activities.DeployStarted:   "Deployment started",
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
		activities.DeployFailed,
		activities.DeployCanceled,
		activities.DeployTimeout,
		activities.DeployRollback:
		m.StopText(fancy.Error(d.makeServiceStatusMessage(serviceID, pre)))
	case
		activities.DeploySucceeded:
		m.StopText(d.makeServiceStatusMessage(serviceID, pre))
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
		activities.DeployCanceled,
		activities.DeployTimeout,
		activities.DeployRollback,
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
		activities.DeployCanceled,
		activities.DeployTimeout,
		activities.DeployRollback,
		activities.DeploySucceeded:
		d.sActivities[serviceID].state = activityType
	}
}

func (d *Deploy) checkActivities() (end bool, err error) {
	var as activities.Activities
	var ctx, cancel = context.WithTimeout(d.ctx, 5*time.Second)
	defer cancel()

	activitiesClient := activities.New(d.ConfigContext)

	as, err = activitiesClient.List(ctx, d.ProjectID, activities.Filter{
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

func (d *Deploy) watchDeployment() {
	d.reorderDeployments()

	rate := rate.NewLimiter(rate.Every(time.Second), 1)

	for {
		if er := rate.Wait(d.ctx); er != nil {
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

func (d *Deploy) maybeOpenLogs(states finalActivitiesStates) {
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
		d.ConfigContext.InfrastructureDomain(),
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

type serviceWatch struct {
	state  string
	msgWLM *waitlivemsg.Message
}

type servicesMap map[string]*serviceWatch

func (s servicesMap) getFinalActivitiesStates() finalActivitiesStates {
	var states = finalActivitiesStates{}

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
