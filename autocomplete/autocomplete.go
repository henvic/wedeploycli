package autocomplete

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/wedeploy/cli/config"
)

// RootCmd is the entry-point of the program
var RootCmd *cobra.Command

// AutoInstall autocomplete
func AutoInstall() {
	if !config.Global.NoAutocomplete {
		autoInstall()
	}
}

// Run autocomplete rules
func Run(args []string) {
	(&autocomplete{}).run(args)
}

type autocomplete struct{}

func hasCommands(matches []string) bool {
	for _, c := range matches {
		// not a flag or shorthand to flag
		if !strings.HasPrefix(c, "-") {
			return true
		}
	}

	return false
}

func getNextCommand(matches []string) (string, bool) {
	for _, c := range matches {
		// not a flag or shorthand to flag
		if !strings.HasPrefix(c, "-") {
			return c, true
		}
	}

	return "", false
}

func filterLastCommand(recArgs []string) (args []string) {
	var last = -1
	for index, arg := range recArgs {
		if !strings.HasPrefix(arg, "-") {
			last = index
		}
	}

	for index, arg := range recArgs {
		if index != last {
			recArgs = append(recArgs, arg)
		}
	}

	return
}

func getPossibleCommands(args []string) (possibleCommands []string) {
	for _, c := range args {
		// not a flag or shorthand to flag
		if !strings.HasPrefix(c, "-") {
			possibleCommands = append(possibleCommands, c)
		}
	}

	return possibleCommands
}

func (a *autocomplete) tryGetMatches(args []string) (cmdMatches []string, flagsMatches []string) {
	cmdMatches, flagsMatches = a.getMatches(args)

	if len(args) == 0 || hasCommands(cmdMatches) {
		return
	}

	if len(getPossibleCommands(args)) != 1 {
		return
	}

	next, ok := getNextCommand(args)

	if !ok {
		return
	}

	// get list for parent and filter all commands, except if it starts with 'next'
	cmdMatches, flagsMatches = a.getMatches(filterLastCommand(args))

	var filteredCmds = []string{}

	for _, c := range cmdMatches {
		if strings.HasPrefix(c, next) && c != next {
			filteredCmds = append(filteredCmds, c)
		}
	}

	cmdMatches = filteredCmds
	return
}

func (a *autocomplete) run(args []string) {
	var cmdMatches, flagsMatches = a.tryGetMatches(args)
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
		if !f.Hidden {
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
