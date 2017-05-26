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
	ctx          context.Context
	projectID    string
	timeout      time.Duration
	filter       Filter
	services     []string
	servicesMsgs map[string]string
	finalStates  map[string]bool
	buildOK      map[string]bool
	buildFail    map[string]bool
	deployOK     map[string]bool
	deployFail   map[string]bool
}

var timeout = time.Minute

var servicesDeploymentActivityTemplates = map[string]string{
	buildFailed:     "{{.Metadata.serviceId}} build failed",
	buildPending:    "{{.Metadata.serviceId}} build pending",
	buildStarted:    "{{.Metadata.serviceId}} build started",
	buildSucceeded:  "{{.Metadata.serviceId}} build successful",
	deployFailed:    "{{.Metadata.serviceId}} deployment failed",
	deployPending:   "{{.Metadata.serviceId}} deployment pending",
	deployStarted:   "{{.Metadata.serviceId}} deployment started",
	deploySucceeded: "{{.Metadata.serviceId}} deployment successful",
}

func (dw *DeployWatcher) iterator(a Activity) {
	var serviceID, ok = a.Metadata["serviceId"]
	var err error

	// stop processing if service has already reached end state,
	// type is not any of the watched deployment cycle types,
	// or if service ID is somehow not available
	if !ok || dw.finalStates[serviceID] || !isActitityTypeDeploymentRelated(a.Type) {
		return
	}

	if _, exists := dw.finalStates[serviceID]; !exists {
		// we want to avoid problems (read: nil pointer panics)
		// if the server sends back a response with an ID we don't have already locally
		fmt.Fprintf(os.Stderr, "skipping activity for non-existing %v service on deployment list: %+v", serviceID, a)
		return
	}

	dw.servicesMsgs[serviceID], err = getActivityMessage(a, servicesDeploymentActivityTemplates)

	if err != nil {
		fmt.Fprintf(os.Stderr, "can't identify activity for %v service: %+v\n", serviceID, err)
	}

	dw.markActivityState(serviceID, a.Type)
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
	case buildSucceeded:
		dw.buildOK[serviceID] = true
	case buildFailed:
		dw.buildFail[serviceID] = true
		dw.finalStates[serviceID] = true
	case deployFailed:
		dw.deployFail[serviceID] = true
		dw.finalStates[serviceID] = true
	case deploySucceeded:
		dw.deployOK[serviceID] = true
		dw.finalStates[serviceID] = true
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
		dw.iterator(a)
	}

	for _, c := range dw.finalStates {
		if !c {
			return false, nil
		}
	}

	return true, nil
}

// NewDeployWatcher creates a deployment watcher
func NewDeployWatcher(ctx context.Context, projectID string, expectServices []string, f Filter) *DeployWatcher {
	return &DeployWatcher{
		ctx:          ctx,
		projectID:    projectID,
		services:     expectServices,
		filter:       f,
		servicesMsgs: map[string]string{},
		finalStates:  map[string]bool{},
		buildOK:      map[string]bool{},
		buildFail:    map[string]bool{},
		deployOK:     map[string]bool{},
		deployFail:   map[string]bool{},
	}
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
func (dw *DeployWatcher) Run() error {
	if len(dw.services) == 0 {
		return errors.New("services parameter required for listening to services deployment")
	}

	for _, c := range dw.services {
		dw.finalStates[c] = false
		dw.servicesMsgs[c] = "waiting signal"
	}

	var wlm = waitlivemsg.WaitLiveMsg{}
	var us = uilive.New()
	wlm.SetStream(us)
	wlm.SetTickSymbolEnd()
	wlm.SetMessage("Deploying")
	go wlm.Wait()

	var err = dw.runLoop()

	var after = color.Format(color.FgBlue, "%v",
		timehelper.RoundDuration(wlm.Duration(), time.Second))

	if err != nil {
		wlm.SetCrossSymbolEnd()
		wlm.StopWithMessage(fmt.Sprintf("Deploy failed after %v", after))
		return err
	}

	wlm.StopWithMessage(fmt.Sprintf("Deploy completed in %v", after))

	return dw.checkSuccess()
}

func getTruthKeysOnMap(m map[string]bool) (truth []string) {
	for k, c := range m {
		if c {
			truth = append(truth, k)
		}
	}

	return truth
}

func (dw *DeployWatcher) checkSuccess() error {
	var (
		fb = getTruthKeysOnMap(dw.buildFail)
		fd = getTruthKeysOnMap(dw.deployFail)
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
		em = fmt.Sprintf("failed deploys: %v", failedDeploys)
	}

	return errors.New(em)
}
