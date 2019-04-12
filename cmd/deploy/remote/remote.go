package deployremote

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/wedeploy/cli/cmd/canceled"
	"github.com/wedeploy/cli/cmd/deploy/internal/getproject"
	"github.com/wedeploy/cli/cmd/internal/we"
	"github.com/wedeploy/cli/color"
	"github.com/wedeploy/cli/deployment"
	"github.com/wedeploy/cli/deployment/transport/git"
	"github.com/wedeploy/cli/deployment/transport/gogit"
	"github.com/wedeploy/cli/fancy"
	"github.com/wedeploy/cli/inspector"
	"github.com/wedeploy/cli/isterm"
	"github.com/wedeploy/cli/namesgenerator"
	"github.com/wedeploy/cli/services"
	"github.com/wedeploy/cli/verbose"
)

// RemoteDeployment of services
type RemoteDeployment struct {
	Params deployment.Params

	Experimental bool

	path     string
	services services.ServiceInfoList
	remap    []string
	ctx      context.Context
}

// Feedback about a deployment.
type Feedback struct {
	GroupUID string
	Services services.ServiceInfoList
}

// Run does the remote deployment procedures
func (rd *RemoteDeployment) Run(ctx context.Context) (f Feedback, err error) {
	rd.ctx = ctx
	wectx := we.Context()

	if rd.path, err = getWorkingDirectory(); err != nil {
		return f, err
	}

	rd.Params.ProjectID, err = getproject.MaybeID(rd.Params.ProjectID, rd.Params.Region)

	if err != nil {
		return f, err
	}

	err = rd.loadServicesList()
	f.Services = rd.services

	if err != nil {
		return f, err
	}

	if len(rd.services) == 0 {
		return f, errors.New("no service available for deployment was found")
	}

	if err = rd.checkImage(); err != nil {
		return f, err
	}

	rd.verboseRemappedServices()

	var deploy = &deployment.Deploy{
		ConfigContext: wectx,

		Params:   rd.Params,
		Path:     rd.path,
		Services: rd.services,
	}

	err = deploy.Do(ctx, rd.getTransport())
	f.GroupUID = deploy.GetGroupUID()
	return f, err
}

func (rd *RemoteDeployment) getTransport() deployment.Transport {
	if rd.Experimental {
		return &gogit.Transport{}
	}

	return &git.Transport{}
}

func (rd *RemoteDeployment) checkImage() error {
	if rd.Params.Image == "" || len(rd.services) <= 1 {
		return nil
	}

	var first = rd.services[0]
	var pkg = first.Package()
	var original = pkg.Image
	var diff = []string{}
	var f = fmt.Sprintf("%s is %s (%s)", first.ServiceID, pkg.Image, first.Location)

	for _, s := range rd.services {
		pkg := s.Package()

		if original == "" {
			original = pkg.Image
		}

		if original != pkg.Image {
			diff = append(diff, fmt.Sprintf("%s is %s (%s)", s.ServiceID, pkg.Image, s.Location))
		}
	}

	// found services with duplicated ID "email" on %v and %v
	if len(diff) != 0 {
		diff = append([]string{f}, diff...)
		return fmt.Errorf("refusing to overwrite image for services with different images:\n%v",
			strings.Join(diff, "\n"))
	}

	return nil
}

func (rd *RemoteDeployment) verboseRemappedServices() {
	if !verbose.Enabled {
		return
	}

	for _, s := range rd.remap {
		verbose.Debug("Service " + s + " had service ID mapped or inferred")
	}
}

func (rd *RemoteDeployment) loadServicesList() (err error) {
	if err = rd.loadServicesListFromPath(); err != nil {
		return err
	}

	if err = rd.checkServiceIDs(); err != nil {
		return err
	}

	return rd.remapServicesWithEmptyIDs()
}

func (rd *RemoteDeployment) getServiceIDFromBaseDirOrRandom(s services.ServiceInfo) (newServiceID string) {
	r := regexp.MustCompile("^[0-9a-z]*$")
	serviceID := strings.ToLower(filepath.Base(s.Location))

	if !r.MatchString(serviceID) {
		serviceID = fmt.Sprintf("%s%d", namesgenerator.GetRandomAdjective(), rand.Intn(99))
	}

	verbose.Debug(fmt.Sprintf("service in %v assigned with id %v", s.Location, serviceID))
	return serviceID
}

func (rd *RemoteDeployment) remapServicesWithEmptyIDs() error {
	for k, s := range rd.services {
		if s.ServiceID != "" {
			continue
		}

		rd.services[k].ServiceID = rd.getServiceIDFromBaseDirOrRandom(s)
		rd.remap = append(rd.remap, s.Location)
	}

	return nil
}

func (rd *RemoteDeployment) checkServiceIDs() error {
	if len(rd.services) == 1 {
		return rd.checkServiceParameter()
	}

	if rd.Params.ServiceID != "" {
		return errors.New("service id parameter is not allowed when deploying multiple services")
	}

	return rd.checkEmptyIDOnMultipleDeployment()
}

func (rd *RemoteDeployment) checkEmptyIDOnMultipleDeployment() error {
	var empty []services.ServiceInfo
	for _, s := range rd.services {
		if s.ServiceID == "" {
			empty = append(empty, s)
		}
	}

	if len(empty) == 0 {
		return nil
	}

	fmt.Println(fancy.Info("Multiple services found without wedeploy.json IDs:"))

	for _, e := range empty {
		fmt.Printf("%v %v\n", e.Location, fancy.Tip("using "+filepath.Base(e.Location)))
	}

	fmt.Println("")

	q := fmt.Sprintf("Do you want to %s?",
		color.Format(color.FgMagenta, color.Bold, "continue"))

	switch ok, askErr := fancy.Boolean(q); {
	case askErr != nil:
		return askErr
	case ok:
		fmt.Println("")
		return nil
	}

	return canceled.CancelCommand("deployment canceled")
}

func (rd *RemoteDeployment) checkServiceParameter() error {
	if rd.Params.ServiceID == "" && rd.services[0].ServiceID != "" {
		return nil
	}

	if rd.Params.ServiceID != "" {
		rd.services[0].ServiceID = rd.Params.ServiceID
		rd.remap = append(rd.remap, rd.services[0].Location)
		return nil
	}

	if !isterm.Check() {
		return errors.New("service ID is missing (and a terminal was not found for typing one)")
	}

	var optionServiceID = rd.getServiceIDFromBaseDirOrRandom(rd.services[0])
	fmt.Println(fancy.Question("Your service doesn't have an ID. Type one") + " " +
		fancy.Tip("or use: "+optionServiceID))

	serviceID, err := fancy.Prompt()
	switch {
	case err != nil:
		return err
	case serviceID == "":
		serviceID = optionServiceID
	}

	rd.Params.ServiceID = serviceID
	rd.services[0].ServiceID = serviceID
	rd.remap = append(rd.remap, rd.services[0].Location)
	return nil
}

func (rd *RemoteDeployment) loadServicesListFromPath() (err error) {
	var overview = inspector.ContextOverview{}
	if err = overview.Load(rd.path); err != nil {
		return err
	}

	rd.services = overview.Services

	if len(rd.services) == 0 {
		rd.services = append(rd.services, services.ServiceInfo{
			Location:  rd.path,
			ServiceID: rd.Params.ServiceID,
		})
	}

	for k := range rd.services {
		rd.services[k].ProjectID = rd.Params.ProjectID
	}

	return nil
}
