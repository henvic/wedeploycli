package cmdenvset

import (
	"context"

	"strings"

	"github.com/spf13/cobra"
	"github.com/wedeploy/cli/cmdargslen"
	"github.com/wedeploy/cli/cmdflagsfromhost"
	"github.com/wedeploy/cli/containers"
)

// Cmd for removing a domain
var Cmd = &cobra.Command{
	Use:     "set",
	Aliases: []string{"add"},
	Short:   "Set an environment variable for a given container",
	Example: `  we env set key value
  we env set key=value`,
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
	if err := cmdargslen.Validate(args, 1, 2); err != nil {
		return err
	}

	return setupHost.Process()
}

func run(cmd *cobra.Command, args []string) error {
	var key = args[0]
	var value string

	if len(args) == 2 {
		value = args[1]
	}

	if value == "" && len(args) == 1 {
		var v = strings.SplitN(key, "=", 2)
		key = v[0]
		value = v[1]
	}

	// don't check if value for environment variable is empty
	// as it is an acceptable value, no matter how useless it might be

	return containers.SetEnvironmentVariable(
		context.Background(),
		setupHost.Project(),
		setupHost.Container(),
		key,
		value)
}
