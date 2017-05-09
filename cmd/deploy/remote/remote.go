package cmddeployremote

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os/signal"
	"path/filepath"
	"syscall"

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

const gitSchema = "http://"

// RemoteDeployment of services
type RemoteDeployment struct {
	ProjectID          string
	ServiceID          string
	Remote             string
	path               string
	containersInfoList containers.ContainerInfoList
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
	if config.Context.Scope != usercontext.GlobalScope {
		switch {
		case rd.ProjectID != "" && rd.ServiceID != "":
			return "", errors.New("--project and --container can not be used inside this context")
		case rd.ServiceID != "":
			return "", errors.New("--container can not be used inside this context")
		}
	}

	if config.Context.Scope == usercontext.ProjectScope {
		return config.Context.ProjectRoot, nil
	}

	if config.Context.Scope == usercontext.ContainerScope {
		return config.Context.ContainerRoot, nil
	}

	wd, err := os.Getwd()

	if err != nil {
		return "", errwrap.Wrapf("Can not get current working directory: {{err}}", err)
	}

	_, err = containers.Read(wd)

	if err == nil {
		if rd.ServiceID == "" {
			return wd, nil
		}

		return "", errors.New("--container can not be used inside a directory with container.json")
	}

	if err != containers.ErrContainerNotFound {
		return "", err
	}

	return wd, createContainerPackage(rd.ServiceID, wd)
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
	var project, err = projects.Read(config.Context.ProjectRoot)
	var projectID = rd.ProjectID

	switch {
	case err == nil:
		projectID = project.ID
	case err != projects.ErrProjectNotFound:
		return "", errwrap.Wrapf("Error trying to read project: {{err}}", err)
	}

	if rd.ProjectID != "" && projectID != rd.ProjectID {
		return "", errwrap.Wrapf("You can not use a different id on --project from inside a project directory", err)
	}

	return projectID, nil
}

// Run does the remote deployment procedures
func (rd *RemoteDeployment) Run() (err error) {
	if config.Context.Scope == usercontext.ContainerScope && rd.ServiceID != "" {
		return errors.New("Can not use --container from inside a project container")
	}

	rd.ProjectID, err = rd.getProjectID()

	if err != nil {
		return err
	}

	rd.path, err = rd.getPath()

	if err != nil {
		return err
	}

	var repoAuthorization, repoAuthorizationErr = getRepoAuthorization()

	if repoAuthorizationErr != nil {
		return repoAuthorizationErr
	}

	var gitServer = fmt.Sprintf("%vgit.%v/%v.git",
		gitSchema,
		config.Context.RemoteAddress,
		rd.ProjectID)

	var ctx = context.Background()

	if rd.containersInfoList, err = getContainersInfoListFromProject(rd.path); err != nil {
		return err
	}

	_, err = projectctx.CreateOrUpdate(rd.ProjectID)

	if err != nil {
		return err
	}

	var deploy = deployment.Deploy{
		Context:           ctx,
		ProjectID:         rd.ProjectID,
		Path:              rd.path,
		Remote:            config.Context.Remote,
		RepoAuthorization: repoAuthorization,
		GitRemoteAddress:  gitServer,
	}

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigs
		_ = deploy.Cleanup()
	}()

	return rd.Feedback(deploy.Do())
}

// Feedback of a remote deployment
func (rd *RemoteDeployment) Feedback(err error) error {
	if err != nil {
		return err
	}

	fmt.Println("Project " + rd.printAddress(""))

	switch {
	case config.Context.Scope == usercontext.ProjectScope:
		for _, c := range rd.containersInfoList {
			fmt.Println(rd.printAddress(c.ServiceID))
		}
	case config.Context.Scope == usercontext.ContainerScope:
		for _, c := range rd.containersInfoList {
			if c.Location == rd.path {
				fmt.Println(rd.printAddress(c.ServiceID))
			}
		}
	default:
		var cp, err = containers.Read(rd.path)

		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading container after deployment: %v", err)
		}

		fmt.Println(rd.printAddress(cp.ID))
	}

	return nil
}

func (rd *RemoteDeployment) printAddress(container string) string {
	var address = rd.ProjectID + "." + config.Global.Remotes[rd.Remote].URL

	if container != "" {
		address = container + "." + address
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
