package console

import (
	"context"
	"fmt"

	"github.com/henvic/browser"
	"github.com/spf13/cobra"
	"github.com/wedeploy/cli/cmd/internal/we"
	"github.com/wedeploy/cli/cmdflagsfromhost"
	"github.com/wedeploy/cli/fancy"
	"github.com/wedeploy/cli/waitlivemsg"
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

func open(m *waitlivemsg.Message, ec chan error) {
	err := browser.OpenURL(fmt.Sprintf("https://console.%v", setupHost.InfrastructureDomain()))

	if err != nil {
		m.StopText(fancy.Error("Failed to open console on your browser [1/2]"))
		ec <- err
		return
	}

	m.StopText("Console opened on your browser [2/2]")
	ec <- err
}

func consoleRun(cmd *cobra.Command, args []string) error {
	var m = waitlivemsg.NewMessage("Opening console on your browser [1/2]")
	var wlm = waitlivemsg.New(nil)
	go wlm.Wait()
	wlm.AddMessage(m)
	var ec = make(chan error, 1)
	go open(m, ec)
	var err = <-ec
	wlm.Stop()
	return err
}
