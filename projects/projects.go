package projects

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

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
	// ErrProjectNotFound happens when a project.json is not found
	ErrProjectNotFound = errors.New("Project not found")

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

// Read a project directory properties (defined by a project.json on it)
func Read(path string) (*Project, error) {
	var content, err = ioutil.ReadFile(filepath.Join(path, "project.json"))
	var data Project

	if err != nil {
		return nil, readValidate(data, err)
	}

	err = json.Unmarshal(content, &data)

	return &data, readValidate(data, err)
}

func readValidate(project Project, err error) error {
	switch {
	case os.IsNotExist(err):
		return ErrProjectNotFound
	case err != nil:
		return err
	case project.ID == "":
		return ErrInvalidProjectID
	default:
		return err
	}
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
	if config.Global.Local {
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
