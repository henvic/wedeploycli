package projects

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"

	"github.com/hashicorp/errwrap"
	"github.com/wedeploy/api-go"
	"github.com/wedeploy/cli/apihelper"
	"github.com/wedeploy/cli/containers"
	"github.com/wedeploy/cli/verbose"
	"github.com/wedeploy/cli/verbosereq"
)

// Project structure
type Project struct {
	ID            string                `json:"id"`
	CustomDomains []string              `json:"customDomains,omitempty"`
	Health        string                `json:"health,omitempty"`
	HomeContainer string                `json:"homeContainer,omitempty"`
	Description   string                `json:"description,omitempty"`
	Containers    containers.Containers `json:"containers,omitempty"`
}

var (
	// ErrProjectNotFound happens when a project.json is not found
	ErrProjectNotFound = errors.New("Project not found")

	// ErrProjectAlreadyExists happens when a Project ID already exists
	ErrProjectAlreadyExists = errors.New("Project already exists")

	// ErrInvalidProjectID happens when a Project ID is invalid
	ErrInvalidProjectID = errors.New("Invalid project ID")
)

// CreateFromJSON a project on WeDeploy
func CreateFromJSON(ctx context.Context, filename string) error {
	var p, err = ioutil.ReadFile(filename)

	if err != nil {
		return err
	}

	verbose.Debug("Creating project from definition:")
	verbose.Debug(filename)

	var req = apihelper.URL(ctx, "/projects")
	apihelper.Auth(req)
	req.Body(bytes.NewReader(p))

	return apihelper.Validate(req, req.Post())
}

// Create project. If id is empty, a random one is created by the backend
func Create(ctx context.Context, id string) (project *Project, err error) {
	var req = apihelper.URL(ctx, "/projects")

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

// AddDomain in project
func AddDomain(ctx context.Context, projectID string, domain string) (err error) {
	var req = apihelper.URL(ctx, "/projects", url.QueryEscape(projectID), "/customDomains")

	apihelper.Auth(req)

	if err := apihelper.SetBody(req, domain); err != nil {
		return errwrap.Wrapf("Can not set body for domain: {{err}}", err)
	}
	return apihelper.Validate(req, req.Patch())
}

// RemoveDomain in project
func RemoveDomain(ctx context.Context, projectID string, domain string) (err error) {
	var req = apihelper.URL(ctx,
		"/projects",
		url.QueryEscape(projectID),
		"/customDomains",
		"/",
		url.QueryEscape(domain))

	apihelper.Auth(req)
	return apihelper.Validate(req, req.Delete())
}

// Get project by ID
func Get(ctx context.Context, id string) (project Project, err error) {
	err = apihelper.AuthGet(ctx, "/projects/"+url.QueryEscape(id), &project)
	return project, err
}

// List projects
func List(ctx context.Context) (list []Project, err error) {
	err = apihelper.AuthGet(ctx, "/projects", &list)
	return list, err
}

// Read a project directory properties (defined by a project.json on it)
func Read(path string) (*Project, error) {
	var content, err = ioutil.ReadFile(filepath.Join(path, "project.json"))
	var data Project

	if err != nil {
		return nil, readValidate(data, err)
	}

	if err = readValidate(data, json.Unmarshal(content, &data)); err != nil {
		return nil, err
	}

	err = checkRemovedCustomDomain(path, content)

	return &data, err
}

func checkRemovedCustomDomain(path string, content []byte) error {
	var mapProject map[string]interface{}

	if err := json.Unmarshal(content, &mapProject); err != nil {
		// silently bail
		return nil
	}

	var removedInterface, ok = mapProject["customDomain"]
	var removed string

	switch removedInterface.(type) {
	case nil:
		return nil
	case string:
		removed = removedInterface.(string)
	default:
		return fmt.Errorf("Invalid value for removed customDomain on %v", path)
	}

	if !ok || removed == "" {
		return nil
	}

	return fmt.Errorf(`CustomDomain string support was removed in favor of CustomDomains []string
Update your %v/project.json file to use:
"customDomains": ["%v"] instead of "customDomain": "%v".`,
		path,
		removed,
		removed)
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
func Restart(ctx context.Context, id string) error {
	var req = apihelper.URL(ctx, "/restart/project?projectId="+url.QueryEscape(id))

	apihelper.Auth(req)
	return apihelper.Validate(req, req.Post())
}

// Unlink project
func Unlink(ctx context.Context, projectID string) error {
	var req = apihelper.URL(ctx, "/projects", projectID)
	apihelper.Auth(req)

	return apihelper.Validate(req, req.Delete())
}

// Validate project
func Validate(ctx context.Context, projectID string) (err error) {
	var req = apihelper.URL(ctx, "/validators/project/id")
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
	project, err := Create(context.Background(), id)

	if err == nil {
		id = project.ID
		return id, nil
	}

	_, err = validateOrCreate(err)
	return id, err
}

// ValidateOrCreateFromJSON project
func ValidateOrCreateFromJSON(filename string) (created bool, err error) {
	return validateOrCreate(CreateFromJSON(context.Background(), filename))
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
