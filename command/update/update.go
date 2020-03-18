package update

import (
	"context"
	"fmt"
	"os"

	"github.com/wedeploy/cli/fancy"
	"github.com/wedeploy/cli/isterm"
	"github.com/wedeploy/cli/update"

	version "github.com/hashicorp/go-version"
	"github.com/spf13/cobra"
	"github.com/wedeploy/cli/command/canceled"
	"github.com/wedeploy/cli/command/internal/we"
	"github.com/wedeploy/cli/command/update/releasenotes"
	"github.com/wedeploy/cli/defaults"
	"github.com/wedeploy/cli/verbose"
)

// UpdateCmd is used for updating this tool
var UpdateCmd = &cobra.Command{
	Use:   "update",
	Args:  cobra.NoArgs,
	RunE:  updateRun,
	Short: "Update CLI to the latest version",
}

var (
	channel       string
	updateVersion string
)

func updateRun(cmd *cobra.Command, args []string) error {
	var wectx = we.Context()
	var conf = wectx.Config()
	var params = conf.GetParams()

	if !cmd.Flag("channel").Changed {
		channel = params.ReleaseChannel
	}

	if err := checkDowngrade(); err != nil {
		return err
	}

	return update.Update(context.Background(), wectx.Config(), channel, updateVersion)
}

func checkDowngrade() error {
	if updateVersion == "" {
		verbose.Debug("updating to latest available version")
		return nil
	}

	fromV, fromErr := version.NewVersion(defaults.Version)
	toV, toErr := version.NewVersion(updateVersion)

	if fromErr != nil {
		verbose.Debug(fmt.Sprintf("bypassing checking updating: current version error: %v", fromErr))
		return nil
	}

	if toErr != nil {
		verbose.Debug(fmt.Sprintf("checking updating to newer version: %v", toErr))
		fmt.Printf("You are using version %s\n", defaults.Version)
		return confirmDowngrade("New version doesn't follow semantic versioning. Update anyway?")
	}

	if toV.GreaterThan(fromV) {
		return nil
	}

	fmt.Printf("You are using version %s\n", defaults.Version)
	return confirmDowngrade("Downgrade to version " + updateVersion + "?")
}

func confirmDowngrade(question string) error {
	if !isterm.Check() {
		verbose.Debug("skipping checking newer version: no tty")
		return nil
	}

	ok, err := fancy.Boolean(question)

	if err != nil {
		fmt.Fprintf(os.Stderr, "bypassing confirming new version: %v\n", err)
		return nil
	}

	if !ok {
		return canceled.CancelCommand("update canceled")
	}

	return nil
}

func init() {
	UpdateCmd.Flags().StringVar(&channel, "channel", defaults.StableReleaseChannel, "Release channel")
	UpdateCmd.Flags().StringVar(&updateVersion, "version", "", "Update to a specific version")
	UpdateCmd.AddCommand(releasenotes.Cmd)

	UpdateCmd.Flag("version").Hidden = true
}
