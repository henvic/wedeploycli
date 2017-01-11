package cmdenvunset

import (
	"context"
	"errors"

	"github.com/spf13/cobra"
	"github.com/wedeploy/cli/cmdargslen"
	"github.com/wedeploy/cli/cmdflagsfromhost"
	"github.com/wedeploy/cli/containers"
)

// Cmd for removing a domain
var Cmd = &cobra.Command{
	Use:     "rm",
	Short:   "Remove an environment variable for a given container",
	Example: "we env rm --key FOO",
	PreRunE: preRun,
	RunE:    run,
}

var key string

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
	Cmd.Flags().StringVar(&key, "key", "", "Environment variable name")
}

func preRun(cmd *cobra.Command, args []string) error {
	if err := cmdargslen.Validate(args, 0, 0); err != nil {
		return err
	}

	return setupHost.Process()
}

func run(cmd *cobra.Command, args []string) error {
	if key == "" {
		return errors.New("Environment variable name is required")
	}

	return containers.UnsetEnvironmentVariable(
		context.Background(),
		setupHost.Project(),
		setupHost.Container(),
		key)
}
