package cmdupdate

import (
	"github.com/launchpad-project/cli/update"
	"github.com/spf13/cobra"
)

// UpdateCmd is used for updating this tool
var UpdateCmd = &cobra.Command{
	Use:   "update",
	Run:   updateRun,
	Short: "Updates this tool to the latest version",
}

var (
	channel string
)

func updateRun(cmd *cobra.Command, args []string) {
	update.Update(channel)
}

func init() {
	UpdateCmd.Flags().StringVar(&channel, "channel", "stable", "Release channel")
}
