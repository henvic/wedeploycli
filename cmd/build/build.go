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
	"github.com/wedeploy/cli/services"
	"github.com/wedeploy/cli/hooks"
	"github.com/wedeploy/cli/usercontext"
	"github.com/wedeploy/cli/wdircontext"
)

// BuildCmd builds the current project or service
var BuildCmd = &cobra.Command{
	Use:     "build",
	Short:   "Build service(s) (current or all services of a project)",
	PreRunE: cmdargslen.ValidateCmd(0, 0),
	RunE:    buildRun,
}

func init() {
	BuildCmd.Hidden = true
}

func getServicesFromScope() ([]string, error) {
	if config.Context.ServiceRoot != "" {
		_, service := filepath.Split(config.Context.ServiceRoot)
		return []string{service}, nil
	}

	var list, listErr = services.GetListFromDirectory(config.Context.ProjectRoot)

	if listErr != nil {
		return []string{}, listErr
	}

	return list.GetLocations(), nil
}

func buildRun(cmd *cobra.Command, args []string) error {
	if err := checkProjectOrService(); err != nil {
		return err
	}

	if config.Context.Scope == usercontext.GlobalScope {
		return buildService(".")
	}

	var list, err = getServicesFromScope()

	if err != nil {
		return err
	}

	var hasError = false

	for _, c := range list {
		var err = buildService(filepath.Join(config.Context.ProjectRoot, c))

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

func buildService(path string) error {
	var cp, err = services.Read(path)

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

func checkProjectOrService() error {
	var _, _, err = wdircontext.GetProjectOrServiceID()
	var _, errc = services.Read(".")

	if err != nil && os.IsNotExist(errc) {
		return errors.New("fatal: not a project or service")
	}

	if err != nil && errc != nil {
		return errwrap.Wrapf("wedeploy.json error: {{err}}", errc)
	}

	return nil
}
