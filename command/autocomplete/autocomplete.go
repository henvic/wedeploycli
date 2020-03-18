package autocomplete

import (
	"github.com/henvic/wedeploycli/autocomplete"
	"github.com/spf13/cobra"
)

// AutocompleteCmd is used for reading the version of this tool
var AutocompleteCmd = &cobra.Command{
	Use:    "autocomplete",
	Run:    autocompleteRun,
	Short:  "Provides zsh / bash auto-completion",
	Hidden: true,
}

func autocompleteRun(c *cobra.Command, args []string) {
	autocomplete.Run(args)
}
