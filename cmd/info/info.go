package info

import (
	"github.com/launchpad-project/cli/info"
	"github.com/spf13/cobra"
)

// InfoCmd is used for getting info about a given scope
var InfoCmd = &cobra.Command{
	Use:   "info",
	Short: "Displays information about current scope",
	Run:   infoRun,
}

func infoRun(cmd *cobra.Command, args []string) {
	info.Print()
}
