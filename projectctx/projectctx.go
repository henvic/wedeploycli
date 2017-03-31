package projectctx

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/wedeploy/cli/config"
	"github.com/wedeploy/cli/projects"
)

// CreateOrUpdate project using context or passed projectID
func CreateOrUpdate(projectIDIfNoCtx string) (projectRec projects.Project, err error) {
	var project projects.Project
	switch config.Context.ProjectRoot {
	case "":
		project.ProjectID = projectIDIfNoCtx
	default:
		var pp, err = projects.Read(config.Context.ProjectRoot)

		if err != nil {
			return projectRec, err
		}

		project = pp.Project()

		if project.ProjectID != projectIDIfNoCtx {
			return projectRec, errors.New("Project ID received does not match context's Project ID")
		}
	}

	var created bool
	projectRec, created, err = projects.CreateOrUpdate(context.Background(), project)

	if created {
		fmt.Fprintf(os.Stdout, "New project %v created.\n", project.ProjectID)
	}

	return projectRec, err
}
