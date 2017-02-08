package cmdenvunset

import (
	"context"

	"github.com/spf13/cobra"
	"github.com/wedeploy/cli/cmdargslen"
	"github.com/wedeploy/cli/cmdflagsfromhost"
	"github.com/wedeploy/cli/containers"
)

// Cmd for removing a domain
var Cmd = &cobra.Command{
	Use:     "rm",
	Short:   "Remove an environment variable for a given container",
	Example: "we env rm foo",
	PreRunE: preRun,
	RunE:    run,
}

var setupHost = cmdflagsfromhost.SetupHost{
	Pattern:               cmdflagsfromhost.FullHostPattern,
	UseProjectDirectory:   true,
	UseContainerDirectory: true,
	Requires: cmdflagsfromhost.Requires{
		Auth:      true,
		Project:   true,
		Container: true,
	},
}

func init() {
	setupHost.Init(Cmd)
}

func preRun(cmd *cobra.Command, args []string) error {
	if err := cmdargslen.Validate(args, 1, 1); err != nil {
		return err
	}

	return setupHost.Process()
}

func run(cmd *cobra.Command, args []string) error {
	return containers.UnsetEnvironmentVariable(
		context.Background(),
		setupHost.Project(),
		setupHost.Container(),
		args[0])
}
