package cmdconsole

import (
	"fmt"

	"github.com/henvic/browser"
	"github.com/spf13/cobra"
	"github.com/wedeploy/cli/cmd/internal/we"
	"github.com/wedeploy/cli/cmdargslen"
	"github.com/wedeploy/cli/cmdflagsfromhost"
	"github.com/wedeploy/cli/fancy"
	"github.com/wedeploy/cli/waitlivemsg"
)

var setupHost = cmdflagsfromhost.SetupHost{
	Pattern: cmdflagsfromhost.FullHostPattern,
}

// ConsoleCmd opens the browser console
var ConsoleCmd = &cobra.Command{
	Use:     "console",
	Short:   "Open the console on your browser",
	PreRunE: consolePreRun,
	RunE:    consoleRun,
}

func init() {
	setupHost.Init(ConsoleCmd)
}

func consolePreRun(cmd *cobra.Command, args []string) error {
	if err := cmdargslen.Validate(args, 0, 0); err != nil {
		return err
	}

	return setupHost.Process(we.Context())
}

func open(m *waitlivemsg.Message, ec chan error) {
	err := browser.OpenURL(fmt.Sprintf("https://console.%v", setupHost.InfrastructureDomain()))

	if err != nil {
		m.StopText(fancy.Error("Failed to open console on your browser [1/2]"))
		ec <- err
		return
	}

	m.StopText(fancy.Success("Console opened on your browser [2/2]"))
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
