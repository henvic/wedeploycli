package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/launchpad-project/cli/cmd/auth"
	cmdcontainers "github.com/launchpad-project/cli/cmd/containers"
	cmdcreate "github.com/launchpad-project/cli/cmd/createctx"
	cmddeploy "github.com/launchpad-project/cli/cmd/deploy"
	cmdlogs "github.com/launchpad-project/cli/cmd/logs"
	cmdprojects "github.com/launchpad-project/cli/cmd/projects"
	cmdrestart "github.com/launchpad-project/cli/cmd/restart"
	"github.com/launchpad-project/cli/cmd/run"
	cmdupdate "github.com/launchpad-project/cli/cmd/update"
	cmdversion "github.com/launchpad-project/cli/cmd/version"
	"github.com/launchpad-project/cli/config"
	"github.com/launchpad-project/cli/defaults"
	"github.com/launchpad-project/cli/update"
	"github.com/launchpad-project/cli/verbose"
	"github.com/spf13/cobra"
)

// WhitelistCmdsNoAuthentication for cmds that doesn't require authentication
var WhitelistCmdsNoAuthentication = map[string]bool{
	"login":   true,
	"logout":  true,
	"build":   true,
	"deploy":  true,
	"update":  true,
	"version": true,
}

// RootCmd is the main command for the CLI
var RootCmd = &cobra.Command{
	Use:   "we",
	Short: "WeDeploy CLI tool",
	Long: `WeDeploy Command Line Interface
Version ` + defaults.Version + `
Copyright 2016 Liferay, Inc.
http://liferay.io`,
	PersistentPreRun: persistentPreRun,
	Run:              run,
}

var version bool

// Execute is the Entry-point for the CLI
func Execute() {
	var cue = checkUpdate()

	if err := RootCmd.Execute(); err != nil {
		os.Exit(-1)
	}

	updateFeedback(<-cue)
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

var commands = []*cobra.Command{
	cmdauth.LoginCmd,
	cmdauth.LogoutCmd,
	cmdcreate.CreateCmd,
	cmdlogs.LogsCmd,
	cmdprojects.ProjectsCmd,
	cmdcontainers.ContainersCmd,
	cmdrestart.RestartCmd,
	cmdrun.RunCmd,
	cmddeploy.DeployCmd,
	cmdupdate.UpdateCmd,
	cmdversion.VersionCmd,
}

func init() {
	config.Setup()

	RootCmd.PersistentFlags().BoolVarP(
		&verbose.Enabled,
		"verbose",
		"v",
		false,
		"Verbose output")

	RootCmd.PersistentFlags().BoolVar(
		&color.NoColor,
		"no-color",
		false,
		"Disable color output")

	RootCmd.Flags().BoolVar(
		&version,
		"version", false, "Print version information and quit")
	RootCmd.Flags().MarkHidden("version")

	if config.Global.NoColor {
		color.NoColor = true
	}

	for _, c := range commands {
		RootCmd.AddCommand(c)
	}
}

func persistentPreRun(cmd *cobra.Command, args []string) {
	verifyAuth(cmd.CommandPath())
}

func run(cmd *cobra.Command, args []string) {
	if version {
		cmdversion.VersionCmd.Run(cmd, args)
	} else {
		cmd.Help()
	}
}

func isCmdWhitelistNoAuth(commandPath string) bool {
	var parts = strings.SplitAfterN(commandPath, " ", 2)

	if len(parts) < 2 {
		return true
	}

	for key := range WhitelistCmdsNoAuthentication {
		if key == parts[1] {
			return true
		}
	}

	return false
}

func verifyAuth(commandPath string) {
	if isCmdWhitelistNoAuth(commandPath) {
		return
	}

	var g = config.Global

	if g.Endpoint != "" && g.Username != "" && g.Password != "" {
		return
	}

	pleaseLoginFeedback()
}

func pleaseLoginFeedback() {
	fmt.Fprintf(os.Stderr, "Please run \"we login\" first.\n")
	os.Exit(1)
}
