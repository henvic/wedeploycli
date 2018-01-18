package env

import (
	"context"
	"fmt"
	"sort"

	"github.com/spf13/cobra"
	cmdenvset "github.com/wedeploy/cli/cmd/env/set"
	cmdenvunset "github.com/wedeploy/cli/cmd/env/unset"
	"github.com/wedeploy/cli/cmd/internal/we"
	"github.com/wedeploy/cli/cmdflagsfromhost"
	"github.com/wedeploy/cli/services"
	"github.com/wedeploy/cli/verbose"
)

// EnvCmd controls the envs for a given project
var EnvCmd = &cobra.Command{
	Use:   "env",
	Short: "Show and configure environment variables for services",
	Long:  `Show and configure environment variables for services. You must restart the service afterwards.`,
	Example: `  we env (to list environment variables)
  we env set foo bar
  we env rm foo`,
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

func preRun(cmd *cobra.Command, args []string) error {
	if _, _, err := cmd.Find(args); err != nil {
		return err
	}

	return setupHost.Process(context.Background(), we.Context())
}

func run(cmd *cobra.Command, args []string) error {
	servicesClient := services.New(we.Context())

	var envs, err = servicesClient.GetEnvironmentVariables(context.Background(),
		setupHost.Project(),
		setupHost.Service())

	if err != nil {
		return err
	}

	if len(envs) == 0 {
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
