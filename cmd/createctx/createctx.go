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
	Run:     createRun,
	Example: `we create relay`,
}

func handleError(err error) {
	if err != nil {
		println(err.Error())
		os.Exit(1)
	}
}

func createRun(cmd *cobra.Command, args []string) {
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

	handleError(err)
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

	err = os.Chdir(abs)
	config.Setup()

	if err != nil {
		return err
	}

	var cerr = createctx.New(id, abs)

	err = os.Chdir(workingDir)

	if err != nil {
		panic(err)
	}

	return cerr
}
