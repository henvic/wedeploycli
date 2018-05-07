package getproject

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/wedeploy/cli/apihelper"
	"github.com/wedeploy/cli/cmd/canceled"
	"github.com/wedeploy/cli/cmd/internal/we"
	"github.com/wedeploy/cli/color"
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
		userProject, err := projectsClient.Get(context.Background(), projectID)

		if err == nil {
			return projectID, nil
		}

		if epf, ok := err.(*apihelper.APIFault); !ok || epf.Status != http.StatusNotFound {
			return "", err
		}

		if err := confirmation(projectID, userProject); err != nil {
			return "", err
		}
	}

	p, ep := projectsClient.Create(context.Background(), projects.Project{
		ProjectID: projectID,
	})

	return p.ProjectID, ep
}

func confirmation(projectID string, userProject projects.Project) error {
	if userProject.ProjectID != "" {
		return nil
	}

	fmt.Println(color.Format(color.FgHiBlack, "Project does not exist."))

	var question = fmt.Sprintf("Do you want to create project \"%s\"?", projectID)

	switch ok, askErr := fancy.Boolean(question); {
	case askErr != nil:
		return askErr
	case ok:
		fmt.Println("")
		return nil
	}

	return canceled.CancelCommand("Deployment canceled.")
}
