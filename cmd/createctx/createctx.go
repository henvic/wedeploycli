package cmdcreate

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/hashicorp/errwrap"
	"github.com/spf13/cobra"
	"github.com/wedeploy/cli/config"
	"github.com/wedeploy/cli/createctx"
)

// CreateCmd creates a project or container
var CreateCmd = &cobra.Command{
	Use:   "create <project/container id>",
	Short: "Creates a project or container",
	Long: `Use "we create" to create projects and containers.
You can create a project anywhere on your machine.
Containers can only be created from inside projects and
are stored on the first subdirectory of its project.`,
	RunE:    createRun,
	Example: `we create relay`,
}

func createRun(cmd *cobra.Command, args []string) error {
	var err error

	switch len(args) {
	case 0:
		err = createctx.New("", "")
	case 1:
		err = createctx.New(args[0], "")
	case 2:
		err = cwdContextAndCreate(args[0], args[1])
	default:
		err = errors.New("Invalid number of arguments.")
	}

	return err
}

func cwdContextAndCreate(id, directory string) error {
	var workingDir, err = os.Getwd()

	if err != nil {
		panic(err)
	}

	err = os.MkdirAll(directory, 0775)

	if err != nil {
		return errwrap.Wrapf("Can't create new directory: {{err}}", err)
	}

	abs, eabs := filepath.Abs(directory)

	if eabs != nil {
		panic(eabs)
	}

	if err = os.Chdir(abs); err != nil {
		return err
	}

	if err = config.Setup(); err != nil {
		return errwrap.Wrapf("Can't reset config object: {{err}}", err)
	}

	var cerr = createctx.New(id, abs)

	if err = os.Chdir(workingDir); err != nil {
		panic(err)
	}

	return cerr
}
