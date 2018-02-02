package services

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
	"strings"

	"github.com/hashicorp/errwrap"
	"github.com/wedeploy/cli/apihelper"
	"github.com/wedeploy/cli/config"
	"github.com/wedeploy/cli/defaults"
	"github.com/wedeploy/cli/verbose"
	"github.com/wedeploy/cli/verbosereq"
	"github.com/wedeploy/wedeploy-sdk-go"
)

// Client for the services
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

// Services of services for helper functions
type Services []Service

// CreateBody is the body for creating a service
type CreateBody struct {
	CPU           json.Number       `json:"cpu"`
	CustomDomains []string          `json:"customDomains,omitempty"`
	Env           map[string]string `json:"env,omitempty"`
	Image         string            `json:"image,omitempty"`
	Memory        json.Number       `json:"memory,omitempty"`
	Port          int               `json:"port,omitempty"`
	Scale         int               `json:"scale,omitempty"`
	ServiceID     string            `json:"serviceId,omitempty"`
	Volume        string            `json:"volume,omitempty"`
}

// Create on the backend
func (c *Client) Create(ctx context.Context, projectID string, body CreateBody) (s Service, err error) {
	var req = c.Client.URL(ctx, "/projects/"+projectID+"/services")

	c.Client.Auth(req)

	if err := apihelper.SetBody(req, body); err != nil {
		return s, err
	}

	if err := apihelper.Validate(req, req.Post()); err != nil {
		return s, err
	}

	err = apihelper.DecodeJSON(req, &s)
	return s, err
}

// Get a service from the service list
func (cs Services) Get(id string) (c Service, err error) {
	for _, c := range cs {
		if c.ServiceID == id {
			return c, nil
		}
	}

	return c, errors.New("No service found")
}

// Service structure
type Service struct {
	ServiceID     string            `json:"serviceId,omitempty"`
	Health        string            `json:"health,omitempty"`
	Image         string            `json:"image,omitempty"`
	ImageHint     string            `json:"imageHint,omitempty"`
	CustomDomains []string          `json:"customDomains,omitempty"`
	Env           map[string]string `json:"env,omitempty"`
	Scale         int               `json:"scale,omitempty"`
	CPU           json.Number       `json:"cpu,omitempty"`
	Memory        json.Number       `json:"memory,omitempty"`
}

// Type returns the image hint or the image
func (s *Service) Type() string {
	if s.ImageHint != "" {
		return s.ImageHint
	}

	return s.Image
}

// ServicePackage is the structure for wedeploy.json
type ServicePackage struct {
	ID            string            `json:"id,omitempty"`
	Scale         int               `json:"scale,omitempty"`
	Image         string            `json:"image,omitempty"`
	CustomDomains []string          `json:"customDomains,omitempty"`
	Env           map[string]string `json:"env,omitempty"`
	Dependencies  []string          `json:"dependencies,omitempty"`
	dockerfile    string
}

// Service returns a Service type created taking wedeploy.json as base
func (cp ServicePackage) Service() *Service {
	return &Service{
		ServiceID:     cp.ID,
		Scale:         cp.Scale,
		Image:         cp.Image,
		CustomDomains: cp.CustomDomains,
		Env:           cp.Env,
	}
}

// Register for the service structure
type Register struct {
	ID          string `json:"id"`
	Image       string `json:"image"`
	Category    string `json:"category"`
	Description string `json:"description"`
}

var (
	// ErrServiceNotFound happens when a wedeploy.json is not found
	ErrServiceNotFound = errors.New("Service not found")

	// ErrServiceAlreadyExists happens when a service ID already exists
	ErrServiceAlreadyExists = errors.New("Service already exists")

	// ErrInvalidServiceID happens when a service ID is invalid
	ErrInvalidServiceID = errors.New("Invalid service ID")

	// ErrEmptyProjectID happens when trying to access a project, but providing an empty ID
	ErrEmptyProjectID = errors.New("Can not get project: Project ID is empty")

	// ErrEmptyServiceID happens when trying to access a service, but providing an empty ID
	ErrEmptyServiceID = errors.New("Can not get service: Service ID is empty")

	// ErrEmptyProjectAndServiceID happens when trying to access a service, but providing empty IDs
	ErrEmptyProjectAndServiceID = errors.New("Can not get service: Project and Service ID is empty")
)

// ServiceInfo is for a tuple of service ID and Location.
type ServiceInfo struct {
	ServiceID string
	Location  string
}

