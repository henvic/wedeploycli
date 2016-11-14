package cmdlink

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/hashicorp/errwrap"
	"github.com/spf13/cobra"
	"github.com/wedeploy/cli/cmdflagsfromhost"
	"github.com/wedeploy/cli/config"
	"github.com/wedeploy/cli/containers"
	"github.com/wedeploy/cli/link"
	"github.com/wedeploy/cli/projects"
)

// LinkCmd links the given project or container locally
var LinkCmd = &cobra.Command{
	Use:     "link <host> or --project <project>",
	Short:   "Links the given project or container locally",
	PreRunE: preRun,
	RunE:    linkRun,
	Example: `we link
	we link data`,
}

var quiet bool

var setupHost = cmdflagsfromhost.SetupHost{
	Pattern:             cmdflagsfromhost.ProjectPattern,
	UseProjectDirectory: true,
}

func init() {
	LinkCmd.Flags().BoolVarP(
		&quiet,
		"quiet",
		"q",
		false,
		"Link without watching status.")

	setupHost.Init(LinkCmd)
}

func getContainersDirectoriesFromScope() ([]string, error) {
	if config.Context.ProjectRoot == "" {
		wd, err := os.Getwd()

		if err != nil {
			return []string{}, err
		}

		_, err = containers.Read(wd)

		if err != nil {
			err = errwrap.Wrapf("Can't find project-orphan container: {{err}}", err)
		}

		return []string{wd}, err
	}

	if config.Context.ContainerRoot != "" {
		return []string{config.Context.ContainerRoot}, nil
	}

	var list, err = containers.GetListFromDirectory(config.Context.ProjectRoot)

	if err != nil {
		err = errwrap.Wrapf("Error retrieving containers list from directory: {{err}}", err)
	}

	var absList = []string{}

	for _, item := range list {
		absList = append(absList, filepath.Join(config.Context.ProjectRoot, item))
	}

	return absList, err
}

func preRun(cmd *cobra.Command, args []string) error {
	return setupHost.Process(args)
}

func linkRun(cmd *cobra.Command, args []string) error {
	var projectID = setupHost.Project()

	csDirs, err := getContainersDirectoriesFromScope()

	if err != nil {
		return err
	}

	if projectID == "" && config.Context.ProjectRoot != "" {
		if err = setupLocallyExistingProject(config.Context.ProjectRoot); err != nil {
			return err
		}
	} else {
		projectID, err = projects.ValidateOrCreate(projectID)

		if err != nil {
			return err
		}
	}

	return linkMachineSetup(projectID, csDirs)
}

func linkMachineSetup(projectID string, csDirs []string) error {
	var m = &link.Machine{
		ProjectID: projectID,
	}

	if err := m.Setup(csDirs); err != nil {
		return err
	}

	if quiet {
		m.ErrStream = os.Stderr
		m.Run()
		return getLinkMachineErrors(m)
	}

	var queue sync.WaitGroup

	queue.Add(1)

	go func() {
		m.Run()
	}()

	go func() {
		m.Watch()
		queue.Done()
	}()

	queue.Wait()
	return getLinkMachineErrors(m)
}

func getLinkMachineErrors(m *link.Machine) error {
	if len(m.Errors.List) != 0 {
		return m.Errors
	}

	return nil
}

func setupLocallyExistingProject(projectPath string) error {
	project, err := projects.Read(projectPath)

	if err != nil {
		return err
	}

	created, err := projects.ValidateOrCreateFromJSON(
		filepath.Join(projectPath, "project.json"))

	if created {
		fmt.Fprintf(os.Stdout, "New project %v created.\n", project.ID)
	}

	return err
}
