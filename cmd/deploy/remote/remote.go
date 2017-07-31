package cmddeployremote

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"path/filepath"

	"os"

	"github.com/hashicorp/errwrap"
	"github.com/wedeploy/cli/apihelper"
	"github.com/wedeploy/cli/config"
	"github.com/wedeploy/cli/deployment"
	"github.com/wedeploy/cli/fancy"
	"github.com/wedeploy/cli/inspector"
	"github.com/wedeploy/cli/namesgenerator"
	"github.com/wedeploy/cli/projects"
	"github.com/wedeploy/cli/services"
	"golang.org/x/crypto/ssh/terminal"
)

const gitSchema = "https://"

// RemoteDeployment of services
type RemoteDeployment struct {
	ProjectID    string
	ServiceID    string
	Remote       string
	Quiet        bool
	path         string
	services     services.ServiceInfoList
	createTmpPkg bool
	changedSID   bool
	ctx          context.Context
}

func (rd *RemoteDeployment) getPath() (path string, err error) {
	wd, err := os.Getwd()

	if err != nil {
		return "", errwrap.Wrapf("can't get current working directory: {{err}}", err)
	}

	_, err = services.Read(wd)

	switch {
	case err == services.ErrServiceNotFound:
		rd.createTmpPkg = true
		return wd, nil
	case err == nil:
		return wd, nil
	}

	return "", err
}

func createServicePackage(id, path string) error {
	var c = &services.ServicePackage{
		ID:    filepath.Base(path),
		Image: "wedeploy/hosting",
	}

	bin, err := json.MarshalIndent(c, "", "    ")
	if err != nil {
		return err
	}

	return ioutil.WriteFile(filepath.Join(path, "wedeploy.json"), bin, 0644)
}

func (rd *RemoteDeployment) getProjectID() (err error) {
	if rd.ProjectID == "" {
		if !terminal.IsTerminal(int(os.Stdin.Fd())) {
			return errors.New("Project ID is missing")
		}

		fmt.Println(fancy.Question("Choose a project ID") + " " + fancy.Tip("default: random"))
		rd.ProjectID, err = fancy.Prompt()

		if err != nil {
			return err
		}
	}

	if rd.ProjectID != "" {
		_, err := projects.Get(context.Background(), rd.ProjectID)

		if err == nil {
			return nil
		}

		if epf, ok := err.(*apihelper.APIFault); !ok || epf.Status != http.StatusNotFound {
			return err
		}
	}

	var p, ep = projects.Create(context.Background(), projects.Project{
		ProjectID: rd.ProjectID,
	})

	if ep != nil {
		return ep
	}

	rd.ProjectID = p.ProjectID
	return nil
}

// Run does the remote deployment procedures
func (rd *RemoteDeployment) Run(ctx context.Context) (groupUID string, err error) {
	rd.ctx = ctx
	if err = rd.getProjectID(); err != nil {
		return "", err
	}

	if err = rd.loadServicesList(); err != nil {
		return "", err
	}

	if len(rd.services) == 0 {
		return "", errors.New("no service available for deployment was found")
	}

	var gitServer = fmt.Sprintf("%vgit.%v/%v.git",
		gitSchema,
		config.Context.InfrastructureDomain,
		rd.ProjectID)

	var deploy = &deployment.Deploy{
		Context:              ctx,
		AuthorEmail:          config.Context.Username,
		ProjectID:            rd.ProjectID,
		ServiceID:            rd.ServiceID,
		ChangedServiceID:     rd.changedSID,
		Path:                 rd.path,
		Remote:               config.Context.Remote,
		InfrastructureDomain: config.Context.InfrastructureDomain,
		ServiceDomain:        config.Context.ServiceDomain,
		Token:                config.Context.Token,
		GitRemoteAddress:     gitServer,
		Services:             rd.services.GetIDs(),
		Quiet:                rd.Quiet,
	}

	err = deploy.Do()
	return deploy.GetGroupUID(), err
}

func (rd *RemoteDeployment) loadServicesList() (err error) {
	rd.path, err = rd.getPath()

	if err != nil {
		return err
	}

	err = rd.loadServicesListFromPath()

	if err == services.ErrServiceNotFound {
		err = nil
	}

	if err == nil && len(rd.services) <= 1 {
		err = rd.maybeFilterOrRenameService()
	}

	return err
}

func (rd *RemoteDeployment) maybeFilterOrRenameService() error {
	if rd.ServiceID == "" && !rd.createTmpPkg {
		return nil
	}

	if rd.ServiceID == "" && rd.createTmpPkg {
		if !terminal.IsTerminal(int(os.Stdin.Fd())) {
			return errors.New("Service ID is missing")
		}

		fmt.Println(fancy.Question("Your service doesn't have an ID. Type one") + " " + fancy.Tip("default: random"))

		var serviceID, err = fancy.Prompt()

		if err != nil {
			return err
		}

		if serviceID == "" {
			serviceID = namesgenerator.GetRandomAdjective()
			// I should check until I find it available
		}

		rd.ServiceID = serviceID
		rd.services[0].ServiceID = serviceID
		rd.changedSID = true
		return nil
	}

	// for service and global...
	if rd.services[0].ServiceID != rd.ServiceID {
		rd.services[0].ServiceID = rd.ServiceID
		rd.changedSID = true
	}

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

func (rd *RemoteDeployment) printAddress(service string) string {
	var address = rd.ProjectID + "." + config.Global.Remotes[rd.Remote].Service

	if service != "" {
		address = service + "-" + address
	}

	return address
}
