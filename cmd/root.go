package cmd

import (
	"errors"
	"os"
	"runtime"

	"github.com/spf13/cobra"
	"github.com/wedeploy/cli/autocomplete"
	"github.com/wedeploy/cli/cmd/activities"
	"github.com/wedeploy/cli/cmd/autocomplete"
	"github.com/wedeploy/cli/cmd/cmdmanager"
	"github.com/wedeploy/cli/cmd/console"
	"github.com/wedeploy/cli/cmd/delete"
	"github.com/wedeploy/cli/cmd/deploy"
	"github.com/wedeploy/cli/cmd/diagnostics"
	"github.com/wedeploy/cli/cmd/docs"
	"github.com/wedeploy/cli/cmd/domain"
	"github.com/wedeploy/cli/cmd/env"
	"github.com/wedeploy/cli/cmd/gitcredentialhelper"
	"github.com/wedeploy/cli/cmd/inspect"
	"github.com/wedeploy/cli/cmd/list"
	"github.com/wedeploy/cli/cmd/log"
	"github.com/wedeploy/cli/cmd/login"
	"github.com/wedeploy/cli/cmd/logout"
	"github.com/wedeploy/cli/cmd/remote"
	"github.com/wedeploy/cli/cmd/restart"
	"github.com/wedeploy/cli/cmd/uninstall"
	"github.com/wedeploy/cli/cmd/update"
	"github.com/wedeploy/cli/cmd/version"
	"github.com/wedeploy/cli/cmd/who"
	"github.com/wedeploy/cli/color"
	"github.com/wedeploy/cli/config"
	"github.com/wedeploy/cli/defaults"
	"github.com/wedeploy/cli/envs"
	"github.com/wedeploy/cli/verbose"
	"github.com/wedeploy/cli/verbosereq"
)

// RootCmd is the main command for the CLI
var RootCmd = &cobra.Command{
	Use:               "we",
	Short:             "WeDeploy CLI Tool",
	PersistentPreRunE: persistentPreRun,
	Run:               run,
	SilenceErrors:     true,
	SilenceUsage:      true,
}

var (
	deferred bool
	version  bool
)

var commands = []*cobra.Command{
	cmdactivities.ActivitiesCmd,
	cmddeploy.DeployCmd,
	cmdlist.ListCmd,
	cmdconsole.ConsoleCmd,
	cmddocs.DocsCmd,
	cmdlog.LogCmd,
	cmddomain.DomainCmd,
	cmdenv.EnvCmd,
	cmdrestart.RestartCmd,
	cmddelete.DeleteCmd,
	cmdlogin.LoginCmd,
	cmdlogout.LogoutCmd,
	cmdautocomplete.AutocompleteCmd,
	cmdremote.RemoteCmd,
	cmdcheck.DiagnosticsCmd,
	cmdversion.VersionCmd,
	cmdupdate.UpdateCmd,
	cmdinspect.InspectCmd,
	cmdwho.WhoCmd,
	cmdgitcredentialhelper.GitCredentialHelperCmd,
	cmduninstall.UninstallCmd,
}

// see note on usage of maybeEnableVerboseByEnv
func maybeEnableVerboseByEnv() {
	if unsafe, _ := os.LookupEnv(envs.UnsafeVerbose); unsafe == "true" {
		verbose.Enabled = true
	}
}

func init() {
	cobra.EnableCommandSorting = false
	autocomplete.RootCmd = RootCmd

	RootCmd.PersistentFlags().BoolVarP(
		&verbose.Enabled,
		"verbose",
		"v",
		false,
		"Show more information about an operation")

	// this has to run after defining the --verbose flag above
	maybeEnableVerboseByEnv()

	RootCmd.PersistentFlags().BoolVarP(
		&deferred,
		"defer-verbose",
		"V",
		false,
		"Defer verbose output")

	RootCmd.PersistentFlags().BoolVar(
		&deferred,
		"defer-verbose-output",
		false,
		"Defer verbose output")

	RootCmd.PersistentFlags().BoolVar(
		&verbosereq.Disabled,
		"no-verbose-requests",
		false,
		"Hide verbose output for requests")

	RootCmd.PersistentFlags().BoolVar(
		&color.NoColorFlag,
		"no-color",
		false,
		"Disable color output")

	RootCmd.Flags().BoolVar(
		&version,
		"version", false, "Print version information and quit")

	cmdmanager.HideFlag("version", RootCmd)
	cmdmanager.HidePersistentFlag("defer-verbose", RootCmd)
	cmdmanager.HidePersistentFlag("defer-verbose-output", RootCmd)
	cmdmanager.HidePersistentFlag("no-verbose-requests", RootCmd)
	cmdmanager.HidePersistentFlag("no-color", RootCmd)

	for _, c := range commands {
		RootCmd.AddCommand(c)
	}
}

func checkCompatibility() error {
	// Heuristics to identify Windows Subsystem for Linux
	// and block it from being used from inside a Linux space working directory
	// due to the subsystem incompatibility
	if runtime.GOOS != "windows" {
		return nil
	}

	if dir, _ := os.Getwd(); dir != `C:\WINDOWS\system32` {
		return nil
	}

	return errors.New(`cowardly refusing to use "we.exe" Windows binary on a Linux working directory.
Windows Subsystem for Linux has no support for accessing Linux fs from a Windows application.
Please install the Linux version of this application with:
curl https://cdn.wedeploy.com/cli/latest/wedeploy.sh -sL | bash`)
}

func persistentPreRun(cmd *cobra.Command, args []string) error {
	if deferred {
		verbose.Enabled = true
		verbose.Deferred = true
	}

	if err := checkCompatibility(); err != nil {
		return err
	}

	// load default cloud remote on config context
	return config.SetEndpointContext(defaults.CloudRemote)
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
