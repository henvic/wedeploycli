package cmdinit

import (
	"errors"
	"os"

	"github.com/launchpad-project/cli/initctx"
	"github.com/spf13/cobra"
)

// InitCmd creates a project or container
var InitCmd = &cobra.Command{
	Use:     "init",
	Short:   "Creates a project or container",
	Run:     initRun,
	Example: `launchpad init (from inside project or container)`,
}

func initRun(cmd *cobra.Command, args []string) {
	var err error
	switch {
	case len(args) == 0:
		err = initctx.New()
		break
	case len(args) == 1 && args[0] == "project":
		err = initctx.NewProject()
	case len(args) == 1 && args[0] == "container":
		err = initctx.NewContainer()
	default:
		err = errors.New("Invalid scope")
	}

	if err != nil {
		println(err.Error())
		os.Exit(1)
	}
}
