package cmddeploy

import (
	"errors"
	"os"

	"github.com/launchpad-project/cli/cmdcontext"
	"github.com/launchpad-project/cli/containers"
	"github.com/launchpad-project/cli/deploy"
	"github.com/launchpad-project/cli/progress"
	"github.com/spf13/cobra"
)

var (
	noHooks bool
	quiet   bool
	output  string
)

// DeployCmd deploys the current project or container
var DeployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "Deploys the current project or container",
	Run:   deployRun,
	Example: `launchpad deploy
launchpad deploy <container>
launchpad deploy -o welcome.pod`,
}

func deployRun(cmd *cobra.Command, args []string) {
	_, _, err := cmdcontext.GetProjectOrContainerID(args)

	if err != nil {
		println("fatal: not a project")
		os.Exit(1)
	}

	list, err := containers.GetListFromScope()

	if err == nil && output != "" && len(list) != 1 {
		err = errors.New("Only one container can be written to a file at once.")
	}

	if err == nil {
		if !quiet {
			progress.Start()
		}

		if output == "" {
			err = deploy.All(list, &deploy.Flags{
				Hooks: !noHooks,
				Quiet: quiet,
			})
		} else {
			err = deploy.Pack(output, list[0])
		}

		if !quiet {
			progress.Stop()
		}
	}

	if err != nil {
		println(err.Error())
		os.Exit(1)
	}
}

func init() {
	DeployCmd.Flags().BoolVar(&noHooks, "no-hooks", false, "bypass the deploy pre/pos hooks")
	DeployCmd.Flags().BoolVarP(&quiet, "quiet", "q", false, "hide progress bar")
	DeployCmd.Flags().StringVarP(&output, "output", "o", "", "Write to a file, instead of deploying")
}
