package projects

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path"

	"github.com/launchpad-project/api.go"
	"github.com/launchpad-project/cli/apihelper"
	"github.com/launchpad-project/cli/verbose"
)

// Project structure
type Project struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Domain      string `json:"domain"`
	State       string `json:"state"`
	Description string `json:"description,omitempty"`
}

var (
	// ErrProjectAlreadyExists happens when a Project ID already exists
	ErrProjectAlreadyExists = errors.New("Project already exists")

	// ErrInvalidProjectID happens when a Project ID is invalid
	ErrInvalidProjectID = errors.New("Invalid project ID")

	outStream io.Writer = os.Stdout
)

// Create new project
func Create(projectID, name string) error {
	var req = apihelper.URL(path.Join("/projects"))
	verbose.Debug("Creating project")

	apihelper.Auth(req)
	req.Form("id", projectID)
	req.Form("name", name)

	return apihelper.Validate(req, req.Post())
}

// GetStatus gets the status for the project
func GetStatus(id string) {
	var status string
	var req = apihelper.URL("/projects/" + id + "/state")

	apihelper.Auth(req)
	apihelper.ValidateOrExit(req, req.Get())
	apihelper.DecodeJSONOrExit(req, &status)
	fmt.Fprintln(outStream, status+" ("+id+")")
}

// List projects
func List() {
	var projects []Project
	var req = apihelper.URL("/projects")

	apihelper.Auth(req)
	apihelper.ValidateOrExit(req, req.Get())
	apihelper.DecodeJSONOrExit(req, &projects)

	for _, project := range projects {
		fmt.Fprintln(outStream, project.ID+"\t"+project.ID+".liferay.io ("+project.Name+") "+project.State)
	}

	fmt.Fprintln(outStream, "total", len(projects))
}

// Restart restarts a project
func Restart(id string) {
	var req = apihelper.URL("/restart/project?projectId=" + id)

	apihelper.Auth(req)
	apihelper.ValidateOrExit(req, req.Post())
}

// Validate project
func Validate(projectID string) (err error) {
	var req = apihelper.URL("/validators/project/id")

	apihelper.Auth(req)

	req.Param("value", projectID)

	err = req.Get()

	apihelper.RequestVerboseFeedback(req)

	if err == nil || err != launchpad.ErrUnexpectedResponse {
		return err
	}

	var errDoc apihelper.APIFault

	err = apihelper.DecodeJSON(req, &errDoc)

	if err != nil {
		return err
	}

	for _, ed := range errDoc.Errors {
		switch ed.Reason {
		case "invalidProjectId":
			return ErrInvalidProjectID
		case "projectAlreadyExists":
			return ErrProjectAlreadyExists
		}
	}

	return errDoc
}

// ValidateOrCreate project
func ValidateOrCreate(projectID, projectName string) (bool, error) {
	var created bool
	var err = Validate(projectID)

	if err == ErrProjectAlreadyExists {
		return false, nil
	}

	if err != nil {
		return false, err
	}

	err = Create(projectID, projectName)

	if err == nil {
		created = true
	}

	return created, err
}
