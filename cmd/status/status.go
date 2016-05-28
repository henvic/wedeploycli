package cmdstatus

import (
	"fmt"
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
	Example: `we status portal
we status portal email`,
}

func statusRun(cmd *cobra.Command, args []string) {
	var project, container, err = cmdcontext.GetProjectOrContainerID(args)
	var status string

	if err != nil {
		cmd.Help()
		os.Exit(1)
	}

	switch container {
	case "":
		status = projects.GetStatus(project)
		fmt.Println(status + " (" + project + ")")
	default:
		status = containers.GetStatus(project, container)
		fmt.Println(status + " (" + project + " " + container + ")")
	}
}
