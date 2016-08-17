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
	"github.com/wedeploy/cli/cmd"
	"github.com/wedeploy/cli/config"
	"github.com/wedeploy/cli/errorhandling"
	"github.com/wedeploy/cli/update"
)

func main() {
	var panickingFlag = true
	defer panickingListener(&panickingFlag)
	var cue = checkUpdate()

	if err := config.Setup(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", errorhandling.Handle("we", err))
		os.Exit(1)
	}

	if ccmd, err := cmd.RootCmd.ExecuteC(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", errorhandling.Handle(ccmd.Name(), err))
		commandErrorConditionalUsage(ccmd, err)
		os.Exit(1)
	}

	updateFeedback(<-cue)
	panickingFlag = false
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
