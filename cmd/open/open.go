package open

import (
	"context"
	"fmt"
	"os"

	"github.com/henvic/browser"
	"github.com/spf13/cobra"
	"github.com/wedeploy/cli/cmd/internal/we"
	"github.com/wedeploy/cli/cmdflagsfromhost"
)

var setupHost = cmdflagsfromhost.SetupHost{
	Pattern: cmdflagsfromhost.FullHostPattern,

	Requires: cmdflagsfromhost.Requires{
		Project: true,
		Service: true,
	},

	PromptMissingService: true,
}

// OpenCmd opens the browser on a service page
var OpenCmd = &cobra.Command{
	Use:     "open",
	Short:   "Open service on your browser",
	Args:    cobra.NoArgs,
	PreRunE: openPreRun,
	RunE:    openRun,
}

func init() {
	setupHost.Init(OpenCmd)
}

func openPreRun(cmd *cobra.Command, args []string) error {
	return setupHost.Process(context.Background(), we.Context())
}

func openRun(cmd *cobra.Command, args []string) error {
	var link = fmt.Sprintf("https://%s", setupHost.Host())
	err := browser.OpenURL(link)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to open %v", link)
		return err
	}

	fmt.Println("Service opened on your browser.")
	return nil
}
