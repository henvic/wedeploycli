package cmddeploy

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/launchpad-project/cli/cmdcontext"
	"github.com/launchpad-project/cli/config"
	"github.com/launchpad-project/cli/containers"
	"github.com/launchpad-project/cli/deploy"
	"github.com/launchpad-project/cli/progress"
	"github.com/launchpad-project/cli/projects"
	"github.com/launchpad-project/cli/verbose"
	"github.com/spf13/cobra"
)

var (
	noHooks bool
	quiet   bool
	output  string
)

// DeployCmd deploys the current project or container
var DeployCmd = &cobra.Command{
	Use:    "deploy",
	Short:  "Deploys the current project or container",
	PreRun: checkContext,
	Run:    deployRun,
	Example: `launchpad deploy
launchpad deploy <container>
launchpad deploy -o welcome.pod`,
}

func deployRun(cmd *cobra.Command, args []string) {
	var list, err = containers.GetListFromScope()

	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if output != "" && config.Context.Scope != "container" {
		err = errors.New("Can only output a single container to file, not a whole project.")
	}

	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	var success []string
	success, err = tryDeployMaybeQuiet(list)

	// wait for next tick to the progress bar cleanup goroutine to clear the
	// buffer and end, so the message here is not erased by it
	time.Sleep(time.Millisecond)

	for _, s := range success {
		fmt.Println(s)
	}

	if len(success) != 0 && err != nil {
		fmt.Println("")
	}

	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if output == "" {
		verbose.Debug("Restarting project")
		projects.Restart(config.Stores["project"].Get("id"))
	}
}

func checkContext(cmd *cobra.Command, args []string) {
	var _, _, err = cmdcontext.GetProjectOrContainerID(args)

	if err != nil {
		println("fatal: not a project")
		os.Exit(1)
	}
}

func tryDeploy(list []string) (success []string, err error) {
	if output == "" {
		var success, err = deploy.All(list, &deploy.Flags{
			Hooks: !noHooks,
			Quiet: quiet,
		})

		return success, err
	}

	return success, deploy.Pack(output, list[0])
}

func tryDeployMaybeQuiet(list []string) (success []string, err error) {
	if !quiet {
		progress.Start()
	}

	success, err = tryDeploy(list)

	if !quiet {
		progress.Stop()
	}

	return success, err
}

func init() {
	DeployCmd.Flags().BoolVar(&noHooks, "no-hooks", false, "Bypass the deploy pre/pos hooks")
	DeployCmd.Flags().BoolVarP(&quiet, "quiet", "q", false, "Hide progress bar")
	DeployCmd.Flags().StringVarP(&output, "output", "o", "", "Write to a file, instead of deploying")
}
