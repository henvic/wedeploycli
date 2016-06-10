package cmdrestart

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/wedeploy/cli/cmdcontext"
	"github.com/wedeploy/cli/containers"
	"github.com/wedeploy/cli/projects"
)

// RestartCmd is used for getting restart
var RestartCmd = &cobra.Command{
	Use:   "restart [project] [container]",
	Short: "Restart project or container running on WeDeploy",
	Run:   restartRun,
	Example: `we restart portal
we restart portal email`,
}

func restartRun(cmd *cobra.Command, args []string) {
	project, container, err := cmdcontext.GetProjectOrContainerID(args)

	if err != nil {
		if err = cmd.Help(); err != nil {
			panic(err)
		}
		os.Exit(1)
	}

	switch container {
	case "":
		projects.Restart(project)
	default:
		containers.Restart(project, container)
	}
}
