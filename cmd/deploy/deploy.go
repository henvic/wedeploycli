package cmddeploy

import (
	"context"
	"errors"
	"fmt"
	"os/exec"

	"github.com/hashicorp/errwrap"
	"github.com/spf13/cobra"
	"github.com/wedeploy/cli/apihelper"
	"github.com/wedeploy/cli/cmdflagsfromhost"
	"github.com/wedeploy/cli/config"
	"github.com/wedeploy/cli/deployment"
	"github.com/wedeploy/cli/projects"
	"github.com/wedeploy/cli/usercontext"
	"github.com/wedeploy/cli/verbose"
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

func preRun(cmd *cobra.Command, args []string) error {
	if err := setupHost.Process(); err != nil {
		return err
	}

	if setupHost.Remote() == "" {
		return errors.New(`You can not deploy in the local infrastructure. Use "we dev" instead`)
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

func run(cmd *cobra.Command, args []string) error {
	if setupHost.Remote() == "" {
		return errors.New(`You can not deploy in the local infrastructure. Use "we dev" instead`)
	}

	if config.Context.Scope == usercontext.GlobalScope {
		return errors.New("You are not inside a project")
	}

	var project, projectErr = projects.Read(config.Context.ProjectRoot)

	if projectErr != nil {
		return errwrap.Wrapf("Error trying to read project: {{err}}", projectErr)
	}

	var repoAuthorization, repoAuthorizationErr = getRepoAuthorization()

	if repoAuthorizationErr != nil {
		return repoAuthorizationErr
	}

	var gitServer = fmt.Sprintf("%vgit.%v/%v.git",
		gitSchema,
		config.Context.RemoteAddress,
		project.ID)

	var deploy = deployment.Deploy{
		Context:           context.Background(),
		ProjectID:         project.ID,
		Path:              config.Context.ProjectRoot,
		Remote:            config.Context.Remote,
		RepoAuthorization: repoAuthorization,
		GitRemoteAddress:  gitServer,
	}

	if err := deploy.Cleanup(); err != nil {
		return errwrap.Wrapf("Can not clean up directory for deployment: {{err}}", err)
	}

	defer func() {
		if err := deploy.Cleanup(); err != nil {
			verbose.Debug(
				errwrap.Wrapf("Error trying to clean up directory after deployment: {{err}}", err))
		}
	}()

	if err := deploy.InitializeRepository(); err != nil {
		return err
	}

	if _, err := deploy.Commit(); err != nil {
		return err
	}

	if err := deploy.AddRemote(); err != nil {
		return err
	}

	if err := deploy.Push(); err != nil {
		if _, ok := err.(*exec.ExitError); ok {
			return errwrap.Wrapf("Can not deploy (push failure)", err)
		}

		return errwrap.Wrapf("Unexpected push failure: can not deploy ({{err}})", err)
	}

	fmt.Println("Deploying project " + project.ID)
	return nil
}

func init() {
	setupHost.Init(DeployCmd)
}
