package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/wedeploy/cli/cmd/auth"
	"github.com/wedeploy/cli/cmd/containers"
	"github.com/wedeploy/cli/cmd/createctx"
	"github.com/wedeploy/cli/cmd/link"
	"github.com/wedeploy/cli/cmd/logs"
	"github.com/wedeploy/cli/cmd/projects"
	"github.com/wedeploy/cli/cmd/remote"
	"github.com/wedeploy/cli/cmd/restart"
	"github.com/wedeploy/cli/cmd/run"
	"github.com/wedeploy/cli/cmd/unlink"
	"github.com/wedeploy/cli/cmd/update"
	"github.com/wedeploy/cli/cmd/version"
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

// ListNoRemoteFlags hides the globals non used --remote and --local flags
var ListNoRemoteFlags = map[string]bool{
	"link":    true,
	"unlink":  true,
	"run":     true,
	"remote":  true,
	"update":  true,
	"version": true,
}

// LocalOnlyCommands sets the --local flag automatically for given commands
var LocalOnlyCommands = map[string]bool{
	"link":   true,
	"unlink": true,
	"run":    true,
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
	cmdlink.LinkCmd,
	cmdunlink.UnlinkCmd,
	cmdremote.RemoteCmd,
	cmdupdate.UpdateCmd,
	cmdversion.VersionCmd,
}

func hideVersionFlag() {
	if err := RootCmd.Flags().MarkHidden("version"); err != nil {
		panic(err)
	}
}

func hideUnusedGlobalRemoteFlags() {
	var args = os.Args

	if len(args) < 2 {
		return
	}

	_, h := ListNoRemoteFlags[args[1]]

	if !h {
		return
	}

	if err := RootCmd.PersistentFlags().MarkHidden("local"); err != nil {
		panic(err)
	}

	if err := RootCmd.PersistentFlags().MarkHidden("remote"); err != nil {
		panic(err)
	}
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

	RootCmd.PersistentFlags().BoolVar(
		&local,
		"local", false, "Local (for development, remote = local)")

	RootCmd.PersistentFlags().StringVar(
		&remote,
		"remote", "", "Remote to use")

	RootCmd.Flags().BoolVar(
		&version,
		"version", false, "Print version information and quit")

	hideVersionFlag()
	hideUnusedGlobalRemoteFlags()

	if config.Global.NoColor {
		color.NoColor = true
	}

	for _, c := range commands {
		RootCmd.AddCommand(c)
	}
}

func setLocal() {
	config.Global.Endpoint = "http://localhost:8080/"
	config.Global.Token = "1"
}

func setRemote() {
	var r, ok = config.Global.Remotes.Get(remote)

	if !ok {
		fmt.Fprintf(os.Stderr, "Remove %v is not configured.\n", remote)
	}

	config.Global.Endpoint = r.URL
}

func persistentPreRun(cmd *cobra.Command, args []string) {
	cmdSetLocalFlag()
	verifyCmdReqAuth(cmd.CommandPath())

	switch {
	case local:
		setLocal()
	case remote != "":
		setRemote()
	}
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

func cmdSetLocalFlag() {
	var args = os.Args

	if len(args) < 2 {
		return
	}

	_, h := LocalOnlyCommands[args[1]]

	if h {
		setLocal()
	}
}

func verifyCmdReqAuth(commandPath string) {
	if isCmdWhitelistNoAuth(commandPath) {
		return
	}

	var g = config.Global

	var hasAuth = (g.Token != "") || (g.Username != "" && g.Password != "")

	if g.Endpoint != "" && hasAuth {
		return
	}

	pleaseLoginFeedback()
}

func pleaseLoginFeedback() {
	fmt.Fprintf(os.Stderr, "Please run \"we login\" first.\n")
	os.Exit(1)
}
