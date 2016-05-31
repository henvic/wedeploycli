package projects

import (
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/launchpad-project/api.go"
	"github.com/launchpad-project/cli/apihelper"
	"github.com/launchpad-project/cli/config"
	"github.com/launchpad-project/cli/verbose"
	"github.com/launchpad-project/cli/verbosereq"
)

// Project structure
type Project struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Domain      string `json:"domain,omitempty"`
	State       string `json:"state,omitempty"`
	Description string `json:"description,omitempty"`
}

var (
	// ErrProjectAlreadyExists happens when a Project ID already exists
	ErrProjectAlreadyExists = errors.New("Project already exists")

	// ErrInvalidProjectID happens when a Project ID is invalid
	ErrInvalidProjectID = errors.New("Invalid project ID")

	outStream io.Writer = os.Stdout
)

// CreateFromDefinition creates a project on WeDeploy using a JSON definition
func CreateFromDefinition(filename string) error {
	var file, err = os.Open(filename)

	if err != nil {
		return err
	}

	verbose.Debug("Creating project from definition:")
	verbose.Debug(filename)

	var req = apihelper.URL("/projects")
	apihelper.Auth(req)

	maybeSetLocalProjectRoot(req)
	req.Body(file)

	return apihelper.Validate(req, req.Post())
}

// GetStatus gets the status for the project
func GetStatus(id string) string {
	var status string
	apihelper.AuthGetOrExit("/projects/"+id+"/state", &status)
	return status
}

// List projects
func List() {
	var projects []Project
	apihelper.AuthGetOrExit("/projects", &projects)
	printProjects(projects)
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
	err = doValidate(projectID, req)

	if err == nil || err != launchpad.ErrUnexpectedResponse {
		return err
	}

	var errDoc apihelper.APIFault

	err = apihelper.DecodeJSON(req, &errDoc)

	if err != nil {
		return err
	}

	return getValidateAPIFaultError(errDoc)
}

// ValidateOrCreate project
func ValidateOrCreate(filename string) (created bool, err error) {
	err = CreateFromDefinition(filename)

	switch err.(type) {
	case nil:
		return true, err
	case *apihelper.APIFault:
		var ae = err.(*apihelper.APIFault)

		if ae.Has("invalidDocumentValue") {
			return false, nil
		}
	}

	return false, err
}

func doValidate(projectID string, req *launchpad.Launchpad) error {
	apihelper.Auth(req)

	req.Param("value", projectID)

	var err = req.Get()

	verbosereq.Feedback(req)
	return err
}

func getValidateAPIFaultError(errDoc apihelper.APIFault) error {
	switch {
	case errDoc.Has("invalidProjectId"):
		return ErrInvalidProjectID
	case errDoc.Has("projectAlreadyExists"):
		return ErrProjectAlreadyExists
	}

	return errDoc
}

func maybeSetLocalProjectRoot(req *launchpad.Launchpad) {
	if config.Stores["global"].Get("local") == "true" {
		req.Param("source", config.Context.ProjectRoot)
	}
}

func printProject(project Project) {
	fmt.Fprintln(outStream,
		project.ID+"\t"+project.ID+".liferay.io ("+project.Name+") "+project.State)
}

func printProjects(projects []Project) {
	for _, project := range projects {
		printProject(project)
	}

	fmt.Fprintln(outStream, "total", len(projects))
}
