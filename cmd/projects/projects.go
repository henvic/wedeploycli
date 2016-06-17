package cmdprojects

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/wedeploy/cli/projects"
)

// ProjectsCmd is used for getting projects about a given scope
var ProjectsCmd = &cobra.Command{
	Use:   "projects",
	Short: "Projects running on WeDeploy",
	Run:   projectsRun,
}

func projectsRun(cmd *cobra.Command, args []string) {
	var list, err = projects.List()

	if err != nil {
		println(err.Error())
		os.Exit(1)
	}

	printProjects(list)
}

func printProject(project projects.Project) {
	fmt.Fprintf(os.Stdout, "%s\t%s.liferay.io (%s) %s\n",
		project.ID,
		project.ID,
		project.Name,
		project.State)
}

func printProjects(projects []projects.Project) {
	for _, project := range projects {
		printProject(project)
	}

	fmt.Fprintln(os.Stdout, "total", len(projects))
}
