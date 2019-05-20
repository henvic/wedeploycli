package project

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/wedeploy/cli/cmd/internal/we"
	"github.com/wedeploy/cli/cmdflagsfromhost"
	"github.com/wedeploy/cli/color"
	"github.com/wedeploy/cli/fancy"
	"github.com/wedeploy/cli/projects"
	"github.com/wedeploy/cli/prompt"
)

// Don't use this anywhere but on Cmd.RunE
var sh = cmdflagsfromhost.SetupHost{
	Pattern: cmdflagsfromhost.RegionPattern | cmdflagsfromhost.ProjectAndRemotePattern,

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

	return Run(sh.Project(), sh.Region())
}

// Run command for creating a project
func Run(projectID, region string) (err error) {
	if projectID != "" {
		return createProject(projectID, region)
	}

	fmt.Println(fancy.Question("Choose a project ID") + " " + fancy.Tip("default: random"))
	projectID, err = fancy.Prompt()

	if err != nil {
		return err
	}

	return createProject(projectID, region)
}

func createProject(projectID, region string) error {
	var err error
	var project projects.Project

	if region == "" {
		region, err = promptRegion()

		if err != nil {
			return err
		}
	}

	wectx := we.Context()
	projectsClient := projects.New(wectx)

	project, err = projectsClient.Create(context.Background(), projects.Project{
		ProjectID: projectID,
		Region:    region,
	})

	if err != nil {
		return err
	}

	if strings.Contains(project.ProjectID, "-") {
		return createdEnvironment(project)
	}

	return createdProject(project)
}

func promptRegion() (string, error) {
	wectx := we.Context()
	projectsClient := projects.New(wectx)

	fmt.Printf("Please %s a region from the list below.\n",
		color.Format(color.FgMagenta, color.Bold, "select"))

	var regions, err = projectsClient.Regions(context.Background())

	if err != nil {
		return "", err
	}

	var m = map[string]int{}

	fmt.Println(color.Format(color.FgHiBlack, "#\tRegion"))

	for k, v := range regions {
		fmt.Printf("%d\t%v (%v)\n", k+1, v.Location, v.Name)
		m[v.Name] = k + 1
	}

	fmt.Print("Choice: ")
	var i int
	i, err = prompt.SelectOption(len(regions), m)

	if err != nil {
		return "", err
	}

	return regions[i].Name, nil
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
