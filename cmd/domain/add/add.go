package cmddomainadd

import (
	"context"

	"github.com/spf13/cobra"
	"github.com/wedeploy/cli/cmd/internal/we"
	"github.com/wedeploy/cli/cmdargslen"
	"github.com/wedeploy/cli/cmdflagsfromhost"
	"github.com/wedeploy/cli/services"
)

// Cmd for adding a domain
var Cmd = &cobra.Command{
	Use:     "add",
	Aliases: []string{"set"},
	Short:   "Add custom domain to a given service",
	Example: "we domain add example.com",
	PreRunE: preRun,
	RunE:    run,
}

var setupHost = cmdflagsfromhost.SetupHost{
	Pattern:             cmdflagsfromhost.FullHostPattern,
	UseServiceDirectory: true,
	Requires: cmdflagsfromhost.Requires{
		Auth:    true,
		Project: true,
		Service: true,
	},
}

func init() {
	setupHost.Init(Cmd)
}

func preRun(cmd *cobra.Command, args []string) error {
	if err := cmdargslen.Validate(args, 1, 1); err != nil {
		return err
	}

	return setupHost.Process(we.Context())
}

func run(cmd *cobra.Command, args []string) error {
	servicesClient := services.New(we.Context())

	return servicesClient.AddDomain(
		context.Background(),
		setupHost.Project(),
		setupHost.Service(),
		args[0])
}
