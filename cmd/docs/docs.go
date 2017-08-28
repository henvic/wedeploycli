package cmddocs

import (
	"github.com/henvic/browser"
	"github.com/spf13/cobra"
	"github.com/wedeploy/cli/cmdargslen"
	"github.com/wedeploy/cli/fancy"
	"github.com/wedeploy/cli/waitlivemsg"
)

// DocsCmd opens the docs on the browser
var DocsCmd = &cobra.Command{
	Use:     "docs",
	Short:   "Open docs on your browser",
	PreRunE: cmdargslen.ValidateCmd(0, 0),
	RunE:    docsRun,
}

func open(m *waitlivemsg.Message, ec chan error) {
	err := browser.OpenURL("https://wedeploy.com/docs/")

	if err != nil {
		m.StopText(fancy.Error("Failed to open docs on your browser [1/2]"))
		ec <- err
		return
	}

	m.StopText(fancy.Success("Docs opened on your browser [2/2]"))
	ec <- err
}

func docsRun(cmd *cobra.Command, args []string) error {
	var m = waitlivemsg.NewMessage("Opening docs on your browser [1/2]")
	var wlm = waitlivemsg.New(nil)
	go wlm.Wait()
	wlm.AddMessage(m)
	var ec = make(chan error, 1)
	go open(m, ec)
	var err = <-ec
	wlm.Stop()
	return err
}