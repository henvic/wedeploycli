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
	var envs, err = containers.GetEnvironmentVariables(context.Background(),
		setupHost.Project(),
		setupHost.Container())

	if err != nil {
		return err
	}

	if envs == nil || len(envs) == 0 {
		verbose.Debug("No environment variable found.")
		return nil
	}

	sort.Slice(envs, func(i, j int) bool {
		return envs[i].Name < envs[j].Name
	})

	for _, v := range envs {
		fmt.Printf("%v=%v\n", v.Name, v.Value)
	}

	return nil
}

func init() {
	setupHost.Init(EnvCmd)
	EnvCmd.AddCommand(cmdenvset.Cmd)
	EnvCmd.AddCommand(cmdenvunset.Cmd)
}
