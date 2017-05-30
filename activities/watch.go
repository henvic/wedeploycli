package activities

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/henvic/uilive"
	"github.com/wedeploy/cli/color"
	"github.com/wedeploy/cli/timehelper"
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

var sMetaServiceID = color.Format(color.FgBlue, "{{.Metadata.serviceId}}")
var servicesDeploymentActivityTemplates = map[string]string{
	buildFailed:    sMetaServiceID + " build failed",
	buildPending:   sMetaServiceID + " build pending",
	buildStarted:   sMetaServiceID + " build started",
	buildSucceeded: sMetaServiceID + " build successful",
	deployFailed:   sMetaServiceID + " deploy failed",
	deployPending:  sMetaServiceID + " deploy pending",
	deployStarted:  sMetaServiceID + " deploy started",
	// deploySucceeded: sMetaServiceID + " deployed successful in {{.time}}",
	deploySucceeded: sMetaServiceID + " deployed successful",
}

func (dw *DeployWatcher) updateActivityState(a Activity) {
	var serviceID, ok = a.Metadata["serviceId"]

	// stop processing if service is not any of the watched deployment cycle types,
	// or if service ID is somehow not available
	if !ok || !isActitityTypeDeploymentRelated(a.Type) {
		return
	}

	if _, exists := dw.services[serviceID]; !exists {
		// we want to avoid problems (read: nil pointer panics)
		// if the server sends back a response with an ID we don't have already locally
		fmt.Fprintf(os.Stderr,
			"skipping activity for non-existing %v service on deployment list: %+v\n",
			serviceID,
			a)
		return
	}

	var m, err = getActivityMessage(a, servicesDeploymentActivityTemplates)

	dw.markActivityState(serviceID, a.Type)
	var wlm = dw.services[serviceID].msgWLM

	if err != nil {
		wlm.SetText(color.Format(color.FgBlue, serviceID) + " error")
		fmt.Fprintf(os.Stderr, "%v activity error for %v service: %+v\n", a.Type, serviceID, err)
		return
	}

	wlm.SetText(m)
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
		var m = waitlivemsg.NewMessage(
			color.Format(color.FgBlue, s))

		dw.services[s] = &serviceWatch{
			msgWLM: m,
		}
		wlm.AddMessage(m)
	}

	go wlm.Wait()

	var err = dw.runLoop()
	defer wlm.Stop()

	for _, s := range services {
		wlm.RemoveMessage(dw.services[s].msgWLM)
	}

	var after = color.Format(color.FgBlue, "%v",
		timehelper.RoundDuration(wlm.Duration(), time.Second))

	if err != nil {
		var m = waitlivemsg.NewMessage(
			fmt.Sprintf("Deploy failed after %v", after),
		)

		m.SetSymbolEnd(waitlivemsg.RedCrossSymbol())
		wlm.AddMessage(m)
		return err
	}

	wlm.AddMessage(waitlivemsg.NewMessage(
		fmt.Sprintf("Deploy completed in %v", after),
	))

	return dw.checkSuccess()
}

func (dw *DeployWatcher) checkSuccess() error {
	var (
		fb = dw.services.GetServicesByState(buildFailed)
		fd = dw.services.GetServicesByState(deployFailed)
		em string
	)

	if len(fb) == 0 && len(fd) == 0 {
		return nil
	}

	var failedBuilds = strings.Join(fb, ", ")
	var failedDeploys = strings.Join(fd, ", ")

	if len(failedBuilds) == 0 {
		em = fmt.Sprintf("failed builds: %v", failedBuilds)
	}

	if len(failedDeploys) != 0 {
		em += " and "
	}

	if len(failedDeploys) == 0 {
		em = fmt.Sprintf("failed deployments: %v", failedDeploys)
	}

	return errors.New(em)
}
