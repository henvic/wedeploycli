package cmdrun

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/wedeploy/cli/cmd/run/unlink"
	"github.com/wedeploy/cli/cmdargslen"
	"github.com/wedeploy/cli/cmdflagsfromhost"
	"github.com/wedeploy/cli/color"
	"github.com/wedeploy/cli/run"
	"github.com/wedeploy/cli/status"
	"github.com/wedeploy/cli/verbose"
)

type commandRunner interface {
	Init()
	PreRun(cmd *cobra.Command, args []string) error
	Run(cmd *cobra.Command, args []string) error
}

var (
	setupHost        cmdflagsfromhost.SetupHost
	runFlags         = run.Flags{}
	quiet            bool
	infra            bool
	skipInfra        bool
	shutdownInfraTmp bool
	image            string
	cmdRunner        commandRunner
)

// RunCmd runs the WeDeploy local infrastructure
// and / or a project or container
var RunCmd = &cobra.Command{
	Use:     "run",
	Short:   "Run development environment\n",
	PreRunE: preRun,
	RunE:    runRun,
	Example: `  we run
  we run stop
  we run --project chat
  we run --start-infra to startup only the local infrastructure
  we run --shutdown-infra to shutdown the local infrastructure`,
}

func maybeShutdown() (err error) {
	if infra {
		return nil
	}

	return run.Stop()
}

func checkInfrastructureIsUp() bool {
	var try = 0

checkStatus:
	var s, err = status.Get(context.Background())

	if err == nil && s.Status == status.Up {
		return true
	}

	if try < 3 {
		try++
		goto checkStatus
	}

	return false
}

func maybeStartInfrastructure() error {
	if skipInfra {
		verbose.Debug("Skipping setting up infra-structure.")
		return nil
	}

	if up := checkInfrastructureIsUp(); up {
		verbose.Debug("Skipping setting up infra-structure: already running.")
		if runFlags.Debug {
			fmt.Fprintf(os.Stderr, "%v\n",
				color.Format(color.BgRed, color.Bold,
					" debug flag ignored: already running infrastructure "))
		}
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

	return run.Run(context.Background(), runFlags)
}

func preRun(cmd *cobra.Command, args []string) error {
	cmdflagsfromhost.SetLocal()

	if shutdownInfraTmp {
		infra = false
	}

	if cmdRunner != nil {
		return cmdRunner.PreRun(cmd, args)
	}

	if err := cmdargslen.Validate(args, 0, 0); err != nil {
		return err
	}

	return nil
}

func runRun(cmd *cobra.Command, args []string) (err error) {
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
	RunCmd.Flags().BoolVar(&runFlags.Debug, "debug", false,
		"Open infra-structure debug ports")
	RunCmd.Flags().BoolVar(&runFlags.DryRun, "dry-run", false,
		"Dry-run the infrastructure")
	RunCmd.Flags().BoolVar(&infra, "start-infra", true,
		"Infrastructure status")
	RunCmd.Flags().BoolVar(&shutdownInfraTmp, "shutdown-infra", false, "")
	RunCmd.Flags().BoolVar(&skipInfra, "skip-infra", false, "")
	RunCmd.Flags().StringVar(&image, "experimental-image", "",
		"Experimental image to run")
	RunCmd.Flags().BoolVarP(&quiet, "quiet", "q", false,
		"Link without watching status")

	if err := RunCmd.Flags().MarkHidden("skip-infra"); err != nil {
		panic(err)
	}

	if err := RunCmd.Flags().MarkHidden("experimental-image"); err != nil {
		panic(err)
	}

	if err := RunCmd.Flags().MarkHidden("shutdown-infra"); err != nil {
		panic(err)
	}

	loadCommandInit()
	RunCmd.AddCommand(cmdunlink.StopCmd)
}

func loadCommandInit() {
	switch {
	case isCommand("--start-infra"):
	case isCommand("--shutdown-infra"):
		// if --shutdown-infra or --start-infra are used,
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
