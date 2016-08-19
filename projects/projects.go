package projects

import (
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/wedeploy/api-go"
	"github.com/wedeploy/cli/apihelper"
	"github.com/wedeploy/cli/containers"
	"github.com/wedeploy/cli/verbose"
	"github.com/wedeploy/cli/verbosereq"
)

// Project structure
type Project struct {
	ID           string                `json:"id"`
	Name         string                `json:"name,omitempty"`
	CustomDomain string                `json:"customDomain,omitempty"`
	Health       string                `json:"health,omitempty"`
	Description  string                `json:"description,omitempty"`
	Containers   containers.Containers `json:"containers,omitempty"`
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

// CreateFromJSON a project on WeDeploy
func CreateFromJSON(filename string) error {
	var file, err = os.Open(filename)

	if err != nil {
		return err
	}

	verbose.Debug("Creating project from definition:")
	verbose.Debug(filename)

	var req = apihelper.URL("/projects")
	apihelper.Auth(req)
	req.Body(file)

	return apihelper.Validate(req, req.Post())
}

// Create project. If id is empty, a random one is created by the backend
func Create(id string) (project *Project, err error) {
	var req = apihelper.URL("/projects")

	apihelper.Auth(req)

	if id != "" {
		req.Param("id", id)
	}

	if err := apihelper.Validate(req, req.Post()); err != nil {
		return project, err
	}

	err = apihelper.DecodeJSON(req, &project)
	return project, err
}

// Get project by ID
func Get(id string) (project Project, err error) {
	err = apihelper.AuthGet("/projects/"+id, &project)
	return project, err
}

// List projects
func List() (list []Project, err error) {
	err = apihelper.AuthGet("/projects", &list)
	return list, err
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
func Restart(id string) error {
	var req = apihelper.URL("/restart/project?projectId=" + id)

	apihelper.Auth(req)
	return apihelper.Validate(req, req.Post())
}

// Unlink project
func Unlink(projectID string) error {
	var req = apihelper.URL("/deploy", projectID)
	apihelper.Auth(req)

	return apihelper.Validate(req, req.Delete())
}

// Validate project
func Validate(projectID string) (err error) {
	var req = apihelper.URL("/validators/project/id")
	err = doValidate(projectID, req)

	if err == nil || err != wedeploy.ErrUnexpectedResponse {
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
func ValidateOrCreate(id string) (fid string, err error) {
	project, err := Create(id)

	if err == nil {
		id = project.ID
		return id, nil
	}

	_, err = validateOrCreate(err)
	return id, err
}

// ValidateOrCreateFromJSON project
func ValidateOrCreateFromJSON(filename string) (created bool, err error) {
	return validateOrCreate(CreateFromJSON(filename))
}

func validateOrCreate(err error) (created bool, e error) {
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

func doValidate(projectID string, req *wedeploy.WeDeploy) error {
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
