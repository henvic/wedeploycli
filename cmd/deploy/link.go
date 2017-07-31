package cmddeploy

import (
	"context"
	"net/http"
	"os"

	"github.com/hashicorp/errwrap"
	"github.com/wedeploy/cli/apihelper"
	"github.com/wedeploy/cli/createuser"
	"github.com/wedeploy/cli/link"
	"github.com/wedeploy/cli/projects"
	"github.com/wedeploy/cli/pullimages"
	"github.com/wedeploy/cli/services"
)

type linker struct {
	Project string
	Service string
	Machine link.Machine
}

func maybeCreateLocalUser(ctx context.Context) error {
	var _, err = projects.List(ctx)

	if err == nil {
		return nil
	}

	ea, ok := err.(*apihelper.APIFault)

	if !ok || ea.Status != http.StatusUnauthorized {
		return err
	}

	return createuser.Try(ctx)
}

func (l *linker) Run() error {
	if err := maybeCreateLocalUser(context.Background()); err != nil {
		return err
	}

	var projectID = setupHost.Project()

	csDirs, err := l.getServicesDirectories()

	if err != nil {
		return err
	}

	if err = pullimages.PullMissingServicesImages(csDirs); err != nil {
		return err
	}

	projectRec, _, err := projects.CreateOrUpdate(context.Background(), projects.Project{
		ProjectID: projectID,
	})

	if !checkProjectOK(err) {
		return err
	}

	return l.linkMachineSetup(projectRec, csDirs)
}

func (l *linker) getServicesDirectories() ([]string, error) {
	wd, err := os.Getwd()

	if err != nil {
		return []string{}, err
	}

	_, err = services.Read(wd)

	switch {
	case err == services.ErrServiceNotFound:
		err = errwrap.Wrapf("Missing wedeploy.json on directory.", err)
	case err != nil:
		err = errwrap.Wrapf("Can not read service with no project: {{err}}", err)
	}

	return []string{wd}, err
}

func (l *linker) linkMachineSetup(project projects.Project, csDirs []string) error {
	l.Machine.Project = project

	if quiet {
		l.Machine.ErrStream = os.Stderr
	}

	var renameServiceIDs = []link.RenameServiceID{}

	if setupHost.Service() != "" {
		renameServiceIDs = append(renameServiceIDs, link.RenameServiceID{
			Any: true,
			To:  setupHost.Service(),
		})
	}

	if err := l.Machine.Setup(csDirs, renameServiceIDs...); err != nil {
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
