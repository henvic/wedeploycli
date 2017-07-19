package projects

import (
	"context"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"
	"sort"

	"github.com/wedeploy/api-go"
	"github.com/wedeploy/cli/apihelper"
	"github.com/wedeploy/cli/services"
	"github.com/wedeploy/cli/verbosereq"
)

// Project structure
type Project struct {
	ProjectID   string `json:"projectId"`
	Health      string `json:"health,omitempty"`
	Description string `json:"description,omitempty"`
	HealthUID   string `json:"healthUid,omitempty"`
}

// ProjectPackage is the structure for project.json
type ProjectPackage struct {
	ID string `json:"id"`
}

// Project returns a Project type created taking project.json as base
func (pp ProjectPackage) Project() Project {
	return Project{
		ProjectID: pp.ID,
	}
}

// Services of a given project
func (p *Project) Services(ctx context.Context) (services.Services, error) {
	return services.List(ctx, p.ProjectID)
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
	ProjectID string `json:"projectId,omitempty"`
}

// Create on the backend
func Create(ctx context.Context, project Project) (p Project, err error) {
	var req = apihelper.URL(ctx, "/projects")
	var reqBody = createRequestBody{
		ProjectID: project.ProjectID,
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

type updateRequestBody struct{}

// Update project
func Update(ctx context.Context, project Project) (p Project, err error) {
	var req = apihelper.URL(ctx, "/projects", url.QueryEscape(project.ProjectID))
	var reqBody = updateRequestBody{}

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

	return &data, err
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
