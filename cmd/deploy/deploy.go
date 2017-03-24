package cmddeploy

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
	"github.com/spf13/cobra"
	"github.com/wedeploy/cli/apihelper"
	"github.com/wedeploy/cli/cmdflagsfromhost"
	"github.com/wedeploy/cli/config"
	"github.com/wedeploy/cli/containers"
	"github.com/wedeploy/cli/deployment"
	"github.com/wedeploy/cli/inspector"
	"github.com/wedeploy/cli/projects"
	"github.com/wedeploy/cli/usercontext"
)

const gitSchema = "http://"

// DeployCmd deploys a given project
var DeployCmd = &cobra.Command{
	Use:     "deploy",
	Short:   "Deploy project or container",
	PreRunE: preRun,
	RunE:    run,
}

var setupHost = cmdflagsfromhost.SetupHost{
	Pattern: cmdflagsfromhost.FullHostPattern,
	Requires: cmdflagsfromhost.Requires{
		Auth: true,
	},
}

type deployCmd struct {
	path      string
	projectID string
}

func preRun(cmd *cobra.Command, args []string) error {
	if err := setupHost.Process(); err != nil {
		return err
	}

	if setupHost.Remote() == "" {
		return errors.New(`You can not deploy in the local infrastructure. Use "we run" instead`)
	}

	return checkContextAmbiguity()
}

func checkContextAmbiguity() error {
	var project = setupHost.Project()
	var container = setupHost.Container()

	if config.Context.Scope == usercontext.ProjectScope && project != "" {
		return errors.New("Can not use --project from inside a project")
	}

	if config.Context.Scope == usercontext.GlobalScope && project == "" {
		return errors.New("--project is required when running this command outside a project")
	}

	if config.Context.Scope == usercontext.ContainerScope {
		switch {
		case project != "" && container != "":
			return errors.New("Can not use --project or --container from inside a project container")
		case project != "":
			return errors.New("Can not use --project from inside a project")
		case container != "":
			return errors.New("Can not use --container from inside a project container")
		}
	}

	return nil
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
	if config.Global.Username == "" {
		return "", errors.New("User is not configured yet")
	}

	return getAuthCredentials(), nil
}

func getPath() (path string, err error) {
	if config.Context.Scope != usercontext.GlobalScope {
		switch {
		case setupHost.Project() != "" && setupHost.Container() != "":
			return "", errors.New("--project and --container can not be used inside this context")
		case setupHost.Project() != "":
			return "", errors.New("--project can not be used inside this context")
		case setupHost.Container() != "":
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
		if setupHost.Container() == "" {
			return wd, nil
		}

		return "", errors.New("--container can not be used inside a directory with container.json")
	}

	if err != containers.ErrContainerNotFound {
		return "", err
	}

	return wd, createContainerJSON(setupHost.Container(), wd)
}

func createContainerJSON(id, path string) error {
	var c = &containers.Container{
		ID:   filepath.Base(path),
		Type: "wedeploy/hosting",
	}

	bin, err := json.MarshalIndent(c, "", "    ")
	if err != nil {
		return err
	}

	return ioutil.WriteFile(filepath.Join(path, "container.json"), bin, 0644)
}

func getProjectID() (string, error) {
	var project, err = projects.Read(config.Context.ProjectRoot)
	var projectID = setupHost.Project()

	switch {
	case err == nil:
		projectID = project.ID
	case err != projects.ErrProjectNotFound:
		return "", errwrap.Wrapf("Error trying to read project: {{err}}", err)
	}

	if setupHost.Project() != "" && projectID != setupHost.Project() {
		return "", errwrap.Wrapf("You can not use a different id on --project from inside a project directory", err)
	}

	return projectID, nil
}

func createProjectIfNotFound(ctx context.Context, id string) error {
	var _, err = projects.Get(ctx, id)

	if err == nil {
		return nil
	}

	_, err = projects.Create(ctx, id)

	return err
}

func run(cmd *cobra.Command, args []string) error {
	if setupHost.Remote() == "" {
		return errors.New(`You can not deploy in the local infrastructure. Use "we run" instead`)
	}

	projectID, err := getProjectID()

	if err != nil {
		return err
	}

	var dc = &deployCmd{
		projectID: projectID,
	}

	dc.path, err = getPath()

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
		projectID)

	var ctx = context.Background()

	if err := createProjectIfNotFound(ctx, projectID); err != nil {
		return err
	}

	var deploy = deployment.Deploy{
		Context:           ctx,
		ProjectID:         projectID,
		Path:              dc.path,
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

	return dc.Feedback(deploy.Do())
}

func (dc *deployCmd) Feedback(err error) error {
	if err != nil {
		return err
	}

	fmt.Println("Project " + dc.printAddress(""))

	switch {
	case config.Context.Scope == usercontext.ProjectScope:
		var containersInfoList, err = getContainersInfoListFromProject(dc.path)

		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading containers after deployment: %v", err)
		}

		for _, c := range containersInfoList {
			fmt.Println(dc.printAddress(c.ID))
		}
	case config.Context.Scope == usercontext.ContainerScope:
		var containersInfoList, err = getContainersInfoListFromProject(dc.path)

		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading containers after deployment: %v", err)
		}

		for _, c := range containersInfoList {
			if c.Location == dc.path {
				fmt.Println(dc.printAddress(c.ID))
			}
		}
	default:
		var c, err = containers.Read(dc.path)

		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading container after deployment: %v", err)
		}

		fmt.Println(dc.printAddress(c.ID))
	}

	return nil
}

func (dc *deployCmd) printAddress(container string) string {
	var address = dc.projectID + "." + config.Global.Remotes[setupHost.Remote()].URL

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

func init() {
	setupHost.Init(DeployCmd)
}
