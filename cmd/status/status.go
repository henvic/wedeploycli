package cmdstatus

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/wedeploy/cli/cmdcontext"
	"github.com/wedeploy/cli/containers"
	"github.com/wedeploy/cli/projects"
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
		if err = cmd.Help(); err != nil {
			panic(err)
		}
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
