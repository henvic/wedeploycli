package cmddev

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/wedeploy/cli/run"
)

var (
	runFlags   = run.Flags{}
	infra      bool
	noInfraTmp bool
	image      string
)

// DevCmd runs the WeDeploy local infrastructure
var DevCmd = &cobra.Command{
	Use:     "dev",
	Short:   "Run development environment for a project or container",
	PreRun:  devPreRun,
	RunE:    devRun,
	Example: `  we dev`,
}

func maybeShutdown() (err error) {
	if infra {
		return nil
	}

	err = run.Stop()

	return err
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

func devPreRun(cmd *cobra.Command, args []string) {
	if noInfraTmp {
		infra = false
	}
}

func devRun(cmd *cobra.Command, args []string) (err error) {
	if !infra {
		return maybeShutdown()
	}

	if err = maybeStartInfrastructure(); err != nil {
		return err
	}

	return nil
}

func init() {
	DevCmd.Flags().BoolVar(&runFlags.Debug, "debug", false,
		"Open infra-structure debug ports")
	DevCmd.Flags().BoolVar(&runFlags.DryRun, "dry-run", false,
		"Dry-run test of the infrastructure")
	DevCmd.Flags().BoolVar(&infra, "infra", true,
		"Infrastructure status")
	DevCmd.Flags().BoolVar(&noInfraTmp, "no-infra", false, "")
	DevCmd.Flags().StringVar(&image, "experimental-image", "",
		"Experimental image to run")

	if err := DevCmd.Flags().MarkHidden("experimental-image"); err != nil {
		panic(err)
	}

	if err := DevCmd.Flags().MarkHidden("no-infra"); err != nil {
		panic(err)
	}
}
