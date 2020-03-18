package console

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
	Pattern: cmdflagsfromhost.RemotePattern,
}

// ConsoleCmd opens the browser console
var ConsoleCmd = &cobra.Command{
	Use:     "console",
	Short:   "Open the console on your browser",
	Args:    cobra.NoArgs,
	PreRunE: consolePreRun,
	RunE:    consoleRun,
}

func init() {
	setupHost.Init(ConsoleCmd)
}

func consolePreRun(cmd *cobra.Command, args []string) error {
	return setupHost.Process(context.Background(), we.Context())
}

func consoleRun(cmd *cobra.Command, args []string) error {
	var link = fmt.Sprintf("https://console.%v", setupHost.InfrastructureDomain())
	err := browser.OpenURL(link)

	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Failed to open %v", link)
		return err
	}

	fmt.Println("Console opened on your browser.")
	return nil
}
