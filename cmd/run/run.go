package cmdrun

import (
	"errors"

	"github.com/spf13/cobra"
	"github.com/wedeploy/cli/run"
)

// RunCmd runs the WeDeploy infrastructure for development locally
var RunCmd = &cobra.Command{
	Use:   "run",
	Short: "Run WeDeploy infrastructure for development locally",
	RunE:  runRun,
}

var (
	detach   bool
	dryRun   bool
	viewMode bool
	noUpdate bool
)

func runRun(cmd *cobra.Command, args []string) error {
	if len(args) != 0 {
		return errors.New("Invalid number of arguments.")
	}

	return run.Run(run.Flags{
		Detach:   detach,
		DryRun:   dryRun,
		ViewMode: viewMode,
		NoUpdate: noUpdate,
	})
}

func init() {
	RunCmd.Flags().BoolVarP(&detach, "detach", "d", false,
		"Run in background")

	RunCmd.Flags().BoolVar(&dryRun, "dry-run", false,
		"Obtain a summary of what docker command is invoked")

	RunCmd.Flags().BoolVar(&viewMode, "view-mode", false,
		"View only mode (no controls)")

	RunCmd.Flags().BoolVar(&noUpdate, "no-update", false,
		"Don't try to update the docker image")
}
