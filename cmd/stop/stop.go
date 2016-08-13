package cmdstop

import (
	"errors"

	"github.com/spf13/cobra"
	"github.com/wedeploy/cli/run"
)

// StopCmd stops the WeDeploy local infrastructure for development
var StopCmd = &cobra.Command{
	Use:    "stop",
	Short:  "Stop WeDeploy local infrastructure for development",
	RunE:   stopRun,
	Hidden: true,
}

func stopRun(cmd *cobra.Command, args []string) error {
	if len(args) != 0 {
		return errors.New("This command doesn't take arguments.")
	}

	return run.Stop()
}
