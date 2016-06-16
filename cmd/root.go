package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/wedeploy/cli/cmd/auth"
	cmdcontainers "github.com/wedeploy/cli/cmd/containers"
	cmdcreate "github.com/wedeploy/cli/cmd/createctx"
	cmddeploy "github.com/wedeploy/cli/cmd/deploy"
	cmdlogs "github.com/wedeploy/cli/cmd/logs"
	cmdprojects "github.com/wedeploy/cli/cmd/projects"
	cmdremote "github.com/wedeploy/cli/cmd/remote"
	cmdrestart "github.com/wedeploy/cli/cmd/restart"
	"github.com/wedeploy/cli/cmd/run"
	cmdupdate "github.com/wedeploy/cli/cmd/update"
	cmdversion "github.com/wedeploy/cli/cmd/version"
	"github.com/wedeploy/cli/config"
	"github.com/wedeploy/cli/defaults"
	"github.com/wedeploy/cli/update"
	"github.com/wedeploy/cli/verbose"
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

var (
	version bool
	local   bool
	remote  string
)

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
	cmdremote.RemoteCmd,
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
		&local,
		"local", false, "Local (for development, remote = local)")

	RootCmd.Flags().StringVar(
		&remote,
		"remote", "", "Remote to use")

	RootCmd.Flags().BoolVar(
		&version,
		"version", false, "Print version information and quit")

	if err := RootCmd.Flags().MarkHidden("version"); err != nil {
		panic(err)
	}

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
		return
	}

	if err := cmd.Help(); err != nil {
		panic(err)
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
