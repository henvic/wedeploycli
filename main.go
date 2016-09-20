/*
api.cmd

	https://github.com/wedeploy/cli

*/

package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/wedeploy/cli/autocomplete"
	"github.com/wedeploy/cli/cmd"
	"github.com/wedeploy/cli/config"
	"github.com/wedeploy/cli/errorhandling"
	"github.com/wedeploy/cli/update"
)

func main() {
	var panickingFlag = true
	defer panickingListener(&panickingFlag)
	setErrorHandlingCommandName()
	load()

	var isAutoComplete = isCommand("autocomplete")
	var cue chan error

	if !isAutoComplete && !isCommand("build") {
		cue = checkUpdate()
	}

	if ccmd, err := cmd.RootCmd.ExecuteC(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", errorhandling.Handle(err))
		commandErrorConditionalUsage(ccmd, err)
		os.Exit(1)
	}

	if cue != nil {
		updateFeedback(<-cue)
	}

	autocomplete.AutoInstall()
	panickingFlag = false
}

func loadConfig() {
	if err := config.Setup(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", errorhandling.Handle(err))
		os.Exit(1)
	}
}

func load() {
	loadConfig()

	if config.Global.PastVersion != "" {
		update.ApplyFixes(config.Global.PastVersion)
		config.Global.PastVersion = ""

		if err := config.Global.Save(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", errorhandling.Handle(err))
			os.Exit(1)
		}

		// reload config as something might have changed
		loadConfig()
	}
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

func commandErrorConditionalUsage(cmd *cobra.Command, err error) {
	// this tries to print the usage for a given command only when one of the
	// errors below is caused by cobra
	var emsg = err.Error()
	if strings.HasPrefix(emsg, "unknown flag: ") ||
		strings.HasPrefix(emsg, "unknown shorthand flag: ") ||
		strings.HasPrefix(emsg, "invalid argument ") ||
		strings.HasPrefix(emsg, "bad flag syntax: ") ||
		strings.HasPrefix(emsg, "flag needs an argument: ") {
		if ue := cmd.Usage(); ue != nil {
			panic(ue)
		}
	} else if strings.HasPrefix(emsg, "unknown command ") {
		println("Run 'we --help' for usage.")
	}
}

func checkUpdate() chan error {
	var euc = make(chan error, 1)
	go func() {
		euc <- update.NotifierCheck()
	}()
	return euc
}

func updateFeedback(err error) {
	switch err {
	case nil:
		update.Notify()
	default:
		println("Update notification error:", err.Error())
	}
}
