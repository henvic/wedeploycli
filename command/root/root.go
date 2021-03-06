package root

import (
	"errors"
	"os"
	"runtime"

	"github.com/henvic/wedeploycli/autocomplete"
	"github.com/henvic/wedeploycli/color"
	"github.com/henvic/wedeploycli/command/canceled"
	"github.com/henvic/wedeploycli/command/cmdmanager"
	"github.com/henvic/wedeploycli/command/internal/template"
	"github.com/henvic/wedeploycli/command/internal/we"
	cmdversion "github.com/henvic/wedeploycli/command/version"
	"github.com/henvic/wedeploycli/defaults"
	"github.com/henvic/wedeploycli/envs"
	"github.com/henvic/wedeploycli/isterm"
	"github.com/henvic/wedeploycli/verbose"
	"github.com/henvic/wedeploycli/verbosereq"
	"github.com/spf13/cobra"
)

// Cmd is the main command for the CLI
var Cmd = &cobra.Command{
	Use:               "lcp",
	Short:             "Liferay Cloud Platform CLI Tool",
	Args:              cobra.NoArgs,
	PersistentPreRunE: persistentPreRun,
	RunE:              runE,
	SilenceErrors:     true,
	SilenceUsage:      true,
}

var (
	longHelp bool

	deferred bool
	version  bool
)

// see note on usage of maybeEnableVerboseByEnv
func maybeEnableVerboseByEnv() {
	if unsafe, _ := os.LookupEnv(envs.UnsafeVerbose); unsafe == "true" {
		verbose.Enabled = true
	}
}

func init() {
	template.Configure(Cmd)
	cobra.EnableCommandSorting = false
	cobra.MousetrapHelpText = defaults.MousetrapHelpText
	autocomplete.RootCmd = Cmd

	Cmd.PersistentFlags().BoolVarP(
		&verbose.Enabled,
		"verbose",
		"v",
		false,
		"Show more information about an operation")

	// this has to run after defining the --verbose flag above
	maybeEnableVerboseByEnv()

	Cmd.PersistentFlags().BoolVarP(
		&longHelp,
		"long-help",
		"H",
		false,
		"Show help message (hidden commands and flags included)")

	Cmd.PersistentFlags().BoolVarP(
		&deferred,
		"defer-verbose",
		"V",
		false,
		"Defer verbose output")

	Cmd.PersistentFlags().BoolVar(
		&verbosereq.Disabled,
		"no-verbose-requests",
		false,
		"Hide verbose output for requests")

	Cmd.PersistentFlags().BoolVar(
		&color.NoColorFlag,
		"no-color",
		false,
		"Disable color output")

	Cmd.PersistentFlags().BoolVar(
		&isterm.NoTTY,
		"no-tty",
		false,
		"Run without terminal support")

	Cmd.Flags().BoolVar(
		&version,
		"version", false, "Print version information and quit")

	cmdmanager.HideFlag("version", Cmd)
	hideHelpFlag()
	cmdmanager.HidePersistentFlag("long-help", Cmd)
	cmdmanager.HidePersistentFlag("defer-verbose", Cmd)
	cmdmanager.HidePersistentFlag("no-verbose-requests", Cmd)
	cmdmanager.HidePersistentFlag("no-color", Cmd)
	cmdmanager.HidePersistentFlag("no-tty", Cmd)

	for _, c := range commands {
		Cmd.AddCommand(c)
	}
}

func hideHelpFlag() {
	// hide the --help flag on all commands, but top-level
	Cmd.PersistentFlags().BoolP("help", "h", false, "Show help message")
	_ = Cmd.PersistentFlags().MarkHidden("help")
	Cmd.Flags().BoolP("help", "h", false, "Show help message")
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

	return errors.New(`cowardly refusing to use "lcp.exe" Windows binary on a Linux working directory.
Windows Subsystem for Linux has no support for accessing Linux filesystem from a Windows application.
Please install the Linux version of this application with:
curl https://cdn.liferay.cloud/cli/latest/lcp.sh -fsSL | bash`)
}

func checkLongHelp(cmd *cobra.Command) error {
	if cmd.Flag("long-help").Value.String() != "true" {
		return nil
	}

	if err := cmd.Flag("help").Value.Set("true"); err != nil {
		panic(err)
	}

	if err := cmd.Help(); err != nil {
		return err
	}

	return canceled.Skip()
}

func persistentPreRun(cmd *cobra.Command, args []string) error {
	// it's problematic to use checkLongHelp here, see https://github.com/wedeploy/cli/issues/418
	if err := checkLongHelp(cmd); err != nil {
		return err
	}

	if deferred {
		verbose.Enabled = true
		verbose.Deferred = true
	}

	if err := checkCompatibility(); err != nil {
		return err
	}

	// load default cloud remote on config context
	var wectx = we.Context()
	return wectx.SetEndpoint(defaults.CloudRemote)
}

func runE(cmd *cobra.Command, args []string) error {
	if version {
		cmdversion.Print()
		return nil
	}

	return cmd.Help()
}
