package cmddeploy

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"
	"github.com/wedeploy/cli/color"
	"github.com/wedeploy/cli/run"
	"github.com/wedeploy/cli/status"
	"github.com/wedeploy/cli/verbose"
)

var (
	runFlags          = run.Flags{}
	infra             bool
	skipInfra         bool
	stopLocalInfraTmp bool
	image             string
)

func init() {
	var df = DeployCmd.Flags()
	df.BoolVar(&runFlags.HTTPS, "https", false, "Enable HTTPS on local remote")
	df.BoolVar(&runFlags.Debug, "debug", false, "Open local infrastructure debug ports")
	df.BoolVar(&runFlags.DryRun, "dry-run-local-infra", false, "Dry-run the local infrastructure")
	df.BoolVar(&infra, "start-local-infra", true, "Start local infrastructure")
	df.BoolVar(&stopLocalInfraTmp, "stop-local-infra", false, "")
	df.BoolVar(&skipInfra, "skip-local-infra", false, "")
	df.StringVar(&image, "experimental-image", "", "Experimental image to run")

	if err := df.MarkHidden("debug"); err != nil {
		panic(err)
	}

	if err := df.MarkHidden("dry-run-local-infra"); err != nil {
		panic(err)
	}

	if err := df.MarkHidden("start-local-infra"); err != nil {
		panic(err)
	}

	if err := df.MarkHidden("skip-local-infra"); err != nil {
		panic(err)
	}

	if err := df.MarkHidden("experimental-image"); err != nil {
		panic(err)
	}

	if err := df.MarkHidden("stop-local-infra"); err != nil {
		panic(err)
	}
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

func maybeShutdown() (err error) {
	if infra {
		return nil
	}

	return run.Stop()
}

func maybeStartInfrastructure(cmd *cobra.Command) error {
	if skipInfra {
		verbose.Debug("Skipping setting up infrastructure.")
		return nil
	}

	if up := checkInfrastructureIsUp(); up {
		if cmd.Flag("start-local-infra").Changed {
			fmt.Fprintf(os.Stderr, "Infrastructure is already running.\n")
		} else {
			verbose.Debug("Infrastructure is already running. Skipping.")
		}

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

	defer signal.Ignore(syscall.SIGINT, syscall.SIGTERM)
	return run.Run(context.Background(), runFlags)
}

func runLocal(cmd *cobra.Command) (err error) {
	if !infra {
		return maybeShutdown()
	}

	if err = maybeStartInfrastructure(cmd); err != nil {
		return err
	}

	if runFlags.DryRun || isCommand("--start-local-infra") || isCommand("--stop-local-infra") {
		return nil
	}

	return (&linker{}).Run()
}

func isCommand(cmd string) bool {
	for _, s := range os.Args {
		if s == cmd {
			return true
		}
	}

	return false
}
