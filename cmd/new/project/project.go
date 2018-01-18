package project

import (
	"context"
	"fmt"

	"github.com/wedeploy/cli/color"
	"github.com/wedeploy/cli/projects"

	"github.com/spf13/cobra"
	"github.com/wedeploy/cli/cmd/internal/we"
	"github.com/wedeploy/cli/cmdflagsfromhost"
	"github.com/wedeploy/cli/fancy"
)

var setupHost = cmdflagsfromhost.SetupHost{
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
	RunE: projectRun,
	Args: cobra.NoArgs,
}

func projectRun(cmd *cobra.Command, args []string) (err error) {
	if err := setupHost.Process(context.Background(), we.Context()); err != nil {
		return err
	}

	var projectID = setupHost.Project()

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
	projectsClient := projects.New(we.Context())

	var project, err = projectsClient.Create(context.Background(), projects.Project{
		ProjectID: projectID,
	})

	if err != nil {
		return err
	}

	fmt.Printf(color.Format(
		color.FgHiBlack, "Project \"")+
		"%v"+
		color.Format(color.FgHiBlack, "\" created.")+
		"\n",
		project.ProjectID)
	return nil
}

func init() {
	setupHost.Init(Cmd)
}
