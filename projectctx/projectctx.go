package projectctx

import (
	"context"

	"github.com/wedeploy/cli/config"
	"github.com/wedeploy/cli/projects"
)

// CreateOrUpdate project using context or passed projectID
func CreateOrUpdate(projectIDOverride string) (projectRec projects.Project, err error) {
	var project = projects.Project{
		ProjectID: projectIDOverride,
	}

	if config.Context.ProjectRoot != "" && projectIDOverride == "" {
		var pp, err = projects.Read(config.Context.ProjectRoot)

		if err != nil {
			return projectRec, err
		}

		project = pp.Project()
	}

	projectRec, _, err = projects.CreateOrUpdate(context.Background(), project)
	return projectRec, err
}
