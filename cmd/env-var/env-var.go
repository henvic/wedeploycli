package env

import (
	"context"

	"github.com/spf13/cobra"
	"github.com/wedeploy/cli/cmd/env-var/internal/commands"
	cmdenvset "github.com/wedeploy/cli/cmd/env-var/set"
	cmdenvshow "github.com/wedeploy/cli/cmd/env-var/show"
	cmdenvunset "github.com/wedeploy/cli/cmd/env-var/unset"
	"github.com/wedeploy/cli/cmd/internal/we"
	"github.com/wedeploy/cli/cmdflagsfromhost"
	"github.com/wedeploy/cli/fancy"
	"github.com/wedeploy/cli/services"
)

type interativeEnvCmd struct {
	ctx context.Context
	c   commands.Command
}

var ie = &interativeEnvCmd{}

// EnvCmd controls the envs for a given project
var EnvCmd = &cobra.Command{
	Use:     "env-var",
	Aliases: []string{"env"},
	Short:   "Show and configure environment variables for services",
	Long:    `Show and configure environment variables for services. You must restart the service afterwards.`,
	Example: `  we env-var (to list and change your environment variables values)
  we env-var set foo bar
  we env-var rm foo`,
	Args:    cobra.NoArgs,
	PreRunE: ie.preRun,
	RunE:    ie.run,
}

var setupHost = cmdflagsfromhost.SetupHost{
	Pattern: cmdflagsfromhost.FullHostPattern,

	Requires: cmdflagsfromhost.Requires{
		Auth:    true,
		Project: true,
		Service: true,
	},

	PromptMissingService: true,
}

func (ie *interativeEnvCmd) preRun(cmd *cobra.Command, args []string) error {
	ie.ctx = context.Background()

	if _, _, err := cmd.Find(args); err != nil {
		return err
	}

	return setupHost.Process(context.Background(), we.Context())
}

func (ie *interativeEnvCmd) run(cmd *cobra.Command, args []string) error {
	ie.c = commands.Command{
		SetupHost:      setupHost,
		ServicesClient: services.New(we.Context()),
	}

	if err := ie.c.Show(ie.ctx); err != nil {
		return err
	}

	var operations = fancy.Options{}

	const (
		addOption   = "a"
		unsetOption = "d"
	)

	operations.Add(addOption, "Add environment variable")
	operations.Add(unsetOption, "Delete environment variable")

	op, err := operations.Ask("Select one of the operations for \"" + setupHost.Host() + "\":")

	if err != nil {
		return err
	}

	switch op {
	case unsetOption:
		err = ie.unsetCmd()
	default:
		err = ie.addCmd()
	}

	if err != nil {
		return err
	}

	return ie.c.Show(ie.ctx)
}

func (ie *interativeEnvCmd) addCmd() error {
	return ie.c.Add(ie.ctx, []string{})
}

func (ie *interativeEnvCmd) unsetCmd() error {
	return ie.c.Delete(ie.ctx, []string{})
}

func init() {
	setupHost.Init(EnvCmd)
	EnvCmd.AddCommand(cmdenvshow.Cmd)
	EnvCmd.AddCommand(cmdenvset.Cmd)
	EnvCmd.AddCommand(cmdenvunset.Cmd)
}
