package show

import (
	"context"

	"github.com/spf13/cobra"
	"github.com/wedeploy/cli/cmd/env-var/internal/commands"
	"github.com/wedeploy/cli/cmd/internal/we"
	"github.com/wedeploy/cli/cmdflagsfromhost"
	"github.com/wedeploy/cli/services"
)

// Cmd for showing
var Cmd = &cobra.Command{
	Use:     "show",
	Aliases: []string{"list"},
	Short:   "Show your environment variable values for a given service",
	Example: `  liferay env-var show
  liferay env-var show key`,
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

	return c.Show(context.Background(), args...)
}
