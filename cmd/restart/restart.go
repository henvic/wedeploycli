package restart

import (
	"fmt"
	"os"

	crestart "github.com/launchpad-project/cli/restart"
	"github.com/spf13/cobra"
)

// RestartCmd is used for getting restart
var RestartCmd = &cobra.Command{
	Use:   "restart",
	Short: "Container running on Launchpad",
	Run:   restartRun,
}

func restartRun(cmd *cobra.Command, args []string) {
	switch len(args) {
	case 1:
		crestart.Project(args[0])
	case 2:
		crestart.Container(args[0], args[1])
	default:
		// add proper command not found message
		fmt.Fprintln(os.Stderr, "Use launchpad restart <project> <optional container>")
		os.Exit(1)
	}
}
