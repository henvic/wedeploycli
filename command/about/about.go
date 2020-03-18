package about

import (
	"github.com/henvic/wedeploycli/command/about/legal"
	"github.com/spf13/cobra"
)

// AboutCmd is used for showing the abouts of all used libraries
var AboutCmd = &cobra.Command{
	Use:   "about",
	Short: "About this software",
}

func init() {
	AboutCmd.AddCommand(legal.LegalCmd)
}
