package cmdhooks

import (
	"os"

	"github.com/launchpad-project/cli/config"
	"github.com/launchpad-project/cli/hooks"
	"github.com/spf13/cobra"
)

// BuildCmd builds the current project or container
var BuildCmd = &cobra.Command{
	Use:    "build",
	Short:  "Builds the current project or container",
	PreRun: preRun,
	Run:    buildRun,
}

// DeployCmd deploys the current project or container
var DeployCmd = &cobra.Command{
	Use:    "deploy",
	Short:  "Deploys the current project or container",
	PreRun: preRun,
	Run:    deployRun,
}

func buildRun(cmd *cobra.Command, args []string) {
	err := hooks.Build(config.Context)

	if err != nil {
		println(err.Error())
		os.Exit(1)
	}
}

func deployRun(cmd *cobra.Command, args []string) {
	err := hooks.Deploy(config.Context)

	if err != nil {
		println(err.Error())
		os.Exit(1)
	}
}

func preRun(cmd *cobra.Command, args []string) {
	if config.Context.Scope == "global" {
		println("fatal: not a project")
		os.Exit(1)
	}
}
