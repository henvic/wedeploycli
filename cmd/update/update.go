package cmdupdate

import (
	"github.com/spf13/cobra"
	"github.com/wedeploy/cli/update"
)

// UpdateCmd is used for updating this tool
var UpdateCmd = &cobra.Command{
	Use:   "update",
	RunE:  updateRun,
	Short: "Updates this tool to the latest version",
}

var (
	channel string
)

func updateRun(cmd *cobra.Command, args []string) error {
	return update.Update(channel)
}

func init() {
	UpdateCmd.Flags().StringVar(&channel, "channel", "stable", "Release channel")
}
