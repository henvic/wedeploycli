package cmdprojects

import (
	"github.com/launchpad-project/cli/projects"
	"github.com/spf13/cobra"
)

// ProjectsCmd is used for getting projects about a given scope
var ProjectsCmd = &cobra.Command{
	Use:   "projects",
	Short: "Projects running on WeDeploy",
	Run:   projectsRun,
}

func projectsRun(cmd *cobra.Command, args []string) {
	projects.List()
}
