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
	Example: `launchpad create (from inside project or container)`,
}

func createRun(cmd *cobra.Command, args []string) {
	var err error
	switch {
	case len(args) == 0:
		err = createctx.New()
		break
	case len(args) == 1 && args[0] == "project":
		err = createctx.NewProject()
	case len(args) == 1 && args[0] == "container":
		err = createctx.NewContainer()
	default:
		err = errors.New("Invalid scope")
	}

	if err != nil {
		println(err.Error())
		os.Exit(1)
	}
}
