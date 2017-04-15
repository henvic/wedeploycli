package cmddeploy

import (
	"context"
	"os"
	"path/filepath"

	"github.com/hashicorp/errwrap"
	"github.com/wedeploy/cli/apihelper"
	"github.com/wedeploy/cli/config"
	"github.com/wedeploy/cli/containers"
	"github.com/wedeploy/cli/link"
	"github.com/wedeploy/cli/projectctx"
	"github.com/wedeploy/cli/projects"
	"github.com/wedeploy/cli/pullimages"
	"github.com/wedeploy/cli/wdircontext"
)

type linker struct {
	Machine link.Machine
}

func (l *linker) Run() error {
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

func (l *linker) getProject() (projectID string, err error) {
	projectID = setupHost.Project()

	if projectID != "" {
		return projectID, nil
	}

	projectID, err = wdircontext.GetProjectID()

	if err == projects.ErrProjectNotFound {
		return "", nil
	}

	return projectID, err
}

func (l *linker) linkMachineSetup(project projects.Project, csDirs []string) error {
	l.Machine.Project = project

	if quiet {
		l.Machine.ErrStream = os.Stderr
	}

	if err := l.Machine.Setup(csDirs); err != nil {
		return err
	}

	var ctx, end = context.WithCancel(context.Background())
	go l.Machine.Run(end)

	if !quiet {
		l.Machine.Watch()
	}

	<-ctx.Done()
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
