package containers

import (
	"fmt"
	"os"

	ccontainers "github.com/launchpad-project/cli/containers"
	"github.com/spf13/cobra"
)

// ContainersCmd is used for getting containers
var ContainersCmd = &cobra.Command{
	Use:   "containers",
	Short: "Container running on Launchpad",
	Run:   containersRun,
}

func containersRun(cmd *cobra.Command, args []string) {
	if len(args) == 1 {
		ccontainers.List(args[0])
		return
	}

	// add proper command not found message
	fmt.Fprintln(os.Stderr, "Use launchpad containers <project>")
	os.Exit(1)
}
