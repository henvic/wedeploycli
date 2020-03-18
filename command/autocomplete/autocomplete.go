package autocomplete

import (
	"github.com/spf13/cobra"
	"github.com/wedeploy/cli/autocomplete"
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
