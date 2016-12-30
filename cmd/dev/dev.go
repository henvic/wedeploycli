package cmddev

import (
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/wedeploy/cli/cmd/dev/unlink"
	"github.com/wedeploy/cli/cmdflagsfromhost"
	"github.com/wedeploy/cli/run"
	"github.com/wedeploy/cli/verbose"
)

type commandRunner interface {
	Init()
	PreRun(cmd *cobra.Command, args []string) error
	Run(cmd *cobra.Command, args []string) error
}

var (
	setupHost  cmdflagsfromhost.SetupHost
	runFlags   = run.Flags{}
	quiet      bool
	infra      bool
	skipInfra  bool
	noInfraTmp bool
	image      string
	cmdRunner  commandRunner
)

// DevCmd runs the WeDeploy local infrastructure
// and / or a project or container
var DevCmd = &cobra.Command{
	Use:     "dev",
	Short:   "Run development environment for a project or container",
	PreRunE: devPreRun,
	RunE:    devRun,
	Example: `  we dev
  we dev stop
  we dev --project chat
  we dev --infra to startup only the infrastructure
  we dev --no-infra to shutdown the local infrastructure`,
}

func maybeShutdown() (err error) {
	if infra {
		return nil
	}

	return run.Stop()
}

func maybeStartInfrastructure() error {
	if skipInfra {
		verbose.Debug("Skipping setting up infra-structure.")
		return nil
	}

	var defaultImage = run.WeDeployImage
	if image != "" {
		run.WeDeployImage = image
		fmt.Fprintf(os.Stderr,
			"INFO: Using experimental image %v instead of default image %v",
			image,
			defaultImage)
	}

	return run.Run(runFlags)
}

func devPreRun(cmd *cobra.Command, args []string) error {
	if noInfraTmp {
		infra = false
	}

	if cmdRunner != nil {
		return cmdRunner.PreRun(cmd, args)
	}

	return nil
}

func devRun(cmd *cobra.Command, args []string) (err error) {
	if runFlags.Debug && (!cmd.Flags().Changed("infra") || !infra || skipInfra) {
		return errors.New("Incompatible use: --debug requires --infra")
	}

	if len(args) != 0 {
		return errors.New("Invalid number of arguments.")
	}

	if !infra {
		return maybeShutdown()
	}

	if err = maybeStartInfrastructure(); err != nil {
		return err
	}

	if runFlags.DryRun || cmdRunner == nil {
		return nil
	}

	return cmdRunner.Run(cmd, args)
}

func init() {
	DevCmd.Flags().BoolVar(&runFlags.Debug, "debug", false,
		"Open infra-structure debug ports")
	DevCmd.Flags().BoolVar(&runFlags.DryRun, "dry-run", false,
		"Dry-run the infrastructure")
	DevCmd.Flags().BoolVar(&infra, "infra", true,
		"Infrastructure status")
	DevCmd.Flags().BoolVar(&noInfraTmp, "no-infra", false, "")
	DevCmd.Flags().BoolVar(&skipInfra, "skip-infra", false, "")
	DevCmd.Flags().StringVar(&image, "experimental-image", "",
		"Experimental image to run")
	DevCmd.Flags().BoolVarP(&quiet, "quiet", "q", false,
		"Link without watching status")

	if err := DevCmd.Flags().MarkHidden("skip-infra"); err != nil {
		panic(err)
	}

	if err := DevCmd.Flags().MarkHidden("experimental-image"); err != nil {
		panic(err)
	}

	if err := DevCmd.Flags().MarkHidden("no-infra"); err != nil {
		panic(err)
	}

	loadCommandInit()
	DevCmd.AddCommand(cmddevunlink.StopCmd)
}

func loadCommandInit() {
	switch {
	case isCommand("--infra"):
	case isCommand("--no-infra"):
		// if --no-infra or --infra are used,
		// don't load any command runner
	default:
		cmdRunner = &linker{}
	}

	if cmdRunner != nil {
		cmdRunner.Init()
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
