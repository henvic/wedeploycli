package projects

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/hashicorp/errwrap"
	"github.com/henvic/wedeploycli/apihelper"
	"github.com/henvic/wedeploycli/config"
	"github.com/henvic/wedeploycli/services"
)

// Client for the projects
type Client struct {
	*apihelper.Client
}

// New Client
func New(wectx config.Context) *Client {
	return &Client{
		apihelper.New(wectx),
	}
}

// Region of the project
type Region struct {
	Location string `json:"location"`
	Name     string `json:"name"`
}

// Project structure
type Project struct {
	ProjectID   string `json:"projectId"`
	Region      string `json:"cluster"`
	Health      string `json:"health,omitempty"`
	Description string `json:"description,omitempty"`
	CreatedAt   int64  `json:"createdAt,omitempty"`

	Services services.Services `json:"services,omitempty"`
}

// CreatedAtTime extracts from the Unix timestamp format and returns the createdAt value
func (p *Project) CreatedAtTime() time.Time {
	var createdAt = p.CreatedAt / 1000
	return time.Unix(createdAt, 0)
}

var (
	// ErrProjectAlreadyExists happens when a Project ID already exists
	ErrProjectAlreadyExists = errors.New("project already exists")

	// ErrInvalidProjectID happens when a Project ID is invalid
	ErrInvalidProjectID = errors.New("invalid project ID")

	// ErrEmptyProjectID happens when trying to access a project, but providing an empty ID
	ErrEmptyProjectID = errors.New("can't get project: ID is empty")
)

// Regions available for the project
func (c *Client) Regions(ctx context.Context) (list []Region, err error) {
	err = c.Client.AuthGet(ctx, "/clusters", &list)
	sort.Slice(list, func(i, j int) bool {
		return list[i].Name < list[j].Name
	})
	return list, err
}

type createRequestBody struct {
	ProjectID   string `json:"projectId,omitempty"`
	Cluster     string `json:"cluster,omitempty"`
	Environment bool   `json:"environment,omitempty"`
}

// Create on the backend
func (c *Client) Create(ctx context.Context, project Project) (p Project, err error) {
	var req = c.Client.URL(ctx, "/projects")
	var reqBody = createRequestBody{
		ProjectID:   project.ProjectID,
		Cluster:     project.Region,
		Environment: strings.Contains(project.ProjectID, "-"),
	}

	c.Client.Auth(req)

	if err = apihelper.SetBody(req, reqBody); err != nil {
		return p, err
	}

	if err = apihelper.Validate(req, req.Post()); err != nil {
		return p, err
	}

	err = apihelper.DecodeJSON(req, &p)
	return p, err
}

type updateRequestBody struct{}

// Update project
func (c *Client) Update(ctx context.Context, project Project) (p Project, err error) {
	var req = c.Client.URL(ctx, "/projects", url.PathEscape(project.ProjectID))
	var reqBody = updateRequestBody{}

	c.Client.Auth(req)

	if err = apihelper.SetBody(req, reqBody); err != nil {
		return p, err
	}

	if err = apihelper.Validate(req, req.Patch()); err != nil {
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

	err = c.Client.AuthGet(ctx, "/projects/"+url.PathEscape(id), &project)
	return project, err
}

// GetWithServices project by ID with a list of its services
func (c *Client) GetWithServices(ctx context.Context, id string) (project Project, err error) {
	if id == "" {
		return project, ErrEmptyProjectID
	}

	err = c.Client.AuthGet(ctx,
		fmt.Sprintf("/projects/%s?listServices=true", url.PathEscape(id)),
		&project)
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

// ListWithServices projects with a list of its services
func (c *Client) ListWithServices(ctx context.Context) (list []Project, err error) {
	err = c.Client.AuthGet(ctx, "/projects?listServices=true", &list)
	sort.Slice(list, func(i, j int) bool {
		return list[i].ProjectID < list[j].ProjectID
	})
	return list, err
}

// Delete project
func (c *Client) Delete(ctx context.Context, projectID string) error {
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

// BuildRequestBody structure
type BuildRequestBody struct {
	Repository string `json:"repository,omitempty"`

	Deploy bool `json:"deploy"`
}

// BuildResponseBody structure
type BuildResponseBody struct {
	ServiceID string `json:"serviceID"`
	GroupUID  string `json:"groupUid"`
}

// Build project
func (c *Client) Build(ctx context.Context, projectID string, build BuildRequestBody) (groupUID string, builds []BuildResponseBody, err error) {
	var req = c.Client.URL(ctx, "/projects", projectID, "/build")
	c.Client.Auth(req)

	if err = apihelper.SetBody(req, build); err != nil {
		return "", builds, errwrap.Wrapf("can't set body for build: {{err}}", err)
	}

	if err = apihelper.Validate(req, req.Post()); err != nil {
		return "", builds, err
	}

	if err = apihelper.DecodeJSON(req, &builds); err != nil {
		return "", builds, err
	}

	if len(builds) > 0 {
		groupUID = builds[0].GroupUID
	}

	return groupUID, builds, err
}

// GetDeploymentOrder gets the order of a given deployment
func (c *Client) GetDeploymentOrder(ctx context.Context, projectID, groupUID string) (order []string, err error) {
	var addr = fmt.Sprintf("/projects/%s/builds/order/%s",
		url.PathEscape(projectID),
		url.PathEscape(groupUID))
	err = c.Client.AuthGet(ctx, addr, &order)
	return order, err
}

// GetBuilds on a given project.
func (c *Client) GetBuilds(ctx context.Context, projectID, groupUID string) (bs []Build, err error) {
	var addr = fmt.Sprintf("/projects/%s/builds?buildGroupUid=%s",
		url.PathEscape(projectID),
		url.QueryEscape(groupUID))
	err = c.Client.AuthGet(ctx, addr, &bs)
	return bs, err
}

// Build information.
type Build struct {
	ProjectID string
	ServiceID string

	GroupUID      string
	BuildGroupUID string

	Environments map[string]json.RawMessage
	Metadata     map[string]interface{}
	Wedeploy     map[string]interface{}
}

// SkippedDeploy returns true when the service is not configured to be deployed.
// TODO(henvic): these rules should actually be moved to the server-side.
// The code below is explicitly repetitive/verbose to clarify how hard it is to
// currently guess whether the service is to be built or deployed.
func (b *Build) SkippedDeploy() bool {
	if deploy, ok := b.Metadata["deploy"].(bool); ok && !deploy {
		return true
	}

	ep := strings.Split(b.ProjectID, "-")
	environment := ep[len(ep)-1]

	if environment == "" {
		if deploy, ok := b.Wedeploy["deploy"].(bool); ok && !deploy {
			return true
		}
	}

	ve := b.Environments[environment]
	var m map[string]interface{}
	var err = json.Unmarshal(ve, &m)

	if err != nil || m == nil {
		if deploy, ok := b.Wedeploy["deploy"].(bool); ok && !deploy {
			return true
		}

		return false
	}

	if deploy, ok := m["deploy"].(bool); ok {
		return ok && !deploy
	}

	if deploy, ok := b.Wedeploy["deploy"].(bool); ok && !deploy {
		return true
	}

	return false
}
