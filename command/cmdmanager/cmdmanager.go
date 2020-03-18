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
