package open

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

func open(m *waitlivemsg.Message, ec chan error) {
	err := browser.OpenURL(fmt.Sprintf("https://%v-%v.%v",
		setupHost.Service(),
		setupHost.Project(),
		setupHost.ServiceDomain()))

	if err != nil {
		m.StopText(fancy.Error("Failed to open service on your browser [1/2]"))
		ec <- err
		return
	}

	m.StopText("Service opened on your browser [2/2]")
	ec <- err
}

func openRun(cmd *cobra.Command, args []string) error {
	var m = waitlivemsg.NewMessage("Opening service on your browser [1/2]")
	var wlm = waitlivemsg.New(nil)
	go wlm.Wait()
	wlm.AddMessage(m)
	var ec = make(chan error, 1)
	go open(m, ec)
	var err = <-ec
	wlm.Stop()
	return err
}
