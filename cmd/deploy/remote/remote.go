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
	"github.com/wedeploy/cli/containers"
	"github.com/wedeploy/cli/deployment"
	"github.com/wedeploy/cli/inspector"
	"github.com/wedeploy/cli/projectctx"
	"github.com/wedeploy/cli/projects"
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
	containers containers.ContainerInfoList
	changedSID bool
}

func getAuthCredentials() string {
	// hacky way to get the credentials
	// instead of duplicating code, let's use existing one
	// that already does so
	var request = apihelper.URL(context.Background(), "")
	apihelper.Auth(request)
	return request.Headers.Get("Authorization")
}

func getRepoAuthorization() (string, error) {
	if config.Context.Username == "" {
		return "", errors.New("User is not configured yet")
	}

	return getAuthCredentials(), nil
}

func (rd *RemoteDeployment) getPath() (path string, err error) {
	switch config.Context.Scope {
	case usercontext.ProjectScope:
		return config.Context.ProjectRoot, nil
	case usercontext.ContainerScope:
		return config.Context.ContainerRoot, nil
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

	_, err = containers.Read(wd)

	switch {
	case err == containers.ErrContainerNotFound:
		return wd, createContainerPackage(rd.ServiceID, wd)
	case err == nil:
		return wd, nil
	}

	return "", err
}

func createContainerPackage(id, path string) error {
	var c = &containers.ContainerPackage{
		ID:   filepath.Base(path),
		Type: "wedeploy/hosting",
	}

	bin, err := json.MarshalIndent(c, "", "    ")
	if err != nil {
		return err
	}

	return ioutil.WriteFile(filepath.Join(path, "container.json"), bin, 0644)
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

	if err = rd.loadContainersList(); err != nil {
		return "", err
	}

	if len(rd.containers) == 0 {
		return "", errors.New("no service available for deployment was found")
	}

	var repoAuthorization, repoAuthorizationErr = getRepoAuthorization()

	if repoAuthorizationErr != nil {
		return "", repoAuthorizationErr
	}

	var gitServer = fmt.Sprintf("%vgit.%v/%v.git",
		gitSchema,
		config.Context.RemoteAddress,
		rd.ProjectID)

	var ctx = context.Background()

	if _, err = projectctx.CreateOrUpdate(rd.ProjectID); err != nil {
		return "", err
	}

	var deploy = &deployment.Deploy{
		Context:           ctx,
		AuthorEmail:       config.Context.Username,
		ProjectID:         rd.ProjectID,
		ServiceID:         rd.ServiceID,
		ChangedServiceID:  rd.changedSID,
		Path:              rd.path,
		Remote:            config.Context.Remote,
		RemoteAddress:     config.Context.RemoteAddress,
		RepoAuthorization: repoAuthorization,
		GitRemoteAddress:  gitServer,
		Services:          rd.containers.GetIDs(),
		Quiet:             rd.Quiet,
	}

	err = deploy.Do()
	return deploy.GetGroupUID(), err
}

func (rd *RemoteDeployment) loadContainersList() (err error) {
	rd.path, err = rd.getPath()

	if err != nil {
		return err
	}

	allProjectContainers, err := getContainersInfoListFromProject(rd.path)

	if err != nil {
		return err
	}

	switch config.Context.Scope {
	case usercontext.ProjectScope:
		err = rd.loadContainersListForProjectScope(allProjectContainers)
	case usercontext.ContainerScope:
		err = rd.loadContainersListForContainerScope(allProjectContainers)
	default:
		err = rd.loadContainersListForGlobalScope()
	}

	if err == nil {
		err = rd.maybeFilterOrRenameContainer()
	}

	return err
}

func (rd *RemoteDeployment) maybeFilterOrRenameContainer() error {
	if rd.ServiceID == "" {
		return nil
	}

	if config.Context.Scope == usercontext.ProjectScope {
		return rd.maybeFilterOrRenameContainerForProjectScope()
	}

	// for container and global...
	if rd.containers[0].ServiceID != rd.ServiceID {
		rd.containers[0].ServiceID = rd.ServiceID
		rd.changedSID = true
	}

	return nil
}

func (rd *RemoteDeployment) maybeFilterOrRenameContainerForProjectScope() error {
	var s, err = rd.containers.Get(rd.ServiceID)

	if err != nil {
		return err
	}

	rd.containers = containers.ContainerInfoList{
		s,
	}

	rd.path = s.Location
	rd.changedSID = true
	return nil

}

func (rd *RemoteDeployment) loadContainersListForProjectScope(containers containers.ContainerInfoList) (err error) {
	for _, c := range containers {
		rd.containers = append(rd.containers, c)
	}

	return nil
}

func (rd *RemoteDeployment) loadContainersListForContainerScope(containers containers.ContainerInfoList) (err error) {
	for _, c := range containers {
		if c.Location == rd.path {
			rd.containers = append(rd.containers, c)
			break
		}
	}

	return nil
}

func (rd *RemoteDeployment) loadContainersListForGlobalScope() (err error) {
	cp, err := containers.Read(rd.path)

	if err != nil {
		return errwrap.Wrapf("Error reading container after deployment: {{err}}", err)
	}

	rd.containers = append(rd.containers, containers.ContainerInfo{
		Location:  rd.path,
		ServiceID: cp.ID,
	})

	return nil
}

func (rd *RemoteDeployment) printAddress(container string) string {
	var address = rd.ProjectID + "." + config.Global.Remotes[rd.Remote].URL

	if container != "" {
		address = container + "-" + address
	}

	return address
}

// getContainersInfoListFromProject get a list of containers on a given project directory
func getContainersInfoListFromProject(projectPath string) (containers.ContainerInfoList, error) {
	var i = &inspector.ContextOverview{}

	if err := i.Load(projectPath); err != nil {
		return containers.ContainerInfoList{}, errwrap.Wrapf("Can not list containers from project: {{err}}", err)
	}

	var list = containers.ContainerInfoList{}

	for _, c := range i.ProjectContainers {
		list = append(list, c)
	}

	return list, nil
}
