package cmdmanager

import "github.com/spf13/cobra"

// HideVersionFlag hides the global version flag
func HideVersionFlag(rootCmd *cobra.Command) {
	if err := rootCmd.Flags().MarkHidden("version"); err != nil {
		panic(err)
	}
}

// HideNoVerboseRequestsFlag hides the --no-verbose-requests global flag
func HideNoVerboseRequestsFlag(rootCmd *cobra.Command) {
	if err := rootCmd.PersistentFlags().MarkHidden("no-verbose-requests"); err != nil {
		panic(err)
	}
}

// HideNoColorFlag hides the --no-color global flag
func HideNoColorFlag(rootCmd *cobra.Command) {
	if err := rootCmd.PersistentFlags().MarkHidden("no-color"); err != nil {
		panic(err)
	}
}

func filterArguments(args []string) []string {
	if len(args) == 0 {
		return []string{}
	}

	if args[0] == "help" {
		if len(args) == 0 {
			return []string{}
		}

		return args[1:]
	}

	if args[0] != "autocomplete" {
		return args
	}

	return filterAutoCompleteArguments(args)
}

func filterAutoCompleteArguments(args []string) []string {
	if len(args) == 0 {
		return []string{}
	}

	args = args[1:]

	if len(args) == 0 {
		return []string{}
	}

	if args[0] == "--" {
		if len(args) == 0 {
			return []string{}
		}

		return args[1:]
	}

	return args
}
