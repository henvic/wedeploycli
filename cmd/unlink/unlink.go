package cmdunlink

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/wedeploy/cli/cmdcontext"
	"github.com/wedeploy/cli/containers"
	"github.com/wedeploy/cli/projects"
)

// UnlinkCmd unlinks the given project or container locally
var UnlinkCmd = &cobra.Command{
	Use:   "unlink",
	Short: "Unlinks the given project or container locally",
	Run:   unlinkRun,
	Example: `we unlink
we unlink <project>
we unlink <project> <container>
we unlink <container>`,
}

func unlinkRun(cmd *cobra.Command, args []string) {
	var project, container, err = cmdcontext.GetProjectOrContainerID(args)

	switch {
	case err != nil:
		println("fatal: not a project")
		os.Exit(1)
	case container == "":
		err = projects.Unlink(project)
	default:
		err = containers.Unlink(project, container)
	}

	if err != nil {
		println(err.Error())
		os.Exit(1)
	}
}
