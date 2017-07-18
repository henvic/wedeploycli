package cmdbuild

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/hashicorp/errwrap"
	"github.com/spf13/cobra"
	"github.com/wedeploy/cli/cmdargslen"
	"github.com/wedeploy/cli/config"
	"github.com/wedeploy/cli/containers"
	"github.com/wedeploy/cli/hooks"
	"github.com/wedeploy/cli/usercontext"
	"github.com/wedeploy/cli/wdircontext"
)

// BuildCmd builds the current project or container
var BuildCmd = &cobra.Command{
	Use:     "build",
	Short:   "Build container(s) (current or all containers of a project)",
	PreRunE: cmdargslen.ValidateCmd(0, 0),
	RunE:    buildRun,
}

func init() {
	BuildCmd.Hidden = true
}

func getContainersFromScope() ([]string, error) {
	if config.Context.ContainerRoot != "" {
		_, container := filepath.Split(config.Context.ContainerRoot)
		return []string{container}, nil
	}

	var list, listErr = containers.GetListFromDirectory(config.Context.ProjectRoot)

	if listErr != nil {
		return []string{}, listErr
	}

	return list.GetLocations(), nil
}

func buildRun(cmd *cobra.Command, args []string) error {
	if err := checkProjectOrContainer(); err != nil {
		return err
	}

	if config.Context.Scope == usercontext.GlobalScope {
		return buildContainer(".")
	}

	var list, err = getContainersFromScope()

	if err != nil {
		return err
	}

	var hasError = false

	for _, c := range list {
		var err = buildContainer(filepath.Join(config.Context.ProjectRoot, c))

		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			hasError = true
		}
	}

	if hasError {
		return errors.New("Build hooks failure")
	}

	return nil
}

func buildContainer(path string) error {
	var cp, err = containers.Read(path)

	if err != nil {
		return err
	}

	if cp.Hooks == nil || (cp.Hooks.BeforeBuild == "" &&
		cp.Hooks.Build == "" &&
		cp.Hooks.AfterBuild == "") {
		println("> [" + cp.ID + "] has no build hooks")
		return nil
	}

	return cp.Hooks.Run(hooks.Build, filepath.Join(path), cp.ID)
}

func checkProjectOrContainer() error {
	var _, _, err = wdircontext.GetProjectOrContainerID()
	var _, errc = containers.Read(".")

	if err != nil && os.IsNotExist(errc) {
		return errors.New("fatal: not a project or container")
	}

	if err != nil && errc != nil {
		return errwrap.Wrapf("wedeploy.json error: {{err}}", errc)
	}

	return nil
}
