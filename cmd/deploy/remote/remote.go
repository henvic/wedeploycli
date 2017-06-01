package cmddeployremote

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	"os"

	"github.com/hashicorp/errwrap"
	"github.com/wedeploy/cli/apihelper"
	"github.com/wedeploy/cli/color"
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
	ProjectID  string
	ServiceID  string
	Remote     string
	Quiet      bool
	path       string
	containers containers.ContainerInfoList
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

func listenCleanupOnCancel(deploy *deployment.Deploy) {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigs
		_ = deploy.Cleanup()
	}()
}

// Run does the remote deployment procedures
func (rd *RemoteDeployment) Run() (groupUID string, err error) {
	if config.Context.Scope == usercontext.ContainerScope && rd.ServiceID != "" {
		return "", errors.New("Can not use --container from inside a project container")
	}

	rd.ProjectID, err = rd.getProjectID()

	if err != nil {
		return "", err
	}

	rd.path, err = rd.getPath()

	if err != nil {
		return "", err
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

	if err = rd.loadContainersList(); err != nil {
		return "", err
	}

	if len(rd.containers) == 0 {
		return "", errors.New("no container available for deployment was found")
	}

	if _, err = projectctx.CreateOrUpdate(rd.ProjectID); err != nil {
		return "", err
	}

	rd.printStartDeployment()

	var deploy = &deployment.Deploy{
		Context:           ctx,
		AuthorEmail:       config.Context.Username,
		ProjectID:         rd.ProjectID,
		Path:              rd.path,
		Remote:            config.Context.Remote,
		RepoAuthorization: repoAuthorization,
		GitRemoteAddress:  gitServer,
		Services:          rd.containers.GetIDs(),
	}

	listenCleanupOnCancel(deploy)
	defer signal.Reset(syscall.SIGINT, syscall.SIGTERM)

	err = deploy.Do()

	if rd.Quiet && err == nil {
		fmt.Println("Deployment Group ID:" + deploy.GetGroupUID())
	}

	return deploy.GetGroupUID(), err
}

func (rd *RemoteDeployment) printStartDeployment() {
	p := &bytes.Buffer{}

	p.WriteString(fmt.Sprintf("%v in %v\n",
		color.Format(color.FgBlue, rd.ProjectID),
		color.Format(color.FgBlue, rd.Remote),
	))

	var cl = rd.containers.GetIDs()

	if rd.Quiet && len(cl) > 0 {
		p.WriteString(fmt.Sprintf("%v Services: %v\n",
			color.Format(color.FgGreen, "•"),
			color.Format(color.FgBlue, strings.Join(cl, ", "))))
	}

	fmt.Println(p)
}

func (rd *RemoteDeployment) loadContainersList() error {
	var allProjectContainers, err = getContainersInfoListFromProject(rd.path)

	if err != nil {
		return err
	}

	if config.Context.Scope == usercontext.ProjectScope {
		for _, c := range allProjectContainers {
			rd.containers = append(rd.containers, c)
		}

		return nil
	}

	if config.Context.Scope == usercontext.ContainerScope {
		for _, c := range allProjectContainers {
			if c.Location == rd.path {
				rd.containers = append(rd.containers, c)
			}
		}

		return nil
	}

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
