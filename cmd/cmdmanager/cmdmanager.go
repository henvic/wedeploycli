package cmdmanager

import "github.com/spf13/cobra"

// HideFlag hides a flag
func HideFlag(flag string, rootCmd *cobra.Command) {
	if err := rootCmd.Flags().MarkHidden(flag); err != nil {
		panic(err)
	}
}

// HidePersistentFlag hides a flag
func HidePersistentFlag(flag string, rootCmd *cobra.Command) {
	if err := rootCmd.PersistentFlags().MarkHidden(flag); err != nil {
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
