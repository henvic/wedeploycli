package cmdcontainers

import (
	"fmt"
	"os"

	"github.com/launchpad-project/cli/cmdcontext"
	ccontainers "github.com/launchpad-project/cli/containers"
	"github.com/spf13/cobra"
)

// ContainersCmd is used for getting containers
var ContainersCmd = &cobra.Command{
	Use:   "containers [project] or containers from inside a project",
	Short: "Container running on Launchpad",
	Run:   containersRun,
}

func errFeedback() {
	fmt.Fprintln(os.Stderr, "Use launchpad containers <project> or launchpad containers from inside a project")
	os.Exit(1)
}

func containersRun(cmd *cobra.Command, args []string) {
	var projectID, err = cmdcontext.GetProjectID(args)

	if err != nil {
		errFeedback()
		return
	}

	ccontainers.List(projectID)
}
