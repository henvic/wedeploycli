package autocomplete

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/wedeploy/cli/deployment"
)

// RootCmd is the entry-point of the program
var RootCmd *cobra.Command

// AutoInstall autocomplete
func AutoInstall() {
	if deployment.IsGitHomeSandbox() {
		return
	}

	autoInstall()
}

// Run autocomplete rules
func Run(args []string) {
	(&autocomplete{}).run(args)
}

type autocomplete struct{}

func (a *autocomplete) run(args []string) {
	var cmdMatches, flagsMatches = a.getMatches(args)
	var already = map[string]bool{}

	for _, a := range args {
		already[a] = true
	}

	var matches = append(cmdMatches, flagsMatches...)

	for _, match := range matches {
		if _, ok := already[match]; !ok {
			fmt.Println(match)
		}
	}
}

func (a *autocomplete) getMatches(args []string) (cmdMatches []string, flagsMatches []string) {
	var current = RootCmd
	var last = current

	for _, name := range args {
		last = a.walkCommand(name, last)

		if last == nil {
			break
		}

		current = last
	}

	if last != nil {
		cmdMatches = a.getVisibleSubCommands(current.Commands())
	}

	flagsMatches = append(a.getVisibleFlags(current.Flags()),
		a.getVisibleFlags(current.InheritedFlags())...,
	)

	return
}

func (a *autocomplete) walkCommand(name string, root *cobra.Command) *cobra.Command {
	for _, c := range root.Commands() {
		if name == c.Name() {
			return c
		}
	}

	return nil
}

func (a *autocomplete) getVisibleFlags(flags *pflag.FlagSet) (flagsMatches []string) {
	flags.VisitAll(func(f *pflag.Flag) {
		if !f.Hidden || f.Name == "help" {
			flagsMatches = append(flagsMatches, "--"+f.Name)
		}
	})
	return
}

func (a *autocomplete) getVisibleSubCommands(cmds []*cobra.Command) (cmdMatches []string) {
	for _, c := range cmds {
		if !c.Hidden {
			cmdMatches = append(cmdMatches, c.Name())
		}
	}
	return
}
