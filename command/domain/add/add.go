package add

import (
	"context"

	"github.com/henvic/wedeploycli/cmdflagsfromhost"
	"github.com/henvic/wedeploycli/command/domain/internal/commands"
	"github.com/henvic/wedeploycli/command/internal/we"
	"github.com/henvic/wedeploycli/services"
	"github.com/spf13/cobra"
)

// Cmd for adding a domain
var Cmd = &cobra.Command{
	Use:     "add",
	Aliases: []string{"set"},
	Short:   "Add custom domain to a given service",
	Example: "  lcp domain add example.com",
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

	return c.Add(context.Background(), args)
}