// ServiceInfoList is a list of ServiceInfo
type ServiceInfoList []ServiceInfo

// GetLocations returns the locations of a given ServiceInfoList.
func (c ServiceInfoList) GetLocations() []string {
	var locations = []string{}

	for _, ci := range c {
		locations = append(locations, ci.Location)
	}

	return locations
}

// GetIDs returns the services ids of a given ServiceInfoList.
func (c ServiceInfoList) GetIDs() []string {
	var ids = []string{}

	for _, ci := range c {
		ids = append(ids, ci.ServiceID)
	}

	return ids
}

// Get service info from service info list
func (c ServiceInfoList) Get(ID string) (ServiceInfo, error) {
	for _, item := range c {
		if ID == item.ServiceID {
			return item, nil
		}
	}

	return ServiceInfo{}, fmt.Errorf("found no service matching ID %v locally", ID)
}

// GetListFromDirectory returns a list of services on the given diretory
func GetListFromDirectory(root string) (ServiceInfoList, error) {
	return (&listFromDirectoryGetter{}).Walk(root)
}

type listFromDirectoryGetter struct {
	list ServiceInfoList
	root string
}

func (l *listFromDirectoryGetter) Walk(root string) (ServiceInfoList, error) {
	var err error

	if len(root) != 0 {
		if l.root, err = filepath.Abs(root); err != nil {
			return nil, err
		}
	}

	l.list = ServiceInfoList{}

	info, err := os.Stat(l.root)

	if err != nil {
		return nil, err
	}

	if err := l.readDir(l.root, info); err != nil {
		return nil, err
	}

	files, err := ioutil.ReadDir(l.root)

	if err != nil {
		return nil, err
	}

	for _, info := range files {
		path := filepath.Join(l.root, info.Name())
		if err := l.readDir(path, info); err != nil {
			return nil, err
		}
	}

	return l.list, nil
}

func (l *listFromDirectoryGetter) readDir(path string, info os.FileInfo) error {
	if strings.HasPrefix(info.Name(), ".") {
		return nil
	}

	if !info.IsDir() {
		return nil
	}

	_, noServiceErr := os.Stat(filepath.Join(path, ".noservice"))

	switch {
	case os.IsNotExist(noServiceErr):
	case noServiceErr == nil:
		return nil
	default:
		return noServiceErr
	}

	return l.readFunc(path)
}

func (l *listFromDirectoryGetter) readFunc(dir string) error {
	switch service, errRead := Read(dir); {
	case errRead == nil:
		return l.addFunc(service, dir)
	case errRead == ErrServiceNotFound:
		return nil
	default:
		return errwrap.Wrapf("can't list services: {{err}}", errRead)
	}
}

func (l *listFromDirectoryGetter) checkExisting(sp *ServicePackage, dir string) error {
	const errCheck = `found services with duplicated ID "%v" on %v and %v`
	if sp.ID == "" {
		return nil
	}

	for _, existing := range l.list {
		if sp.ID == existing.ServiceID {
			return fmt.Errorf(errCheck, sp.ID, existing.Location, dir)
		}
	}

	return nil
}

func (l *listFromDirectoryGetter) addFunc(sp *ServicePackage, dir string) error {
	if err := l.checkExisting(sp, dir); err != nil {
		return err
	}

	l.list = append(l.list, ServiceInfo{
		ServiceID: sp.ID,
		Location:  dir,
	})

	return nil
}

// CatalogItem is a item on the WeDeploy services registry
type CatalogItem struct {
	Category    string   `json:"category"`
	Description string   `json:"description"`
	Image       string   `json:"image"`
	Name        string   `json:"name"`
	State       string   `json:"state"`
	Versions    []string `json:"versions"`
}

// Catalog of services
func (c *Client) Catalog(ctx context.Context) (catalog map[string]CatalogItem, err error) {
	err = c.Client.AuthGet(ctx, "/catalog/services", &catalog)
	return catalog, err
}

// List services of a given project
func (c *Client) List(ctx context.Context, projectID string) (Services, error) {
	var cs Services

	var err = c.Client.AuthGet(ctx, "/projects/"+url.QueryEscape(projectID)+"/services", &cs)
	sort.Slice(cs, func(i, j int) bool {
		return cs[i].ServiceID < cs[j].ServiceID
	})
	return cs, err
}

