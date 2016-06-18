package cmdstop

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/wedeploy/cli/run"
)

// StopCmd stops the WeDeploy local infrastructure for development
var StopCmd = &cobra.Command{
	Use:    "stop",
	Short:  "Stop WeDeploy local infrastructure for development",
	Run:    stopRun,
	Hidden: true,
}

func stopRun(cmd *cobra.Command, args []string) {
	if len(args) != 0 {
		println("This command doesn't take arguments.")
		os.Exit(1)
	}

	run.Stop()
}
