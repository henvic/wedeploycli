package cmddeploy

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/hashicorp/errwrap"
	"github.com/spf13/cobra"
	"github.com/wedeploy/cli/apihelper"
	"github.com/wedeploy/cli/cmdflagsfromhost"
	"github.com/wedeploy/cli/config"
	"github.com/wedeploy/cli/deployment"
	"github.com/wedeploy/cli/projects"
	"github.com/wedeploy/cli/prompt"
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

var (
	tmpAutoCommit   = false
	tmpNoAutoCommit = false
	force           = false
)

var setupHost = cmdflagsfromhost.SetupHost{
	Pattern: cmdflagsfromhost.RemotePattern,
	Requires: cmdflagsfromhost.Requires{
		Auth: true,
	},
}

func preRun(cmd *cobra.Command, args []string) error {
	return setupHost.Process()
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

func maybeInitializeRepositoryIfNotExists(d *deployment.Deploy) (inited bool, err error) {
	switch _, err := os.Stat(filepath.Join(d.Path, ".git")); {
	case os.IsNotExist(err):
		if errAskInitAndCommit := askInit(); errAskInitAndCommit != nil {
			return false, errAskInitAndCommit
		}

		err = d.InitializeRepository()

		if err != nil {
			return false, err
		}

		return true, nil
	case err != nil:
		return false, errwrap.Wrapf("Unexpected error when trying to find .git: {{err}}", err)
	default:
		return false, nil
	}
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
		Force:             force,
	}

	var inited, initErr = maybeInitializeRepositoryIfNotExists(&deploy)

	if initErr != nil {
		return initErr
	}

	if err := deploy.CheckCurrentBranchIsMaster(); err != nil {
		return err
	}

	stage, errStage := deploy.CheckUncommittedChanges()
	if errStage != nil {
		return errStage
	}

	if len(stage) != 0 {
		if !inited {
			if err := askAutoCommit(cmd, stage); err != nil {
				return err
			}
		}

		if _, err := deploy.Commit(); err != nil {
			return err
		}
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

	return nil
}

func askAutoCommit(cmd *cobra.Command, stage string) error {
	var autoCommit bool
	var flagChanged = cmd.Flags().Changed("no-auto-commit")

	switch flagChanged {
	case true:
		autoCommit = !tmpNoAutoCommit
	default:
		flagChanged = cmd.Flags().Changed("auto-commit")
		autoCommit = tmpAutoCommit
	}

	if flagChanged && autoCommit != config.Global.AskAutoCommit {
		config.Global.AskAutoCommit = autoCommit
		if err := config.Global.Save(); err != nil {
			return errwrap.Wrapf("Can not save auto commit config: {{err}}", err)
		}
	}

	if !config.Global.AskAutoCommit {
		return nil
	}

	return askAutoCommitSelect()
}

func askAutoCommitSelect() error {
	var options = []string{}

	options = append(options, "yes")
	options = append(options, "no")
	var action, err = prompt.Prompt("Stage and commit changes? [" + strings.Join(options, "/") + "]")
	action = strings.ToLower(action)

	if err != nil {
		return errwrap.Wrapf("Can not auto-commit: %v", err)
	}

	switch {
	case action == "y" || action == "yes":
		return nil
	case action == "n" || action == "no":
		return errors.New("Aborting deployment")
	default:
		return errors.New("Invalid option")
	}
}

func askInit() error {
	var options = []string{}

	options = append(options, "yes")
	options = append(options, "no")
	fmt.Println("Your project does not have a git repository yet.")
	var action, err = prompt.Prompt("Initialize a repo, stage files, and commit changes? [" + strings.Join(options, "/") + "]")
	action = strings.ToLower(action)

	if err != nil {
		return errwrap.Wrapf("Can not initialize git and auto-commit: %v", err)
	}

	switch {
	case action == "y" || action == "yes":
		return nil
	case action == "n" || action == "no":
		return errors.New("Aborting deployment")
	default:
		return errors.New("Invalid option")
	}
}

func init() {
	setupHost.Init(DeployCmd)
	DeployCmd.Flags().BoolVarP(&force, "force", "f", false, "Force updates")
	DeployCmd.Flags().BoolVar(&tmpAutoCommit, "auto-commit", false, "Stage and auto-commit changes on this and future deployments")
	DeployCmd.Flags().BoolVar(&tmpNoAutoCommit, "no-auto-commit", false, "Do not auto-commit")

	if err := DeployCmd.Flags().MarkHidden("no-auto-commit"); err != nil {
		panic(err)
	}
}
