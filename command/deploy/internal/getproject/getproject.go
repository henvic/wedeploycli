package getproject

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/henvic/wedeploycli/apihelper"
	"github.com/henvic/wedeploycli/color"
	"github.com/henvic/wedeploycli/command/canceled"
	"github.com/henvic/wedeploycli/command/internal/we"
	"github.com/henvic/wedeploycli/fancy"
	"github.com/henvic/wedeploycli/isterm"
	"github.com/henvic/wedeploycli/projects"
)

// MaybeID tries to get a project ID for using on deployment
func MaybeID(maybe, region string) (projectID string, err error) {
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
			if region == "" || region == userProject.Region {
				return projectID, nil
			}

			return "", errors.New("cannot change region of existing project")
		}

		if epf, ok := err.(apihelper.APIFault); !ok || epf.Status != http.StatusNotFound {
			return "", err
		}

		if err := confirmation(projectID, userProject); err != nil {
			return "", err
		}
	}

	p, ep := projectsClient.Create(context.Background(), projects.Project{
		ProjectID: projectID,
		Region:    region,
	})

	return p.ProjectID, ep
}

func confirmation(projectID string, userProject projects.Project) error {
	if userProject.ProjectID != "" {
		return nil
	}

	var question = fmt.Sprintf("No project found. %s project \"%s\" and %s the deployment?",
		color.Format(color.FgMagenta, color.Bold, "Create"),
		color.Format(color.FgHiBlack, projectID),
		color.Format(color.FgMagenta, color.Bold, "continue"))

	switch ok, askErr := fancy.Boolean(question); {
	case askErr != nil:
		return askErr
	case ok:
		fmt.Println("")
		return nil
	}

	return canceled.CancelCommand("deployment canceled")
}
