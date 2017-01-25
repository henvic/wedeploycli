package cmddeploy

import (
	"context"
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
	ask   = false
	force = false
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

	var gitServer = fmt.Sprintf("%vgit.%v/%v.git",
		gitSchema,
		config.Context.RemoteAddress,
		project.ID)

	var deploy = deployment.Deploy{
		Context:          context.Background(),
		ProjectID:        project.ID,
		Path:             config.Context.ProjectRoot,
		Remote:           config.Context.Remote,
		GitRemoteAddress: gitServer,
		Force:            force,
	}

	if err := deploy.InitalizeRepositoryIfNotExists(); err != nil {
		return err
	}

	if err := deploy.CheckCurrentBranchIsMaster(); err != nil {
		return err
	}

	stage, errStage := deploy.CheckUncommittedChanges()
	if errStage != nil {
		return errStage
	}

	if len(stage) != 0 {
		if err := maybeAskAutoCommit(cmd, stage); err != nil {
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

func maybeAskAutoCommit(cmd *cobra.Command, stage string) error {
	if ask || (config.Global.AskAutoCommit && !cmd.Flag("ask-on-changes").Changed) {
		return askAutoCommit(cmd, stage)
	}

	return nil
}

func askAutoCommit(cmd *cobra.Command, stage string) error {
	var options = []string{}

	if !config.Global.AskAutoCommit {
		options = append(options, "always")
	}

	options = append(options, "yes")
	options = append(options, "no")
	var action, err = prompt.Prompt("Stage and commit changes? [" + strings.Join(options, "/") + "]")
	action = strings.ToLower(action)

	if err != nil {
		return errwrap.Wrapf("Can not auto-commit: %v", err)
	}

	return askAutoCommitSelect(action)
}

func askAutoCommitSelect(action string) error {
	switch {
	case action == "a" || action == "always":
		config.Global.AskAutoCommit = true
		if err := config.Global.Save(); err != nil {
			return errwrap.Wrapf("Can not save auto commit config: {{err}}", err)
		}

		return nil
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
	DeployCmd.Flags().BoolVar(&ask, "ask-on-changes", false, "Ask before staging and auto-committing changes on deployment")
}
