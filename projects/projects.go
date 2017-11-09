package projects

import (
	"context"
	"errors"
	"net/url"
	"sort"

	"github.com/wedeploy/cli/apihelper"
	"github.com/wedeploy/cli/config"
	"github.com/wedeploy/cli/services"
)

// Client for the projects
type Client struct {
	*apihelper.Client
}

// New Client
func New(wectx config.Context) *Client {
	return &Client{
		&apihelper.Client{
			Context: wectx,
		},
	}
}

// Project structure
type Project struct {
	ProjectID   string `json:"projectId"`
	Health      string `json:"health,omitempty"`
	Description string `json:"description,omitempty"`
}

// Services of a given project
func (p *Project) Services(ctx context.Context, s *services.Client) (services.Services, error) {
	return s.List(ctx, p.ProjectID)
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
func (c *Client) Create(ctx context.Context, project Project) (p Project, err error) {
	var req = c.Client.URL(ctx, "/projects")
	var reqBody = createRequestBody{
		ProjectID: project.ProjectID,
	}

	c.Client.Auth(req)

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
func (c *Client) Update(ctx context.Context, project Project) (p Project, err error) {
	var req = c.Client.URL(ctx, "/projects", url.QueryEscape(project.ProjectID))
	var reqBody = updateRequestBody{}

	c.Client.Auth(req)

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
func (c *Client) Get(ctx context.Context, id string) (project Project, err error) {
	if id == "" {
		return project, ErrEmptyProjectID
	}

	err = c.Client.AuthGet(ctx, "/projects/"+url.QueryEscape(id), &project)
	return project, err
}

// List projects
func (c *Client) List(ctx context.Context) (list []Project, err error) {
	err = c.Client.AuthGet(ctx, "/projects", &list)
	sort.Slice(list, func(i, j int) bool {
		return list[i].ProjectID < list[j].ProjectID
	})
	return list, err
}

// Unlink project
func (c *Client) Unlink(ctx context.Context, projectID string) error {
	var req = c.Client.URL(ctx, "/projects", projectID)
	c.Client.Auth(req)

	return apihelper.Validate(req, req.Delete())
}

// CreateOrUpdate project
func (c *Client) CreateOrUpdate(ctx context.Context, project Project) (pRec Project, created bool, err error) {
	var _, errExisting = c.Get(ctx, project.ProjectID)

	if errExisting == nil {
		var pUpdated, errUpdate = c.Update(ctx, project)
		return pUpdated, false, errUpdate
	}

	pRec, err = c.Create(ctx, project)

	if err == nil {
		created = true
	}

	return pRec, created, err
}
