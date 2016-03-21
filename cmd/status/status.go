package cmdstatus

import (
	"os"

	"github.com/launchpad-project/cli/cmdcontext"
	"github.com/launchpad-project/cli/containers"
	"github.com/launchpad-project/cli/projects"
	"github.com/spf13/cobra"
)

// StatusCmd is used for getting status
var StatusCmd = &cobra.Command{
	Use:   "status [project] [container]",
	Short: "Get running status for project or container",
	Run:   statusRun,
	Example: `launchpad status portal
launchpad status portal email`,
}

func statusRun(cmd *cobra.Command, args []string) {
	project, container, err := cmdcontext.GetProjectOrContainerID(args)

	if err != nil {
		cmd.Help()
		os.Exit(1)
	}

	switch container {
	case "":
		projects.GetStatus(project)
	default:
		containers.GetStatus(project, container)
	}
}
