package docs

import (
	"fmt"

	"github.com/henvic/browser"
	"github.com/spf13/cobra"
	"github.com/wedeploy/cli/links"
)

// DocsCmd opens the docs on the browser
var DocsCmd = &cobra.Command{
	Use:   "docs",
	Short: "Open docs on your browser\n\t\t",
	Args:  cobra.NoArgs,
	RunE:  docsRun,
}

func docsRun(cmd *cobra.Command, args []string) error {
	err := browser.OpenURL(links.Docs)

	if err != nil {
		return err
	}

	fmt.Println("Docs opened on your browser.")
	return nil
}
