package autocomplete

import (
	"fmt"

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
	(&autocomplete{
		args: args,
	}).run()
}

type autocomplete struct {
	args    []string
	matches []string
}

func (a *autocomplete) run() {
	var current = RootCmd
	var last = current

	for _, name := range a.args {
		last = a.walkCommand(name, last)

		if last == nil {
			break
		}

		current = last
	}

	if last != nil {
		a.getVisibleSubCommands(current.Commands())
	}

	a.getVisibleFlags(current.Flags())
	a.getVisibleFlags(current.InheritedFlags())

	for _, match := range a.matches {
		fmt.Println(match)
	}
}

func (a *autocomplete) walkCommand(name string, root *cobra.Command) *cobra.Command {
	for _, c := range root.Commands() {
		if name == c.Name() {
			return c
		}
	}

	return nil
}

func (a *autocomplete) getVisibleFlags(flags *pflag.FlagSet) {
	flags.VisitAll(func(f *pflag.Flag) {
		if !f.Hidden {
			a.matches = append(a.matches, "--"+f.Name)
		}
	})
}

func (a *autocomplete) getVisibleSubCommands(cmds []*cobra.Command) {
	for _, c := range cmds {
		if !c.Hidden {
			a.matches = append(a.matches, c.Name())
		}
	}
}
