package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/wedeploy/cli/autocomplete"
	"github.com/wedeploy/cli/cmd/autocomplete"
	"github.com/wedeploy/cli/cmd/build"
	"github.com/wedeploy/cli/cmd/cmdmanager"
	"github.com/wedeploy/cli/cmd/deploy"
	"github.com/wedeploy/cli/cmd/diagnostics"
	"github.com/wedeploy/cli/cmd/domain"
	"github.com/wedeploy/cli/cmd/env"
	"github.com/wedeploy/cli/cmd/generate"
	"github.com/wedeploy/cli/cmd/inspect"
	"github.com/wedeploy/cli/cmd/list"
	"github.com/wedeploy/cli/cmd/log"
	"github.com/wedeploy/cli/cmd/login"
	"github.com/wedeploy/cli/cmd/logout"
	"github.com/wedeploy/cli/cmd/metrics"
	"github.com/wedeploy/cli/cmd/remote"
	"github.com/wedeploy/cli/cmd/remove"
	"github.com/wedeploy/cli/cmd/removed"
	"github.com/wedeploy/cli/cmd/restart"
	"github.com/wedeploy/cli/cmd/start"
	"github.com/wedeploy/cli/cmd/update"
	"github.com/wedeploy/cli/cmd/version"
	"github.com/wedeploy/cli/cmd/who"
	"github.com/wedeploy/cli/color"
	"github.com/wedeploy/cli/config"
	"github.com/wedeploy/cli/defaults"
	"github.com/wedeploy/cli/verbose"
	"github.com/wedeploy/cli/verbosereq"
)

// RootCmd is the main command for the CLI
var RootCmd = &cobra.Command{
	Use:               "we",
	Short:             "WeDeploy CLI tool",
	Long:              "WeDeploy CLI tool wedeploy.com",
	PersistentPreRunE: persistentPreRun,
	Run:               run,
	SilenceErrors:     true,
	SilenceUsage:      true,
}

var (
	version bool
)

var commands = []*cobra.Command{
	cmddeploy.DeployCmd,
	cmdlist.ListCmd,
	cmdlog.LogCmd,
	cmddomain.DomainCmd,
	cmdenv.EnvCmd,
	cmdrestart.RestartCmd,
	cmdremove.RemoveCmd,
	cmdlogin.LoginCmd,
	cmdlogout.LogoutCmd,
	cmdautocomplete.AutocompleteCmd,
	cmdgenerate.GenerateCmd,
	cmdbuild.BuildCmd,
	cmdstart.StartCmd,
	cmdremote.RemoteCmd,
	cmdmetrics.MetricsCmd,
	cmdupdate.UpdateCmd,
	cmdversion.VersionCmd,
	cmdinspect.InspectCmd,
	cmddiagnostics.DiagnosticsCmd,
	cmdwho.WhoCmd,
}

// see note on usage of maybeEnableVerboseByEnv
func maybeEnableVerboseByEnv() {
	if unsafe, _ := os.LookupEnv("WEDEPLOY_UNSAFE_VERBOSE"); unsafe == "true" {
		verbose.Enabled = true
	}
}

func init() {
	cobra.EnableCommandSorting = false
	autocomplete.RootCmd = RootCmd
	commands = append(commands, cmdremoved.List...)

	RootCmd.PersistentFlags().BoolVarP(
		&verbose.Enabled,
		"verbose",
		"v",
		false,
		"Verbose output")

	// this has to run after defining the --verbose flag above
	maybeEnableVerboseByEnv()

	RootCmd.PersistentFlags().BoolVar(
		&verbose.Defered,
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
	cmdmanager.HidePersistentFlag("defer-verbose-output", RootCmd)
	cmdmanager.HidePersistentFlag("no-verbose-requests", RootCmd)
	cmdmanager.HidePersistentFlag("no-color", RootCmd)

	for _, c := range commands {
		RootCmd.AddCommand(c)
	}
}

func persistentPreRun(cmd *cobra.Command, args []string) error {
	if verbose.Defered {
		fmt.Fprintf(os.Stderr, color.Format(color.FgBlue, "Defering verbose output (might lose messages on failures)\n"))
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
