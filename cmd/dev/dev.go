package cmddev

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/wedeploy/cli/cmdflagsfromhost"
	"github.com/wedeploy/cli/run"
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
	stop       bool
	infra      bool
	noInfraTmp bool
	image      string
	cmdRunner  commandRunner
)

// DevCmd runs the WeDeploy local infrastructure
// and / or a project or container
// This module makes some abuse of the cmdflagsfromhost module
// setting up different options if --stop is used or not
var DevCmd = &cobra.Command{
	Use:     "dev",
	Short:   "Run development environment for a project or container",
	PreRunE: devPreRun,
	RunE:    devRun,
	Example: `  we dev
  we dev --stop
  we dev data.chat
  we dev --stop data.chat
  we dev --project chat
  we dev --stop --project data --container chat
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
	if !infra {
		return maybeShutdown()
	}

	if stop {
		return cmdRunner.Run(cmd, args)
	}

	if err = maybeStartInfrastructure(); err != nil {
		return err
	}

	if cmdRunner != nil {
		return cmdRunner.Run(cmd, args)
	}

	return nil
}

func init() {
	DevCmd.Flags().BoolVar(&runFlags.Debug, "debug", false,
		"Open infra-structure debug ports")
	DevCmd.Flags().BoolVar(&runFlags.DryRun, "dry-run", false,
		"Dry-run the infrastructure")
	DevCmd.Flags().BoolVar(&stop, "stop", false,
		"Stop project or container")
	DevCmd.Flags().BoolVar(&infra, "infra", true,
		"Infrastructure status")
	DevCmd.Flags().BoolVar(&noInfraTmp, "no-infra", false, "")
	DevCmd.Flags().StringVar(&image, "experimental-image", "",
		"Experimental image to run")
	DevCmd.Flags().BoolVarP(&quiet, "quiet", "q", false,
		"Minimize output feedback.")

	if err := DevCmd.Flags().MarkHidden("experimental-image"); err != nil {
		panic(err)
	}

	if err := DevCmd.Flags().MarkHidden("no-infra"); err != nil {
		panic(err)
	}

	loadCommandInit()
}

func loadCommandInit() {
	switch {
	// only --stop / unlink accepts both --project and --container
	// so we want to use it for these
	case isCommand("--stop") ||
		isCommand("--help") ||
		isCommand("-h"):
		cmdRunner = &unlinker{}
	case isCommand("--infra"):
	case isCommand("--no-infra"):
		// if --no-infra or --infra are used,
		// don't load any command runner
		// this should be after the --stop case above
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