// Get service
func (c *Client) Get(ctx context.Context, projectID, serviceID string) (s Service, err error) {
	switch {
	case projectID == "" && serviceID == "":
		return s, ErrEmptyProjectAndServiceID
	case projectID == "":
		return s, ErrEmptyProjectID
	case serviceID == "":
		return s, ErrEmptyServiceID
	}

	err = c.Client.AuthGet(ctx, "/projects/"+
		url.QueryEscape(projectID)+
		"/services/"+
		url.QueryEscape(serviceID), &s)
	return s, err
}

// AddDomain in project
func (c *Client) AddDomain(ctx context.Context, projectID, serviceID string, domain string) (err error) {
	var service, perr = c.Get(ctx, projectID, serviceID)

	if perr != nil {
		return errwrap.Wrapf("Can not get current domains: {{err}}", perr)
	}

	var customDomains = maybeAppendDomain(service.CustomDomains, domain)
	return c.updateDomains(ctx, projectID, serviceID, customDomains)
}

func maybeAppendDomain(customDomains []string, domain string) []string {
	for _, k := range customDomains {
		if k == domain {
			return customDomains
		}
	}

	customDomains = append(customDomains, domain)
	return customDomains
}

// RemoveDomain in project
func (c *Client) RemoveDomain(ctx context.Context, projectID string, serviceID, domain string) (err error) {
	var service, perr = c.Get(ctx, projectID, serviceID)

	if perr != nil {
		return errwrap.Wrapf("Can not get current domains: {{err}}", perr)
	}

	var customDomains = []string{}

	for _, d := range service.CustomDomains {
		if domain != d {
			customDomains = append(customDomains, d)
		}
	}

	return c.updateDomains(ctx, projectID, serviceID, customDomains)
}

type updateDomainsReq struct {
	Value []string `json:"value"`
}

func (c *Client) updateDomains(ctx context.Context, projectID, serviceID string, domains []string) (err error) {
	var req = c.Client.URL(ctx, "/projects",
		url.QueryEscape(projectID),
		"/services",
		url.QueryEscape(serviceID),
		"/custom-domains")

	c.Client.Auth(req)

	if err := apihelper.SetBody(req, updateDomainsReq{domains}); err != nil {
		return errwrap.Wrapf("Can not set body for domain: {{err}}", err)
	}
	return apihelper.Validate(req, req.Put())
}

// EnvironmentVariable of a service
type EnvironmentVariable struct {
	Name  string `json:"name,omitempty"`
	Value string `json:"value,omitempty"`
}

// GetEnvironmentVariables of a service
func (c *Client) GetEnvironmentVariables(ctx context.Context, projectID, serviceID string) (envs []EnvironmentVariable, err error) {
	err = c.Client.AuthGet(ctx, "/projects/"+
		url.QueryEscape(projectID)+
		"/services/"+
		url.QueryEscape(serviceID)+
		"/environment-variables", &envs)
	return envs, err
}

type linkRequestBody struct {
	ServiceID string            `json:"serviceId,omitempty"`
	Image     string            `json:"image,omitempty"`
	Scale     int               `json:"scale,omitempty"`
	Env       map[string]string `json:"env,omitempty"`
	Version   string            `json:"version,omitempty"`
	Source    string            `json:"source,omitempty"`
}

// Link service to project
func (c *Client) Link(ctx context.Context, projectID string, service Service, source string) (err error) {
	var reqBody = linkRequestBody{
		ServiceID: service.ServiceID,
		Image:     service.Image,
		Scale:     service.Scale,
		Env:       service.Env,
		Source:    source,
	}

	if reqBody.Scale == 0 {
		reqBody.Scale = 1
	}

	verbose.Debug("Linking service " + service.ServiceID + " to project " + projectID)

	var req = c.Client.URL(ctx, "/projects", url.QueryEscape(projectID), "/services")
	c.Client.Auth(req)

	err = apihelper.SetBody(req, reqBody)

	if err != nil {
		return err
	}

	return apihelper.Validate(req, req.Post())
}

// Unlink service
func (c *Client) Unlink(ctx context.Context, projectID, serviceID string) error {
	var req = c.Client.URL(ctx,
		"/projects",
		url.QueryEscape(projectID),
		"/services",
		url.QueryEscape(serviceID))
	c.Client.Auth(req)

	return apihelper.Validate(req, req.Delete())
}

// GetRegistry gets a list of service images
func GetRegistry(ctx context.Context) (registry []Register, err error) {
	var request = wedeploy.URL(defaults.Hub, "/registry")
	request.SetContext(ctx)

	err = apihelper.Validate(request, request.Get())

	if err != nil {
		return nil, err
	}

	err = apihelper.DecodeJSON(request, &registry)

	return registry, err
}

