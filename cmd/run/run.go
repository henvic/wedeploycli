package cmdrun

import (
	"errors"

	"github.com/spf13/cobra"
	"github.com/wedeploy/cli/run"
)

// RunCmd runs the WeDeploy infrastructure for development locally
var RunCmd = &cobra.Command{
	Use:   "run",
	Short: "Run WeDeploy infrastructure for development locally",
	RunE:  runRun,
}

var (
	image  string
	debug  bool
	detach bool
	dryRun bool
)

func runRun(cmd *cobra.Command, args []string) error {
	if len(args) != 0 {
		return errors.New("Invalid number of arguments.")
	}

	if image != "" {
		println("INFO: Using experimental image " + image)
		run.WeDeployImage = image
	}

	return run.Run(run.Flags{
		Debug:  debug,
		Detach: detach,
		DryRun: dryRun,
	})
}

func init() {
	// debug can only run on the first time
	RunCmd.Flags().BoolVar(&debug, "debug", false, "Open debug ports")

	RunCmd.Flags().BoolVarP(&detach, "detach", "d", false,
		"Run in background")

	RunCmd.Flags().BoolVar(&dryRun, "dry-run", false,
		"Obtain a summary of what docker command is invoked")

	RunCmd.Flags().StringVar(&image, "experimental-image", "", "Experimental image to run")

	if err := RunCmd.Flags().MarkHidden("experimental-image"); err != nil {
		panic(err)
	}
}
