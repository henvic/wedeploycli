package cmdbuild

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/wedeploy/cli/cmdcontext"
	"github.com/wedeploy/cli/config"
	"github.com/wedeploy/cli/containers"
	"github.com/wedeploy/cli/hooks"
	"github.com/wedeploy/cli/verbose"
)

// BuildCmd builds the current project or container
var BuildCmd = &cobra.Command{
	Use:   "build",
	Short: "Build container(s) (current or all containers of a project)",
	Run:   buildRun,
}

func getContainersFromScope() []string {
	if config.Context.ContainerRoot != "" {
		_, container := filepath.Split(config.Context.ContainerRoot)
		return []string{container}
	}

	var list, err = containers.GetListFromDirectory(config.Context.ProjectRoot)

	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	return list
}

func buildRun(cmd *cobra.Command, args []string) {
	// calling it for side-effects
	getProjectOrContainerID(args)
	var list = getContainersFromScope()

	var hasError = false

	for _, c := range list {
		var container, err = containers.Read(
			filepath.Join(config.Context.ProjectRoot, c))

		if err != nil {
			println(err.Error())
			hasError = true
		}

		if container.Hooks == nil || (container.Hooks.BeforeBuild == "" &&
			container.Hooks.Build == "" &&
			container.Hooks.AfterBuild == "") {
			verbose.Debug("container " + container.ID + " has no build hooks")
			continue
		}

		if err = container.Hooks.Run(hooks.Build); err != nil {
			println(err.Error())
		}
	}

	if hasError {
		os.Exit(1)
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
