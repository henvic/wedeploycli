package cmdrun

import "github.com/spf13/cobra"

// RunCmd runs the Launchpad structure for development locally
var RunCmd = &cobra.Command{
	Use:     "run",
	Short:   "Run Launchpad infrastructure for development locally",
	Run:     initRun,
	Example: `launchpad init`,
}

func initRun(cmd *cobra.Command, args []string) {
}
