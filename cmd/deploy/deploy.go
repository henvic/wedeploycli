package cmddeploy

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/wedeploy/cli/cmdcontext"
	"github.com/wedeploy/cli/config"
	"github.com/wedeploy/cli/containers"
	"github.com/wedeploy/cli/deploy"
	"github.com/wedeploy/cli/deploymachine"
	"github.com/wedeploy/cli/progress"
	"github.com/wedeploy/cli/projects"
	"github.com/wedeploy/cli/verbose"
)

var (
	noHooks bool
	quiet   bool
	output  string
)

// ErrOutputScope happens when we deploy -o is used outside a container scope
var ErrOutputScope = errors.New("Can only output a single container to file, not a whole project.")

// DeployCmd deploys the current project or container
var DeployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "Deploys the current project or container",
	Run:   deployRun,
	Example: `we deploy
we deploy <container>
we deploy -o welcome.pod`,
}

func getDeployListFromScope() []string {
	var list, err = containers.GetListFromScope()

	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	return list
}

func verifyOutputScope() {
	// should output to file only when scope is container
	if output != "" && config.Context.Scope != "container" {
		fmt.Fprintln(os.Stderr, ErrOutputScope)
		os.Exit(1)
	}
}

func deployContainersFeedback(success []string, err error) {
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
}

func deployRun(cmd *cobra.Command, args []string) {
	var projectID, _ = getProjectOrContainerID(args)
	verifyOutputScope()
	var success, err = tryDeployMaybeQuiet(
		projectID,
		getDeployListFromScope())

	// wait for next tick to the progress bar cleanup goroutine to clear the
	// buffer and end, so the message here is not erased by it
	time.Sleep(time.Millisecond)

	deployContainersFeedback(success, err)

	if output == "" {
		verbose.Debug("Restarting project")
		projects.Restart(projectID)
	}
}

func getProjectOrContainerID(args []string) (string, string) {
	var project, container, err = cmdcontext.GetProjectOrContainerID(args)

	if err != nil {
		println("fatal: not a project")
		os.Exit(1)
	}

	return project, container
}

func tryDeploy(projectID string, list []string) (success []string, err error) {
	if output == "" {
		var success, err = deploymachine.All(projectID, list, &deploy.Flags{
			Hooks: !noHooks,
			Quiet: quiet,
		})

		return success, err
	}

	return success, deploy.Pack(output, list[0])
}

func tryDeployMaybeQuiet(projectID string, list []string) (success []string, err error) {
	var local = config.Global.Local
	if !quiet && !local {
		progress.Start()
	}

	success, err = tryDeploy(projectID, list)

	if !quiet && !local {
		progress.Stop()
	}

	return success, err
}

func init() {
	DeployCmd.Flags().BoolVar(&noHooks, "no-hooks", false, "Bypass the deploy pre/pos hooks")
	DeployCmd.Flags().BoolVarP(&quiet, "quiet", "q", false, "Hide progress bar")
	DeployCmd.Flags().StringVarP(&output, "output", "o", "", "Write to a file, instead of deploying")
}
