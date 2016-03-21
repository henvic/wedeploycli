package cmdrestart

import (
	"os"

	"github.com/launchpad-project/cli/cmdcontext"
	"github.com/launchpad-project/cli/containers"
	"github.com/launchpad-project/cli/projects"
	"github.com/spf13/cobra"
)

// RestartCmd is used for getting restart
var RestartCmd = &cobra.Command{
	Use:   "restart [project] [container]",
	Short: "Restart project or container running on Launchpad",
	Run:   restartRun,
	Example: `launchpad restart portal
launchpad restart portal email`,
}

func restartRun(cmd *cobra.Command, args []string) {
	project, container, err := cmdcontext.GetProjectOrContainerID(args)

	if err != nil {
		cmd.Help()
		os.Exit(1)
	}

	switch container {
	case "":
		projects.Restart(project)
	default:
		containers.Restart(project, container)
	}
}
