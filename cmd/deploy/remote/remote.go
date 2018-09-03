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
	"github.com/wedeploy/cli/fancy"
	"github.com/wedeploy/cli/inspector"
	"github.com/wedeploy/cli/isterm"
	"github.com/wedeploy/cli/namesgenerator"
	"github.com/wedeploy/cli/services"
	"github.com/wedeploy/cli/verbose"
)

// RemoteDeployment of services
type RemoteDeployment struct {
	ProjectID string
	ServiceID string
	Remote    string

	CopyPackage string

	OnlyBuild    bool
	SkipProgress bool
	Quiet        bool

	path     string
	services services.ServiceInfoList
	remap    []string
	ctx      context.Context
}

// Run does the remote deployment procedures
func (rd *RemoteDeployment) Run(ctx context.Context) (groupUID string, err error) {
	rd.ctx = ctx
	wectx := we.Context()

	if rd.path, err = getWorkingDirectory(); err != nil {
		return "", err
	}

	rd.ProjectID, err = getproject.MaybeID(rd.ProjectID)

	if err != nil {
		return "", err
	}

	if err = rd.loadServicesList(); err != nil {
		return "", err
	}

	if len(rd.services) == 0 {
		return "", errors.New("no service available for deployment was found")
	}

	rd.verboseRemappedServices()

	var deploy = &deployment.Deploy{
		ProjectID:     rd.ProjectID,
		ServiceID:     rd.ServiceID,
		Path:          rd.path,
		ConfigContext: wectx,
		Services:      rd.services,

		CopyPackage: rd.CopyPackage,

		OnlyBuild:    rd.OnlyBuild,
		SkipProgress: rd.SkipProgress,
		Quiet:        rd.Quiet,
	}

	err = deploy.Do(ctx)
	return deploy.GetGroupUID(), err
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

	if rd.ServiceID != "" {
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

	return canceled.CancelCommand("Deployment canceled.")
}

func (rd *RemoteDeployment) checkServiceParameter() error {
	if rd.ServiceID == "" && rd.services[0].ServiceID != "" {
		return nil
	}

	if rd.ServiceID != "" {
		rd.services[0].ServiceID = rd.ServiceID
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

	rd.ServiceID = serviceID
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
			ServiceID: rd.ServiceID,
		})
	}

	return nil
}
