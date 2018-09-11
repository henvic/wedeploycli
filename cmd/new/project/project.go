package project

import (
	"context"
	"fmt"
	"strings"

	"github.com/wedeploy/cli/color"
	"github.com/wedeploy/cli/projects"

	"github.com/spf13/cobra"
	"github.com/wedeploy/cli/cmd/internal/we"
	"github.com/wedeploy/cli/cmdflagsfromhost"
	"github.com/wedeploy/cli/fancy"
)

// Don't use this anywhere but on Cmd.RunE
var sh = cmdflagsfromhost.SetupHost{
	Pattern: cmdflagsfromhost.ProjectAndRemotePattern,

	Requires: cmdflagsfromhost.Requires{
		Auth: true,
	},
}

// Cmd is the command for creating a new project
var Cmd = &cobra.Command{
	Use:   "project",
	Short: "Create new project",
	// Don't use other run functions besides RunE here
	// or fix NewCmd to call it correctly
	RunE: runE,
	Args: cobra.NoArgs,
}

func runE(cmd *cobra.Command, args []string) (err error) {
	if err := sh.Process(context.Background(), we.Context()); err != nil {
		return err
	}

	return Run(sh.Project())
}

// Run command for creating a project
func Run(projectID string) (err error) {
	if projectID != "" {
		return createProject(projectID)
	}

	fmt.Println(fancy.Question("Choose a project ID") + " " + fancy.Tip("default: random"))
	projectID, err = fancy.Prompt()

	if err != nil {
		return err
	}

	return createProject(projectID)
}

func createProject(projectID string) error {
	wectx := we.Context()
	projectsClient := projects.New(wectx)

	var project, err = projectsClient.Create(context.Background(), projects.Project{
		ProjectID: projectID,
	})

	if err != nil {
		return err
	}

	if strings.Contains(project.ProjectID, "-") {
		return createdEnvironment(project)
	}

	return createdProject(project)
}

func createdProject(p projects.Project) error {
	wectx := we.Context()

	fmt.Printf(color.Format(
		color.FgHiBlack, "Project \"")+
		"%v"+
		color.Format(color.FgHiBlack, "\" created on ")+
		wectx.InfrastructureDomain()+
		color.Format(color.FgHiBlack, ".")+
		"\n",
		p.ProjectID)
	return nil
}

func createdEnvironment(p projects.Project) error {
	wectx := we.Context()

	fmt.Printf(color.Format(
		color.FgHiBlack, "Project environment \"")+
		"%v"+
		color.Format(color.FgHiBlack, "\" created on ")+
		wectx.InfrastructureDomain()+
		color.Format(color.FgHiBlack, ".")+
		"\n",
		p.ProjectID)
	return nil
}

func init() {
	sh.Init(Cmd)
}
