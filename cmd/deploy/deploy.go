package cmddeploy

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"os/exec"
	"strings"

	"github.com/hashicorp/errwrap"
	"github.com/spf13/cobra"
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
	Short:   "Deploy the current project",
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

// basicAuth creates the basic auth parameter
// extracted from golang/go/src/net/http/client.go
func basicAuth(username, password string) string {
	auth := username + ":" + password
	return base64.StdEncoding.EncodeToString([]byte(auth))
}

func getRepoAuthorization() (string, error) {
	if config.Global.Username == "" {
		return "", errors.New("User is not configured yet")
	}

	return "Basic " + basicAuth(config.Global.Username, config.Global.Password), nil
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

	var initializeErr = deploy.InitalizeRepositoryIfNotExists()

	if initializeErr != nil {
		return initializeErr
	}

	if err := deploy.CheckCurrentBranchIsMaster(); err != nil {
		return err
	}

	stage, errStage := deploy.CheckUncommittedChanges()
	if errStage != nil {
		return errStage
	}

	if len(stage) != 0 {
		if err := askAutoCommit(cmd, stage); err != nil {
			return err
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

func init() {
	setupHost.Init(DeployCmd)
	DeployCmd.Flags().BoolVarP(&force, "force", "f", false, "Force updates")
	DeployCmd.Flags().BoolVar(&tmpAutoCommit, "auto-commit", false, "Stage and auto-commit changes on this and future deployments")
	DeployCmd.Flags().BoolVar(&tmpNoAutoCommit, "no-auto-commit", false, "Do not auto-commit")

	if err := DeployCmd.Flags().MarkHidden("no-auto-commit"); err != nil {
		panic(err)
	}
}
