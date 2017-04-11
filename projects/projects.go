package projects

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"
	"sort"

	"github.com/hashicorp/errwrap"
	"github.com/wedeploy/api-go"
	"github.com/wedeploy/cli/apihelper"
	"github.com/wedeploy/cli/containers"
	"github.com/wedeploy/cli/verbosereq"
)

// Project structure
type Project struct {
	ProjectID     string   `json:"projectId"`
	CustomDomains []string `json:"customDomains,omitempty"`
	Health        string   `json:"health,omitempty"`
	Description   string   `json:"description,omitempty"`
	HealthUID     string   `json:"healthUid,omitempty"`
}

// ProjectPackage is the structure for project.json
type ProjectPackage struct {
	ID            string   `json:"id"`
	CustomDomains []string `json:"customDomains,omitempty"`
}

// Project returns a Project type created taking project.json as base
func (pp ProjectPackage) Project() Project {
	return Project{
		ProjectID:     pp.ID,
		CustomDomains: pp.CustomDomains,
	}
}

// Services of a given project
func (p *Project) Services(ctx context.Context) (containers.Containers, error) {
	return containers.List(ctx, p.ProjectID)
}

var (
	// ErrProjectNotFound happens when a project.json is not found
	ErrProjectNotFound = errors.New("Project not found")

	// ErrProjectAlreadyExists happens when a Project ID already exists
	ErrProjectAlreadyExists = errors.New("Project already exists")

	// ErrInvalidProjectID happens when a Project ID is invalid
	ErrInvalidProjectID = errors.New("Invalid project ID")

	// ErrEmptyProjectID happens when trying to access a project, but providing an empty ID
	ErrEmptyProjectID = errors.New("Can not get project: ID is empty")
)

type createRequestBody struct {
	ProjectID     string   `json:"projectId,omitempty"`
	CustomDomains []string `json:"customDomains,omitempty"`
}

// Create on the backend
func Create(ctx context.Context, project Project) (p Project, err error) {
	var req = apihelper.URL(ctx, "/projects")
	var reqBody = createRequestBody{
		ProjectID:     project.ProjectID,
		CustomDomains: project.CustomDomains,
	}

	apihelper.Auth(req)

	if err := apihelper.SetBody(req, reqBody); err != nil {
		return p, err
	}

	if err := apihelper.Validate(req, req.Post()); err != nil {
		return p, err
	}

	err = apihelper.DecodeJSON(req, &p)
	return p, err
}

type updateRequestBody struct {
	CustomDomains []string `json:"customDomains,omitempty"`
}

// Update project
func Update(ctx context.Context, project Project) (p Project, err error) {
	var req = apihelper.URL(ctx, "/projects", url.QueryEscape(project.ProjectID))
	var reqBody = updateRequestBody{
		CustomDomains: project.CustomDomains,
	}

	apihelper.Auth(req)

	if err := apihelper.SetBody(req, reqBody); err != nil {
		return p, err
	}

	if err := apihelper.Validate(req, req.Patch()); err != nil {
		return p, err
	}

	err = apihelper.DecodeJSON(req, &p)
	return p, err
}

// AddDomain in project
func AddDomain(ctx context.Context, projectID string, domain string) (err error) {
	var project, perr = Get(context.Background(), projectID)

	if perr != nil {
		return errwrap.Wrapf("Can not get current domains: {{err}}", perr)
	}

	var customDomains = project.CustomDomains
	customDomains = append(customDomains, domain)

	return updateDomains(ctx, projectID, customDomains)
}

// RemoveDomain in project
func RemoveDomain(ctx context.Context, projectID string, domain string) (err error) {
	var project, perr = Get(context.Background(), projectID)

	if perr != nil {
		return errwrap.Wrapf("Can not get current domains: {{err}}", perr)
	}

	var customDomains = []string{}

	for _, d := range project.CustomDomains {
		if domain != d {
			customDomains = append(customDomains, d)
		}
	}

	return updateDomains(ctx, projectID, customDomains)
}

func updateDomains(ctx context.Context, projectID string, domains []string) (err error) {
	var req = apihelper.URL(ctx, "/projects", url.QueryEscape(projectID), "/custom-domains")

	apihelper.Auth(req)

	if err := apihelper.SetBody(req, domains); err != nil {
		return errwrap.Wrapf("Can not set body for domain: {{err}}", err)
	}
	return apihelper.Validate(req, req.Put())
}

// Get project by ID
func Get(ctx context.Context, id string) (project Project, err error) {
	if id == "" {
		return project, ErrEmptyProjectID
	}

	err = apihelper.AuthGet(ctx, "/projects/"+url.QueryEscape(id), &project)
	return project, err
}

// List projects
func List(ctx context.Context) (list []Project, err error) {
	err = apihelper.AuthGet(ctx, "/projects", &list)
	sort.Slice(list, func(i, j int) bool {
		return list[i].ProjectID < list[j].ProjectID
	})
	return list, err
}

// Read a project directory properties (defined by a project.json on it)
func Read(path string) (*ProjectPackage, error) {
	var content, err = ioutil.ReadFile(filepath.Join(path, "project.json"))
	var data ProjectPackage

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

func readValidate(pp ProjectPackage, err error) error {
	switch {
	case os.IsNotExist(err):
		return ErrProjectNotFound
	case err != nil:
		return err
	case pp.ID == "":
		return ErrInvalidProjectID
	default:
		return err
	}
}

// Restart restarts a project
func Restart(ctx context.Context, id string) error {
	var req = apihelper.URL(ctx, "/projects/"+url.QueryEscape(id)+"/restart")

	apihelper.Auth(req)
	return apihelper.Validate(req, req.Post())
}

// Unlink project
func Unlink(ctx context.Context, projectID string) error {
	var req = apihelper.URL(ctx, "/projects", projectID)
	apihelper.Auth(req)

	return apihelper.Validate(req, req.Delete())
}

// CreateOrUpdate project
func CreateOrUpdate(ctx context.Context, project Project) (pRec Project, created bool, err error) {
	var _, errExisting = Get(ctx, project.ProjectID)

	if errExisting == nil {
		var pUpdated, errUpdate = Update(ctx, project)
		return pUpdated, false, errUpdate
	}

	pRec, err = Create(ctx, project)

	if err == nil {
		created = true
	}

	return pRec, created, err
}

func doValidate(projectID string, req *wedeploy.WeDeploy) error {
	apihelper.Auth(req)

	req.Param("value", projectID)

	var err = req.Get()

	verbosereq.Feedback(req)
	return err
}
