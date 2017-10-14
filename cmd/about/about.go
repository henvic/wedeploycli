package about

import (
	"github.com/spf13/cobra"
	"github.com/wedeploy/cli/cmd/about/legal"
)

// AboutCmd is used for showing the abouts of all used libraries
var AboutCmd = &cobra.Command{
	Use:   "about",
	Short: "About this software",
}

func init() {
	AboutCmd.AddCommand(legal.LegalCmd)
}
