package cmdconsole

import (
	"fmt"

	"github.com/pkg/browser"
	"github.com/spf13/cobra"
	"github.com/wedeploy/cli/cmdargslen"
	"github.com/wedeploy/cli/cmdflagsfromhost"
	"github.com/wedeploy/cli/fancy"
	"github.com/wedeploy/cli/waitlivemsg"
)

var setupHost = cmdflagsfromhost.SetupHost{
	Pattern: cmdflagsfromhost.FullHostPattern,
	Requires: cmdflagsfromhost.Requires{
		Auth: true,
	},
	UseProjectDirectoryForService: true,
}

// ConsoleCmd runs the WeDeploy structure for development locally
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

	return setupHost.Process()
}

func consoleRun(cmd *cobra.Command, args []string) error {
	var m = waitlivemsg.NewMessage("Opening console on your browser [1/2]")
	var wlm = waitlivemsg.New(nil)
	go wlm.Wait()
	// defer wlm.Stop()

	wlm.AddMessage(m)
	var ec = make(chan error, 1)

	go func() {
		if err := browser.OpenURL(fmt.Sprintf("https://console.%v", setupHost.InfrastructureDomain())); err != nil {
			m.StopText(fancy.Error("Failed to open console on your browser [1/2]"))
			ec <- err
			return
		}

		m.StopText(fancy.Success("Console opened on your browser [2/2]"))
		var err error
		ec <- err
	}()

	var err error
	err = <-ec
	wlm.Stop()
	return err
}