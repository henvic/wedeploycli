package update

import (
	"context"

	"github.com/spf13/cobra"
	"github.com/wedeploy/cli/cmd/internal/we"
	"github.com/wedeploy/cli/cmd/update/releasenotes"
	"github.com/wedeploy/cli/defaults"
	"github.com/wedeploy/cli/update"
)

// UpdateCmd is used for updating this tool
var UpdateCmd = &cobra.Command{
	Use:   "update",
	Args:  cobra.NoArgs,
	RunE:  updateRun,
	Short: "Update CLI to the latest version",
}

var (
	channel string
	version string
)

func updateRun(cmd *cobra.Command, args []string) error {
	var wectx = we.Context()
	var conf = wectx.Config()
	var params = conf.GetParams()

	if !cmd.Flag("channel").Changed {
		channel = params.ReleaseChannel
	}

	return update.Update(context.Background(), wectx.Config(), channel, version)
}

func init() {
	UpdateCmd.Flags().StringVar(&channel, "channel", defaults.StableReleaseChannel, "Release channel")
	UpdateCmd.Flags().StringVar(&version, "version", "", "Update to a specific version")
	UpdateCmd.AddCommand(releasenotes.Cmd)

	UpdateCmd.Flag("version").Hidden = true
}
