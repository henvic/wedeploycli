package cmd

import (
	"github.com/spf13/cobra"
	"github.com/wedeploy/cli/autocomplete"
	"github.com/wedeploy/cli/cmd/autocomplete"
	"github.com/wedeploy/cli/cmd/build"
	"github.com/wedeploy/cli/cmd/cmdmanager"
	"github.com/wedeploy/cli/cmd/createctx"
	"github.com/wedeploy/cli/cmd/inspect"
	"github.com/wedeploy/cli/cmd/link"
	"github.com/wedeploy/cli/cmd/list"
	"github.com/wedeploy/cli/cmd/log"
	"github.com/wedeploy/cli/cmd/login"
	"github.com/wedeploy/cli/cmd/logout"
	"github.com/wedeploy/cli/cmd/remote"
	"github.com/wedeploy/cli/cmd/restart"
	"github.com/wedeploy/cli/cmd/run"
	"github.com/wedeploy/cli/cmd/start"
	"github.com/wedeploy/cli/cmd/stop"
	"github.com/wedeploy/cli/cmd/unlink"
	"github.com/wedeploy/cli/cmd/update"
	"github.com/wedeploy/cli/cmd/version"
	"github.com/wedeploy/cli/cmdflagsfromhost"
	"github.com/wedeploy/cli/color"
	"github.com/wedeploy/cli/config"
	"github.com/wedeploy/cli/defaults"
	"github.com/wedeploy/cli/verbose"
	"github.com/wedeploy/cli/verbosereq"
)

// LocalOnlyCommands for local-only commands
var LocalOnlyCommands = map[string]bool{
	"we link":   true,
	"we unlink": true,
	"we run":    true,
	"we stop":   true,
}

// RootCmd is the main command for the CLI
var RootCmd = &cobra.Command{
	Use:   "we",
	Short: "WeDeploy CLI tool",
	Long: `WeDeploy Command Line Interface
Version ` + defaults.Version + `
Copyright 2016 Liferay, Inc.
http://wedeploy.com`,
	PersistentPreRunE: persistentPreRun,
	Run:               run,
	SilenceErrors:     true,
	SilenceUsage:      true,
}

var (
	version bool
)

var commands = []*cobra.Command{
	cmdautocomplete.AutocompleteCmd,
	cmdlogin.LoginCmd,
	cmdlogout.LogoutCmd,
	cmdcreate.CreateCmd,
	cmdlog.LogCmd,
	cmdlist.ListCmd,
	cmdrestart.RestartCmd,
	cmdbuild.BuildCmd,
	cmdrun.RunCmd,
	cmdstart.StartCmd,
	cmdstop.StopCmd,
	cmdlink.LinkCmd,
	cmdunlink.UnlinkCmd,
	cmdinspect.InspectCmd,
	cmdremote.RemoteCmd,
	cmdupdate.UpdateCmd,
	cmdversion.VersionCmd,
}

func init() {
	autocomplete.RootCmd = RootCmd

	RootCmd.PersistentFlags().BoolVarP(
		&verbose.Enabled,
		"verbose",
		"v",
		false,
		"Verbose output")

	RootCmd.PersistentFlags().BoolVar(
		&verbosereq.Disabled,
		"no-verbose-requests",
		false,
		"Hide verbose output for requests")

	RootCmd.PersistentFlags().BoolVar(
		&color.NoColor,
		"no-color",
		false,
		"Disable color output")

	RootCmd.Flags().BoolVar(
		&version,
		"version", false, "Print version information and quit")

	cmdmanager.HideVersionFlag(RootCmd)
	cmdmanager.HideNoVerboseRequestsFlag(RootCmd)

	for _, c := range commands {
		RootCmd.AddCommand(c)
	}
}

func persistentPreRun(cmd *cobra.Command, args []string) error {
	if config.Global.NoColor {
		color.NoColor = true
	}

	// load default remote (local) on config context
	cmdflagsfromhost.SetLocal()

	return nil
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
