package cmdmanager

import (
	"os"

	"github.com/spf13/cobra"
)

// RootCmd is the entry-point of the program
var RootCmd *cobra.Command

// ListNoRemoteFlags hides the globals non used --remote
var ListNoRemoteFlags = map[string]bool{
	"link":    true,
	"unlink":  true,
	"run":     true,
	"stop":    true,
	"remote":  true,
	"update":  true,
	"version": true,
}

// HideVersionFlag hides the global version flag
func HideVersionFlag() {
	if err := RootCmd.Flags().MarkHidden("version"); err != nil {
		panic(err)
	}
}

// HideNoVerboseRequestsFlag hides the --no-verbose-requests global flag
func HideNoVerboseRequestsFlag() {
	if err := RootCmd.PersistentFlags().MarkHidden("no-verbose-requests"); err != nil {
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

// HideUnusedGlobalRemoteFlags hides the --remote flag for commands that doesn't use them
func HideUnusedGlobalRemoteFlags() {
	var args = os.Args

	args = filterArguments(args[1:])

	if len(args) == 0 {
		return
	}

	if _, h := ListNoRemoteFlags[args[0]]; !h {
		return
	}

	if err := RootCmd.PersistentFlags().MarkHidden("remote"); err != nil {
		panic(err)
	}
}
