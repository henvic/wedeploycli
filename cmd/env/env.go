package cmdenv

import (
	"context"
	"fmt"
	"sort"

	"github.com/spf13/cobra"
	cmdenvset "github.com/wedeploy/cli/cmd/env/set"
	cmdenvunset "github.com/wedeploy/cli/cmd/env/unset"
	"github.com/wedeploy/cli/cmdflagsfromhost"
	"github.com/wedeploy/cli/containers"
	"github.com/wedeploy/cli/verbose"
)

// EnvCmd controls the envs for a given project
var EnvCmd = &cobra.Command{
	Use:   "env",
	Short: "Environment variables for containers",
	Long: `Environment variables for containers

Use "we env" to list environment variables on the infrastructure.

Changing these values does not change container.json hard coded values.
You must restart containers for changed values to apply.`,
	Example: `  we env
  we env set foo bar
  we env rm foo`,
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

func preRun(cmd *cobra.Command, args []string) error {
	if _, _, err := cmd.Find(args); err != nil {
		return err
	}

	return setupHost.Process()
}

func run(cmd *cobra.Command, args []string) error {
	var container, err = containers.Get(context.Background(), setupHost.Project(), setupHost.Container())

	if err != nil {
		return err
	}

	if container.Env == nil {
		verbose.Debug("No environment variable found.")
		return nil
	}

	var keys = make([]string, 0)

	for k := range container.Env {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	for _, k := range keys {
		fmt.Printf("%v=%v\n", k, container.Env[k])
	}

	return nil
}

func init() {
	setupHost.Init(EnvCmd)
	EnvCmd.AddCommand(cmdenvset.Cmd)
	EnvCmd.AddCommand(cmdenvunset.Cmd)
}
