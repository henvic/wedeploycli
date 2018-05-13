package getproject

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/wedeploy/cli/apihelper"
	"github.com/wedeploy/cli/cmd/internal/we"
	"github.com/wedeploy/cli/fancy"
	"github.com/wedeploy/cli/isterm"
	"github.com/wedeploy/cli/projects"
)

// MaybeID tries to get a project ID for using on deployment
func MaybeID(maybe string) (projectID string, err error) {
	projectsClient := projects.New(we.Context())
	projectID = maybe

	if projectID == "" {
		if !isterm.Check() {
			return projectID, errors.New("project ID is missing")
		}

		fmt.Println(fancy.Question("Choose a project ID") + " " + fancy.Tip("default: random"))
		projectID, err = fancy.Prompt()

		if err != nil {
			return projectID, err
		}
	}

	if projectID != "" {
		_, err := projectsClient.Get(context.Background(), projectID)

		if err == nil {
			return projectID, nil
		}

		if epf, ok := err.(*apihelper.APIFault); !ok || epf.Status != http.StatusNotFound {
			return "", err
		}
	}

	p, ep := projectsClient.Create(context.Background(), projects.Project{
		ProjectID: projectID,
	})

	if ep == nil {
		fmt.Printf("Project \"%v\" created.\n", projectID)
	}

	return p.ProjectID, ep
}
