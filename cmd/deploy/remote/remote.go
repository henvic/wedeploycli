package cmddeployremote

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"path/filepath"

	"os"

	"github.com/hashicorp/errwrap"
	"github.com/wedeploy/cli/apihelper"
	"github.com/wedeploy/cli/config"
	"github.com/wedeploy/cli/deployment"
	"github.com/wedeploy/cli/inspector"
	"github.com/wedeploy/cli/projectctx"
	"github.com/wedeploy/cli/projects"
	"github.com/wedeploy/cli/services"
	"github.com/wedeploy/cli/usercontext"
)

const gitSchema = "https://"

// RemoteDeployment of services
type RemoteDeployment struct {
	ProjectID  string
	ServiceID  string
	Remote     string
	Quiet      bool
	path       string
	services   services.ServiceInfoList
	changedSID bool
}

func (rd *RemoteDeployment) getPath() (path string, err error) {
	switch config.Context.Scope {
	case usercontext.ProjectScope:
		return config.Context.ProjectRoot, nil
	case usercontext.ServiceScope:
		return config.Context.ServiceRoot, nil
	case usercontext.GlobalScope:
		return rd.getPathForGlobalScope()
	}

	return "", fmt.Errorf("Scope not identified: %v", err)
}

func (rd *RemoteDeployment) getPathForGlobalScope() (path string, err error) {
	wd, err := os.Getwd()

	if err != nil {
		return "", errwrap.Wrapf("Can not get current working directory: {{err}}", err)
	}

	_, err = services.Read(wd)

	switch {
	case err == services.ErrServiceNotFound:
		return wd, createServicePackage(rd.ServiceID, wd)
	case err == nil:
		return wd, nil
	}

	return "", err
}

func createServicePackage(id, path string) error {
	var c = &services.ServicePackage{
		ID:   filepath.Base(path),
		Type: "wedeploy/hosting",
	}

	bin, err := json.MarshalIndent(c, "", "    ")
	if err != nil {
		return err
	}

	return ioutil.WriteFile(filepath.Join(path, "wedeploy.json"), bin, 0644)
}

func (rd *RemoteDeployment) getProjectID() (string, error) {
	var projectID = rd.ProjectID

	if projectID == "" {
		var pp, err = projects.Read(config.Context.ProjectRoot)

		switch {
		case err == nil:
			projectID = pp.ID
		case err != projects.ErrProjectNotFound:
			return "", errwrap.Wrapf("Error trying to read project: {{err}}", err)
		}

		if projectID != "" {
			return projectID, nil
		}
	}

	var p, ep = projects.Create(context.Background(), projects.Project{
		ProjectID: projectID,
	})

	if epf, ok := ep.(*apihelper.APIFault); ok && epf.Has("projectAlreadyExists") {
		return projectID, nil
	}

	return p.ProjectID, ep
}

// Run does the remote deployment procedures
func (rd *RemoteDeployment) Run() (groupUID string, err error) {
	rd.ProjectID, err = rd.getProjectID()

	if err != nil {
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

	var ctx = context.Background()

	if _, err = projectctx.CreateOrUpdate(rd.ProjectID); err != nil {
		return "", err
	}

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

	allProjectServices, err := getServicesInfoListFromProject(rd.path)

	if err != nil {
		return err
	}

	switch config.Context.Scope {
	case usercontext.ProjectScope:
		err = rd.loadServicesListForProjectScope(allProjectServices)
	case usercontext.ServiceScope:
		err = rd.loadServicesListForServiceScope(allProjectServices)
	default:
		err = rd.loadServicesListForGlobalScope()
	}

	if err == nil {
		err = rd.maybeFilterOrRenameService()
	}

	return err
}

func (rd *RemoteDeployment) maybeFilterOrRenameService() error {
	if rd.ServiceID == "" {
		return nil
	}

	if config.Context.Scope == usercontext.ProjectScope {
		return rd.maybeFilterOrRenameServiceForProjectScope()
	}

	// for service and global...
	if rd.services[0].ServiceID != rd.ServiceID {
		rd.services[0].ServiceID = rd.ServiceID
		rd.changedSID = true
	}

	return nil
}

func (rd *RemoteDeployment) maybeFilterOrRenameServiceForProjectScope() error {
	var s, err = rd.services.Get(rd.ServiceID)

	if err != nil {
		return err
	}

	rd.services = services.ServiceInfoList{
		s,
	}

	rd.path = s.Location
	rd.changedSID = true
	return nil

}

func (rd *RemoteDeployment) loadServicesListForProjectScope(services services.ServiceInfoList) (err error) {
	for _, c := range services {
		rd.services = append(rd.services, c)
	}

	return nil
}

func (rd *RemoteDeployment) loadServicesListForServiceScope(services services.ServiceInfoList) (err error) {
	for _, c := range services {
		if c.Location == rd.path {
			rd.services = append(rd.services, c)
			break
		}
	}

	return nil
}

func (rd *RemoteDeployment) loadServicesListForGlobalScope() (err error) {
	cp, err := services.Read(rd.path)

	if err != nil {
		return errwrap.Wrapf("Error reading service after deployment: {{err}}", err)
	}

	rd.services = append(rd.services, services.ServiceInfo{
		Location:  rd.path,
		ServiceID: cp.ID,
	})

	return nil
}

func (rd *RemoteDeployment) printAddress(service string) string {
	var address = rd.ProjectID + "." + config.Global.Remotes[rd.Remote].Service

	if service != "" {
		address = service + "-" + address
	}

	return address
}

// getServicesInfoListFromProject get a list of services on a given project directory
func getServicesInfoListFromProject(projectPath string) (services.ServiceInfoList, error) {
	var i = &inspector.ContextOverview{}

	if err := i.Load(projectPath); err != nil {
		return services.ServiceInfoList{}, errwrap.Wrapf("Can not list services from project: {{err}}", err)
	}

	var list = services.ServiceInfoList{}

	for _, c := range i.ProjectServices {
		list = append(list, c)
	}

	return list, nil
}
