package update

import (
	"github.com/spf13/cobra"
	"github.com/wedeploy/cli/cmd/internal/we"
	"github.com/wedeploy/cli/cmdargslen"
	"github.com/wedeploy/cli/defaults"
	"github.com/wedeploy/cli/update"
)

// UpdateCmd is used for updating this tool
var UpdateCmd = &cobra.Command{
	Use:     "update",
	PreRunE: cmdargslen.ValidateCmd(0, 0),
	RunE:    updateRun,
	Short:   "Update CLI to the latest version",
}

var (
	channel string
)

func updateRun(cmd *cobra.Command, args []string) error {
	var wectx = we.Context()
	if !cmd.Flag("channel").Changed {
		channel = wectx.Config().ReleaseChannel
	}

	return update.Update(wectx.Config(), channel)
}

func init() {
	UpdateCmd.Flags().StringVar(&channel, "channel", defaults.StableReleaseChannel, "Release channel")
}
