package cmdcreate

import (
	"errors"
	"os"

	"github.com/launchpad-project/cli/createctx"
	"github.com/spf13/cobra"
)

// CreateCmd creates a project or container
var CreateCmd = &cobra.Command{
	Use:     "create",
	Short:   "Creates a project or container",
	Run:     createRun,
	Example: `we create (from inside project or container)`,
}

func handleError(err error) {
	if err != nil {
		println(err.Error())
		os.Exit(1)
	}
}

func createRun(cmd *cobra.Command, args []string) {
	var err error
	var length = len(args)
	switch {
	case length == 0:
		err = createctx.New()
	case length == 1 && args[0] == "project":
		err = createctx.NewProject()
	case length == 1 && args[0] == "container":
		err = createctx.NewContainer()
	default:
		err = errors.New("Invalid scope")
	}

	handleError(err)
}
