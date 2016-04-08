package projects

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"strings"

	"github.com/launchpad-project/api.go"
	"github.com/launchpad-project/cli/apihelper"
	"github.com/launchpad-project/cli/verbose"
)

// Project structure
type Project struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Domain      string `json:"domain"`
	Description string `json:"description,omitempty"`
}

var (
	// ErrProjectAlreadyExists happens when a Project ID already exists
	ErrProjectAlreadyExists = errors.New("Project already exists")

	// ErrInvalidProjectID happens when a Project ID is invalid
	ErrInvalidProjectID = errors.New("Invalid project ID")

	outStream io.Writer = os.Stdout
)

func Create(projectID, name string) error {
	var req = apihelper.URL(path.Join("/api/projects"))
	verbose.Debug("Creating project")

	apihelper.Auth(req)
	req.Form("id", projectID)
	req.Form("name", name)

	return apihelper.Validate(req, req.Post())
}

// GetStatus gets the status for the project
func GetStatus(id string) {
	var status string
	var req = apihelper.URL("/api/projects/" + id + "/state")

	apihelper.Auth(req)
	apihelper.ValidateOrExit(req, req.Get())
	apihelper.DecodeJSONOrExit(req, &status)
	fmt.Fprintln(outStream, status+" ("+id+")")
}

// List projects
func List() {
	var projects []Project
	var req = apihelper.URL("/api/projects")

	apihelper.Auth(req)
	apihelper.ValidateOrExit(req, req.Get())
	apihelper.DecodeJSONOrExit(req, &projects)

	for _, project := range projects {
		fmt.Fprintln(outStream, project.ID+" ("+project.Name+")")
	}

	fmt.Fprintln(outStream, "total", len(projects))
}

// Restart restarts a project
func Restart(id string) {
	var req = apihelper.URL("/api/restart/project?projectId=" + id)

	apihelper.Auth(req)
	apihelper.ValidateOrExit(req, req.Post())
}

func Validate(projectID string) (err error) {
	var req = apihelper.URL("/api/validators/project/id")

	apihelper.Auth(req)

	req.Param("value", projectID)

	err = req.Get()

	// @Everything here is to be refactored, this is a hack
	if err == launchpad.ErrUnexpectedResponse {
		body, err := ioutil.ReadAll(req.Response.Body)

		if err != nil {
			return err
		}

		b := string(body)

		if strings.Contains(b, "invalidProjectId") {
			return ErrInvalidProjectID
		}

		if strings.Contains(b, "projectAlreadyExists") {
			return ErrProjectAlreadyExists
		}
	}

	return err
}
