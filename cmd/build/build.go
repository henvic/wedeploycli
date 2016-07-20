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
	checkProjectOrContainer(args)

	if config.Context.Scope == "global" {
		var err = buildContainer(".")

		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			os.Exit(1)
		}

		return
	}

	var list = getContainersFromScope()

	var hasError = false

	for _, c := range list {
		var err = buildContainer(filepath.Join(config.Context.ProjectRoot, c))

		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			hasError = true
		}
	}

	if hasError {
		os.Exit(1)
	}
}

func buildContainer(path string) error {
	var container, err = containers.Read(path)

	if err != nil {
		return err
	}

	if container.Hooks == nil || (container.Hooks.BeforeBuild == "" &&
		container.Hooks.Build == "" &&
		container.Hooks.AfterBuild == "") {
		verbose.Debug("container " + container.ID + " has no build hooks")
		return nil
	}

	return container.Hooks.Run(hooks.Build)
}

func checkProjectOrContainer(args []string) {
	var _, _, err = cmdcontext.GetProjectOrContainerID(args)
	var _, errc = containers.Read(".")

	if err != nil && os.IsNotExist(errc) {
		println("fatal: not a project or container")
		os.Exit(1)
	}

	if err != nil && errc != nil {
		fmt.Fprintf(os.Stderr, "container.json error: %v\n", errc)
		os.Exit(1)
	}
}
