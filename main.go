/*
api.cmd

	https://github.com/wedeploy/cli

*/

package main

import (
	"fmt"
	"os"
	"strings"

	"reflect"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/wedeploy/cli/autocomplete"
	"github.com/wedeploy/cli/cmd"
	"github.com/wedeploy/cli/config"
	"github.com/wedeploy/cli/errorhandling"
	"github.com/wedeploy/cli/flagsfromhost"
	"github.com/wedeploy/cli/update"
)

func main() {
	var panickingFlag = true
	defer panickingListener(&panickingFlag)
	setErrorHandlingCommandName()
	(&mainProgram{}).run()
	panickingFlag = false
}

type mainProgram struct {
	cue            chan error
	cmd            *cobra.Command
	cmdErr         error
	cmdFriendlyErr error
}

func (m *mainProgram) run() {
	(&configLoader{}).load()
	m.checkUpdate()
	m.executeCommand()
	m.updateFeedback()
	autocomplete.AutoInstall()
}

func (m *mainProgram) executeCommand() {
	m.cmd, m.cmdErr = cmd.RootCmd.ExecuteC()
	m.cmdFriendlyErr = errorhandling.Handle(m.cmdErr)

	if m.cmdErr != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", m.cmdFriendlyErr)
	}

	if m.cmdErr != nil {
		m.commandErrorConditionalUsage()
		os.Exit(1)
	}
}

func (m *mainProgram) getCommandFlags() []string {
	var flags = []string{}

	m.cmd.Flags().Visit(func(f *pflag.Flag) {
		flags = append(flags, f.Name)
	})

	return flags
}

func (m *mainProgram) getCommandErrorDetails() map[string]string {
	if m.cmdErr == nil {
		return nil
	}

	var extra = map[string]string{
		"error_type": reflect.TypeOf(m.cmdErr).String(),
	}

	if m.cmdErr.Error() != m.cmdFriendlyErr.Error() {
		// currently (?; as of 3rd, Nov 2016) the friendly error message list is
		// static; this might need to be improved later if the list starts accepting
		// values like it was a template system
		extra["friendly_error"] = m.cmdFriendlyErr.Error()
	}

	return extra
}

type configLoader struct {
	reload bool
}

func (cl *configLoader) loadConfig() {
	if err := config.Setup(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", errorhandling.Handle(err))
		os.Exit(1)
	}

	flagsfromhost.InjectRemotes(&(config.Global.Remotes))
}

func (cl *configLoader) checkPastVersion() {
	if config.Global.PastVersion != "" {
		update.ApplyFixes(config.Global.PastVersion)
		config.Global.PastVersion = ""
		cl.reload = true
	}
}

func (cl *configLoader) applyChanges() {
	if !cl.reload {
		return
	}

	if err := config.Global.Save(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", errorhandling.Handle(err))
		os.Exit(1)
	}

	cl.loadConfig()
}

func (cl *configLoader) load() {
	cl.loadConfig()
	cl.checkPastVersion()
	cl.applyChanges()
}

func isCommand(cmd string) bool {
	for _, s := range os.Args {
		if s == cmd {
			return true
		}
	}

	return false
}

func setErrorHandlingCommandName() {
	ccmd, _, err := cmd.RootCmd.Find(os.Args[1:])

	if err != nil {
		return
	}

	errorhandling.CommandName = ccmd.Name()
}

func panickingListener(panicking *bool) {
	if !*panicking {
		return
	}

	errorhandling.Info()
}

func (m *mainProgram) commandErrorConditionalUsage() {
	// this tries to print the usage for a given command only when one of the
	// errors below is caused by cobra
	var emsg = m.cmdErr.Error()
	if strings.HasPrefix(emsg, "unknown flag: ") ||
		strings.HasPrefix(emsg, "unknown shorthand flag: ") ||
		strings.HasPrefix(emsg, "invalid argument ") ||
		strings.HasPrefix(emsg, "bad flag syntax: ") ||
		strings.HasPrefix(emsg, "flag needs an argument: ") {
		if ue := m.cmd.Usage(); ue != nil {
			panic(ue)
		}
	} else if strings.HasPrefix(emsg, "unknown command ") {
		println("Run 'we --help' for usage.")
	}
}

func (m *mainProgram) checkUpdate() {
	if !isCommand("autocomplete") && !isCommand("analytics-report") && !isCommand("build") {
		m.cue = make(chan error, 1)
		go func() {
			m.cue <- update.NotifierCheck()
		}()
	}
}

func (m *mainProgram) updateFeedback() {
	if m.cue == nil {
		return
	}

	var err = <-m.cue
	switch err {
	case nil:
		update.Notify()
	default:
		println("Update notification error:", err.Error())
	}
}
