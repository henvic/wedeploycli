package activities

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"bytes"

	"github.com/hashicorp/errwrap"
	"github.com/henvic/uilive"
	"github.com/pkg/browser"
	"github.com/wedeploy/cli/color"
	"github.com/wedeploy/cli/config"
	"github.com/wedeploy/cli/prompt"
	"github.com/wedeploy/cli/timehelper"
	"github.com/wedeploy/cli/verbose"
	"github.com/wedeploy/cli/waitlivemsg"
)

// DeployWatcher activities
type DeployWatcher struct {
	ctx       context.Context
	projectID string
	timeout   time.Duration
	filter    Filter
	services  servicesMap
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

var timeout = time.Minute

func (s servicesMap) isFinalState(key string) bool {
	if s == nil || s[key] == nil {
		return false
	}

	switch s[key].state {
	case buildFailed, deployFailed, deploySucceeded:
		return true
	}

	return false
}

func (dw *DeployWatcher) updateActivityState(a Activity) {
	var serviceID, ok = a.Metadata["serviceId"]

	// stop processing if service is not any of the watched deployment cycle types,
	// or if service ID is somehow not available
	if !ok || !isActitityTypeDeploymentRelated(a.Type) {
		return
	}

	if _, exists := dw.services[serviceID]; !exists {
		// skip activity
		// we want to avoid problems (read: nil pointer panics)
		// if the server sends back a response with an ID we don't have already locally
		return
	}

	dw.markActivityState(serviceID, a.Type)
	var wlm = dw.services[serviceID].msgWLM
	var pre, post string

	var suffixes = map[string]string{
		buildFailed:     "build failed",
		buildPending:    "build pending",
		buildStarted:    "build started",
		buildSucceeded:  "build successful",
		deployFailed:    "deploy failed",
		deployPending:   "deploy pending",
		deploySucceeded: "deployed " + serviceID,
		deployStarted:   "", // should not show a suffix
	}

	switch a.Type {
	case buildFailed, buildPending, buildStarted, buildSucceeded:
		pre = "building"
	case deployFailed, deployPending, deployStarted, deploySucceeded:
		pre = "deploying"
	}

	if post, ok = suffixes[a.Type]; !ok {
		post = a.Type
	}

	wlm.SetText(dw.makeServiceStatusMessage(serviceID, pre, post))

	switch a.Type {
	case buildFailed, deployFailed:
		wlm.SetSymbolEnd(waitlivemsg.RedCrossSymbol())
		wlm.End()
	case deploySucceeded:
		wlm.End()
	}
}

func isActitityTypeDeploymentRelated(activityType string) bool {
	switch activityType {
	case buildPending,
		buildStarted,
		buildFailed,
		buildSucceeded,
		deployPending,
		deployStarted,
		deployFailed,
		deploySucceeded:
		return true
	}

	return false
}

func (dw *DeployWatcher) markActivityState(serviceID, activityType string) {
	switch activityType {
	case buildSucceeded,
		buildFailed,
		deployFailed,
		deploySucceeded:
		dw.services[serviceID].state = activityType
	}
}

func (dw *DeployWatcher) loop() (end bool, err error) {
	var as Activities
	as, err = List(dw.ctx, dw.projectID, dw.filter)
	as = as.Reverse()

	if err != nil {
		return false, err
	}

	for _, a := range as {
		dw.updateActivityState(a)
	}

	for id := range dw.services {
		if !dw.services.isFinalState(id) {
			return false, nil
		}
	}

	return true, nil
}

func (dw *DeployWatcher) runLoop() error {
l:
	switch end, err := dw.loop(); {
	case err != nil:
		return err
	case !end:
		goto l
	}

	return nil
}

func (dw *DeployWatcher) printServiceAddress(service string) string {
	var address = dw.projectID + "." + config.Context.RemoteAddress

	if service != "" {
		address = service + "-" + address
	}

	return address
}

func (dw *DeployWatcher) makeServiceStatusMessage(serviceID, pre string, posts ...string) string {
	var buff bytes.Buffer

	if pre != "" {
		buff.WriteString(pre)
		buff.WriteString(" ")
	}

	buff.WriteString(color.Format(
		color.FgBlue,
		dw.printServiceAddress(serviceID)))

	var post = strings.Join(posts, " ")

	if post != "" {
		buff.WriteString(" (")
		buff.WriteString(post)
		buff.WriteString(")")
	}

	return buff.String()
}

// Run the activities watcher
func (dw *DeployWatcher) Run(ctx context.Context, projectID string, services []string, f Filter) error {
	dw.projectID = projectID
	dw.services = servicesMap{}
	dw.filter = f

	if len(services) == 0 {
		return errors.New("services parameter required for listening to services deployment")
	}

	var wlm = waitlivemsg.WaitLiveMsg{}
	var us = uilive.New()
	wlm.SetStream(us)

	for _, s := range services {
		var m = waitlivemsg.NewMessage(dw.makeServiceStatusMessage(s, "waiting for deployment"))

		dw.services[s] = &serviceWatch{
			msgWLM: m,
		}
		wlm.AddMessage(m)
	}

	go wlm.Wait()

	var err = dw.runLoop()
	wlm.Stop()

	var after = color.Format(color.FgBlue, "%v",
		timehelper.RoundDuration(wlm.Duration(), time.Second))

	if err != nil {
		return errwrap.Wrapf("deployment failure: {{err}}", err)
	}

	return dw.checkSuccess(after)
}

func (dw *DeployWatcher) checkSuccess(timeElapsed string) error {
	var (
		fb       = dw.services.GetServicesByState(buildFailed)
		fd       = dw.services.GetServicesByState(deployFailed)
		feedback string
	)

	fmt.Println("")

	if len(fb) == 0 && len(fd) == 0 {
		fmt.Printf("Deployment successful in %v\n", timeElapsed)
		return nil
	}

	dw.maybeOpenLogs(fb, fd)

	switch len(dw.services) {
	case len(fb) + len(fd):
		feedback = "deployment failed partially in %v"
	default:
		feedback = "deployment failed in %v"
	}

	return errors.New(feedback)
}

func (dw *DeployWatcher) maybeOpenLogs(failedBuilds, failedDeploys []string) {
shouldOpenPrompt:
	var p, err = prompt.Prompt("Do you want to check the logs (yes/no)? [no]")

	if err != nil {
		verbose.Debug(err)
		return
	}

	switch p {
	case "y":
		break
	case "no", "n":
		return
	default:
		goto shouldOpenPrompt
	}

	var logsURL = fmt.Sprintf("https://%v/projects/%v/logs", config.Context.RemoteAddress, dw.projectID)

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
