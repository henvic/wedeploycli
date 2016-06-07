package cmdrun

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/wedeploy/cli/run"
)

// RunCmd runs the WeDeploy structure for development locally
var RunCmd = &cobra.Command{
	Use:   "run",
	Short: "Run WeDeploy infrastructure for development locally",
	Run:   runRun,
}

var (
	detach   bool
	dryRun   bool
	viewMode bool
)

func runRun(cmd *cobra.Command, args []string) {
	if len(args) != 0 {
		println("This command doesn't take arguments.")
		os.Exit(1)
	}

	run.Run(run.Flags{
		Detach:   detach,
		DryRun:   dryRun,
		ViewMode: viewMode,
	})
}

func init() {
	RunCmd.Flags().BoolVarP(&detach, "detach", "d", false,
		"Run in background")

	RunCmd.Flags().BoolVar(&dryRun, "dry-run", false,
		"Obtain a summary of what docker command is invoked")

	RunCmd.Flags().BoolVar(&viewMode, "view-mode", false,
		"View only mode (no controls)")
}
