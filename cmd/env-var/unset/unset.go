package unset

import (
	"context"

	"github.com/spf13/cobra"
	"github.com/wedeploy/cli/cmd/env-var/internal/commands"
	"github.com/wedeploy/cli/cmd/internal/we"
	"github.com/wedeploy/cli/cmdflagsfromhost"
	"github.com/wedeploy/cli/services"
)

// Cmd for removing an environment variable
var Cmd = &cobra.Command{
	Use:     "rm",
	Aliases: []string{"unset", "del", "delete"},
	Short:   "Remove an environment variable for a given service",
	Example: "  liferay env-var rm foo",
	PreRunE: preRun,
	RunE:    run,
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

func init() {
	setupHost.Init(Cmd)
}

func preRun(cmd *cobra.Command, args []string) error {
	return setupHost.Process(context.Background(), we.Context())
}

func run(cmd *cobra.Command, args []string) error {
	var c = commands.Command{
		SetupHost:      setupHost,
		ServicesClient: services.New(we.Context()),
	}

	ctx := context.Background()

	if len(args) == 0 {
		if err := c.Show(ctx); err != nil {
			return err
		}
	}

	return c.Delete(ctx, args)
}
