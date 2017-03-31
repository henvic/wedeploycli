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
	"github.com/wedeploy/cli/projectctx"
	"github.com/wedeploy/cli/projects"
	"github.com/wedeploy/cli/prompt"
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

	if errProjectID != nil {
		return errProjectID
	}

	csDirs, err := l.getContainersDirectoriesFromScope()

	if err != nil {
		return err
	}

	if err = pullimages.PullMissingContainersImages(csDirs); err != nil {
		return err
	}

	var projectRec projects.Project
	projectRec, err = projectctx.CreateOrUpdate(projectID)

	if !checkProjectOK(err) {
		return err
	}

	return l.linkMachineSetup(projectRec, csDirs)
}

func (l *linker) getContainersDirectoriesFromScope() ([]string, error) {
	if config.Context.ProjectRoot == "" {
		wd, err := os.Getwd()

		if err != nil {
			return []string{}, err
		}

		_, err = containers.Read(wd)

		switch {
		case err == containers.ErrContainerNotFound:
			err = errwrap.Wrapf("Missing container.json on directory.", err)
		case err != nil:
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

func trySelectProject() (string, error) {
	fmt.Fprintf(os.Stderr, `No project or container on the current context.
Press Enter to cancel or type a project ID to use.

`)

	var id, err = prompt.Prompt("Project ID")

	if len(id) == 0 {
		fmt.Fprintf(os.Stderr, "\nSkipping creating project.\n")
		return "", errors.New(`See http://wedeploy.com/docs`)
	}

	return id, err
}

func (l *linker) getProject() (projectID string, err error) {
	projectID = setupHost.Project()

	if (config.Context.Scope == usercontext.ProjectScope ||
		config.Context.Scope == usercontext.ContainerScope) && projectID != "" {
		return "", errors.New(`Can not use "we run --project" when inside a project`)
	}

	if projectID != "" {
		return projectID, nil
	}

	projectID, err = wdircontext.GetProjectID()

	if err == projects.ErrProjectNotFound {
		return trySelectProject()
	}

	return projectID, err
}

func (l *linker) linkMachineSetup(project projects.Project, csDirs []string) error {
	l.Machine.ProjectID = project.ProjectID

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

func checkProjectOK(err error) bool {
	if ea, ok := err.(*apihelper.APIFault); ok && ea.Has("projectAlreadyExists") {
		return true
	}

	return err == nil
}
