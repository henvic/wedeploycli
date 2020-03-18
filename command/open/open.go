package open

import (
	"context"
	"fmt"
	"os"

	"github.com/henvic/browser"
	"github.com/henvic/wedeploycli/cmdflagsfromhost"
	"github.com/henvic/wedeploycli/command/internal/we"
	"github.com/spf13/cobra"
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
		_, _ = fmt.Fprintf(os.Stderr, "Failed to open %v", link)
		return err
	}

	fmt.Println("Service opened on your browser.")
	return nil
}