// Read a service directory properties (defined by a wedeploy.json and/or Dockerfile on it)
func Read(path string) (*ServicePackage, error) {
	var (
		data          = ServicePackage{}
		hasDockerfile bool
	)

	dockerfile, err := ioutil.ReadFile(filepath.Join(path, "Dockerfile"))

	switch {
	case err == nil:
		hasDockerfile = true
	case !os.IsNotExist(err):
		return nil, errwrap.Wrapf("error reading Dockerfile: {{err}}", err)
	}

	wedeployJSON, err := ioutil.ReadFile(filepath.Join(path, "wedeploy.json"))

	switch {
	case err == nil:
		if err = json.Unmarshal(wedeployJSON, &data); err != nil {
			return nil, errwrap.Wrapf("error parsing wedeploy.json on "+path+": {{err}}", err)
		}
	case os.IsNotExist(err):
		if !hasDockerfile {
			return nil, ErrServiceNotFound
		}
	default:
		return nil, errwrap.Wrapf("error reading wedeploy.json: {{err}}", err)
	}

	if hasDockerfile {
		data.dockerfile = string(dockerfile)
	}

	return &data, nil
}

// SetEnvironmentVariable sets an environment variable
func (c *Client) SetEnvironmentVariable(ctx context.Context, projectID, serviceID, key, value string) error {
	var req = c.Client.URL(ctx,
		"/projects",
		url.QueryEscape(projectID),
		"/services",
		url.QueryEscape(serviceID),
		"/environment-variables/"+
			url.QueryEscape(key))

	c.Client.Auth(req)

	b := map[string]string{
		"value": value,
	}

	if err := apihelper.SetBody(req, b); err != nil {
		return errwrap.Wrapf("Can not set body for setting environment variable: {{err}}", err)
	}

	return apihelper.Validate(req, req.Put())
}

// UnsetEnvironmentVariable removes an environment variable
func (c *Client) UnsetEnvironmentVariable(ctx context.Context, projectID, serviceID, key string) error {
	var req = c.Client.URL(ctx,
		"/projects",
		url.QueryEscape(projectID),
		"/services",
		url.QueryEscape(serviceID),
		"/environment-variables/"+
			url.QueryEscape(key))

	c.Client.Auth(req)

	return apihelper.Validate(req, req.Delete())
}

// Scale of the service
type Scale struct {
	Current int `json:"value"`
}

// Scale sets the scale for a given service
func (c *Client) Scale(ctx context.Context, projectID, serviceID string, s Scale) (err error) {
	var req = c.Client.URL(ctx,
		"/projects",
		url.QueryEscape(projectID),
		"/services",
		url.QueryEscape(serviceID),
		"/scale")
	c.Client.Auth(req)

	err = apihelper.SetBody(req, &s)

	if err != nil {
		return err
	}

	return apihelper.Validate(req, req.Patch())
}

// Restart restarts a service inside a project
func (c *Client) Restart(ctx context.Context, projectID, serviceID string) error {
	var req = c.Client.URL(ctx, "/projects/"+
		url.QueryEscape(projectID)+
		"/services/"+
		url.QueryEscape(serviceID)+
		"/restart")

	c.Client.Auth(req)
	return apihelper.Validate(req, req.Post())
}

// Validate service
func (c *Client) Validate(ctx context.Context, projectID, serviceID string) (err error) {
	var req = c.Client.URL(ctx, "/validators/services/id")
	err = c.doValidate(projectID, serviceID, req)

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

func (c *Client) doValidate(projectID, serviceID string, req *wedeploy.WeDeploy) error {
	c.Client.Auth(req)

	req.Param("projectId", projectID)
	req.Param("value", serviceID)

	var err = req.Get()

	verbosereq.Feedback(req)
	return err
}

func getValidateAPIFaultError(errDoc apihelper.APIFault) error {
	switch {
	case errDoc.Has("invalidServiceId"):
		return ErrInvalidServiceID
	case errDoc.Has("serviceAlreadyExists"):
		return ErrServiceAlreadyExists
	}

	return errDoc
}

// normalizePathToUnix is a filter for normalizing Windows paths to Unix-style
func normalizePathToUnix(s string) string {
	var ps = strings.SplitN(s, ":\\", 2)

	if len(ps) == 1 {
		return s
	}

	return "/" + strings.Replace(
		filepath.Join(ps[0], ps[1]), "\\", "/", -1)
}
