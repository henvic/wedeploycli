package cmdrun

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/hashicorp/errwrap"
	"github.com/spf13/cobra"
	"github.com/wedeploy/cli/apihelper"
	"github.com/wedeploy/cli/cmdargslen"
	"github.com/wedeploy/cli/cmdflagsfromhost"
	"github.com/wedeploy/cli/config"
	"github.com/wedeploy/cli/containers"
	"github.com/wedeploy/cli/link"
	"github.com/wedeploy/cli/projects"
	"github.com/wedeploy/cli/pullimages"
	"github.com/wedeploy/cli/usercontext"
	"github.com/wedeploy/cli/wdircontext"
)

type linker struct {
	Machine link.Machine
}

func (l *linker) Init() {
	setupHost = cmdflagsfromhost.SetupHost{
		Pattern: cmdflagsfromhost.ProjectPattern,
		Requires: cmdflagsfromhost.Requires{
			Local: true,
		},
	}

	setupHost.Init(RunCmd)
}

func (l *linker) PreRun(cmd *cobra.Command, args []string) error {
	if err := cmdargslen.Validate(args, 0, 0); err != nil {
		return err
	}

	return setupHost.Process()
}

func (l *linker) Run(cmd *cobra.Command, args []string) error {
	var projectID, errProjectID = l.getProject()

	switch errProjectID {
	case nil:
	case projects.ErrProjectNotFound:
		fmt.Fprintf(os.Stderr, "No project or container on the current context.\n")

		if _, ec := containers.Read("."); ec == nil {
			fmt.Fprintf(os.Stderr,
				`container.json found: maybe you want to try "we run --project <project>" instead

`)
		}

		fmt.Println(`See http://wedeploy.com/docs`)
		return nil
	default:
		return errProjectID
	}

	csDirs, err := l.getContainersDirectoriesFromScope()

	if err != nil {
		return err
	}

	if err = pullimages.PullMissingContainersImages(csDirs); err != nil {
		return err
	}

	if config.Context.ProjectRoot != "" {
		if err = l.setupLocallyExistingProject(config.Context.ProjectRoot); err != nil {
			return err
		}
	} else {
		projectID, err = projects.ValidateOrCreate(projectID)

		if err != nil {
			return err
		}
	}

	return l.linkMachineSetup(projectID, csDirs)
}

func (l *linker) getContainersDirectoriesFromScope() ([]string, error) {
	if config.Context.ProjectRoot == "" {
		wd, err := os.Getwd()

		if err != nil {
			return []string{}, err
		}

		_, err = containers.Read(wd)

		if err != nil {
			err = errwrap.Wrapf("Can not read container with no project: {{err}}", err)
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
		absList = append(absList, filepath.Join(config.Context.ProjectRoot, item.Location))
	}

	return absList, err
}

func (l *linker) getProject() (string, error) {
	var projectID = setupHost.Project()

	if (config.Context.Scope == usercontext.ProjectScope ||
		config.Context.Scope == usercontext.ContainerScope) && projectID != "" {
		return "", errors.New(`Can not use "we run --project" when inside a project`)
	}

	if projectID != "" {
		return projectID, nil
	}

	return wdircontext.GetProjectID()
}

func (l *linker) linkMachineSetup(projectID string, csDirs []string) error {
	l.Machine.ProjectID = projectID

	if err := l.Machine.Setup(csDirs); err != nil {
		return err
	}

	if quiet {
		l.Machine.ErrStream = os.Stderr
		l.Machine.Run()
		return l.getLinkMachineErrors()
	}

	var queue sync.WaitGroup

	queue.Add(1)

	go func() {
		l.Machine.Run()
	}()

	go func() {
		l.Machine.Watch()
		queue.Done()
	}()

	queue.Wait()
	return l.getLinkMachineErrors()
}

func (l *linker) getLinkMachineErrors() error {
	if len(l.Machine.Errors.List) != 0 {
		return l.Machine.Errors
	}

	return nil
}

func (l *linker) setupLocallyExistingProject(projectPath string) error {
	project, err := projects.Read(projectPath)

	if err != nil {
		return err
	}

	created, err := projects.ValidateOrCreateFromJSON(
		filepath.Join(projectPath, "project.json"))

	if ea, ok := err.(*apihelper.APIFault); ok && ea.Has("exists") {
		return nil
	}

	if created {
		fmt.Fprintf(os.Stdout, "New project %v created.\n", project.ID)
	}

	return err
}
