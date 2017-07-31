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
	"github.com/wedeploy/api-go"
	"github.com/wedeploy/cli/apihelper"
	"github.com/wedeploy/cli/defaults"
	"github.com/wedeploy/cli/verbose"
	"github.com/wedeploy/cli/verbosereq"
)

// Services list
type Services []Service

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
	CustomDomains []string          `json:"customDomains,omitempty"`
	Env           map[string]string `json:"env,omitempty"`
	Scale         int               `json:"scale,omitempty"`
	HealthUID     string            `json:"healthUid,omitempty"`
}

// ServicePackage is the structure for wedeploy.json
type ServicePackage struct {
	ID            string            `json:"id,omitempty"`
	Scale         int               `json:"scale,omitempty"`
	Image         string            `json:"image,omitempty"`
	CustomDomains []string          `json:"customDomains,omitempty"`
	Env           map[string]string `json:"env,omitempty"`
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
	list     ServiceInfoList
	idDirMap map[string]string
	root     string
}

func (l *listFromDirectoryGetter) Walk(root string) (ServiceInfoList, error) {
	var err error

	if len(root) != 0 {
		if l.root, err = filepath.Abs(root); err != nil {
			return nil, err
		}
	}

	l.list = ServiceInfoList{}
	l.idDirMap = map[string]string{}

	if err = filepath.Walk(l.root, l.walkFunc); err != nil {
		return nil, err
	}

	return l.list, nil
}

func (l *listFromDirectoryGetter) walkFunc(path string, info os.FileInfo, err error) error {
	if err != nil {
		return err
	}

	if !info.IsDir() {
		return nil
	}

	_, noServiceErr := os.Stat(filepath.Join(path, ".noservice"))

	switch {
	case os.IsNotExist(noServiceErr):
	case noServiceErr == nil:
		return filepath.SkipDir
	default:
		return noServiceErr
	}

	return l.readFunc(path)
}

func (l *listFromDirectoryGetter) readFunc(dir string) error {
	var wedeployFile = filepath.Join(dir, "wedeploy.json")
	var service, errRead = Read(dir)

	switch {
	case errRead == nil:
		return l.addFunc(service, dir)
	case errRead == ErrServiceNotFound:
		return nil
	default:
		return errwrap.Wrapf("Can not list services: error reading "+wedeployFile+": {{err}}", errRead)
	}
}

func (l *listFromDirectoryGetter) addFunc(cp *ServicePackage, dir string) error {
	if cpv, ok := l.idDirMap[cp.ID]; ok {
		return fmt.Errorf(`Can not list services: ID "%v" was found duplicated on services %v and %v`,
			cp.ID,
			cpv,
			dir)
	}

	l.idDirMap[cp.ID] = dir
	l.list = append(l.list, ServiceInfo{
		ServiceID: cp.ID,
		Location:  strings.TrimPrefix(dir, l.root+string(os.PathSeparator)),
	})

	return filepath.SkipDir
}

// List services of a given project
func List(ctx context.Context, projectID string) (Services, error) {
	var cs Services

	var err = apihelper.AuthGet(ctx, "/projects/"+url.QueryEscape(projectID)+"/services", &cs)
	sort.Slice(cs, func(i, j int) bool {
		return cs[i].ServiceID < cs[j].ServiceID
	})
	return cs, err
}

// Get service
func Get(ctx context.Context, projectID, serviceID string) (c Service, err error) {
	switch {
	case projectID == "" && serviceID == "":
		return c, ErrEmptyProjectAndServiceID
	case projectID == "":
		return c, ErrEmptyProjectID
	case serviceID == "":
		return c, ErrEmptyServiceID
	}

	err = apihelper.AuthGet(ctx, "/projects/"+
		url.QueryEscape(projectID)+
		"/services/"+
		url.QueryEscape(serviceID), &c)
	return c, err
}

// AddDomain in project
func AddDomain(ctx context.Context, projectID, serviceID string, domain string) (err error) {
	var service, perr = Get(context.Background(), projectID, serviceID)

	if perr != nil {
		return errwrap.Wrapf("Can not get current domains: {{err}}", perr)
	}

	var customDomains = maybeAppendDomain(service.CustomDomains, domain)
	return updateDomains(ctx, projectID, serviceID, customDomains)
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
func RemoveDomain(ctx context.Context, projectID string, serviceID, domain string) (err error) {
	var service, perr = Get(context.Background(), projectID, serviceID)

	if perr != nil {
		return errwrap.Wrapf("Can not get current domains: {{err}}", perr)
	}

	var customDomains = []string{}

	for _, d := range service.CustomDomains {
		if domain != d {
			customDomains = append(customDomains, d)
		}
	}

	return updateDomains(ctx, projectID, serviceID, customDomains)
}

type updateDomainsReq struct {
	Value []string `json:"value"`
}

func updateDomains(ctx context.Context, projectID, serviceID string, domains []string) (err error) {
	var req = apihelper.URL(ctx, "/projects",
		url.QueryEscape(projectID),
		"/services",
		url.QueryEscape(serviceID),
		"/custom-domains")

	apihelper.Auth(req)

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
func GetEnvironmentVariables(ctx context.Context, projectID, serviceID string) (envs []EnvironmentVariable, err error) {
	err = apihelper.AuthGet(ctx, "/projects/"+
		url.QueryEscape(projectID)+
		"/services/"+
		url.QueryEscape(serviceID)+
		"/environment-variables", &envs)
	return envs, err
}

type linkRequestBody struct {
	ServiceID string            `json:"serviceId,omitempty"`
	Image     string            `json:"image,omitempty"`
	Port      string            `json:"port,omitempty"`
	Scale     int               `json:"scale,omitempty"`
	Env       map[string]string `json:"env,omitempty"`
	Version   string            `json:"version,omitempty"`
	Source    string            `json:"source,omitempty"`
}

// Link service to project
func Link(ctx context.Context, projectID string, service Service, source string) (err error) {
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

	var req = apihelper.URL(ctx, "/projects", url.QueryEscape(projectID), "/services")
	apihelper.Auth(req)

	err = apihelper.SetBody(req, reqBody)

	if err != nil {
		return err
	}

	return apihelper.Validate(req, req.Post())
}

// Unlink service
func Unlink(ctx context.Context, projectID, serviceID string) error {
	var req = apihelper.URL(ctx,
		"/projects",
		url.QueryEscape(projectID),
		"/services",
		url.QueryEscape(serviceID))
	apihelper.Auth(req)

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

// Read a service directory properties (defined by a wedeploy.json on it)
func Read(path string) (*ServicePackage, error) {
	var content, err = ioutil.ReadFile(filepath.Join(path, "wedeploy.json"))
	var data ServicePackage

	if err != nil {
		return nil, readValidate(data, err)
	}

	err = json.Unmarshal(content, &data)

	return &data, readValidate(data, err)
}

func readValidate(service ServicePackage, err error) error {
	switch {
	case os.IsNotExist(err):
		return ErrServiceNotFound
	case err != nil:
		return err
	case service.ID == "":
		return ErrInvalidServiceID
	default:
		return err
	}
}

// SetEnvironmentVariable sets an environment variable
func SetEnvironmentVariable(ctx context.Context, projectID, serviceID, key, value string) error {
	var req = apihelper.URL(ctx,
		"/projects",
		url.QueryEscape(projectID),
		"/services",
		url.QueryEscape(serviceID),
		"/environment-variables/"+
			url.QueryEscape(key))

	apihelper.Auth(req)

	b := map[string]string{
		"value": value,
	}

	if err := apihelper.SetBody(req, b); err != nil {
		return errwrap.Wrapf("Can not set body for setting environment variable: {{err}}", err)
	}

	return apihelper.Validate(req, req.Put())
}

// UnsetEnvironmentVariable removes an environment variable
func UnsetEnvironmentVariable(ctx context.Context, projectID, serviceID, key string) error {
	var req = apihelper.URL(ctx,
		"/projects",
		url.QueryEscape(projectID),
		"/services",
		url.QueryEscape(serviceID),
		"/environment-variables/"+
			url.QueryEscape(key))

	apihelper.Auth(req)

	return apihelper.Validate(req, req.Delete())
}

// Restart restarts a service inside a project
func Restart(ctx context.Context, projectID, serviceID string) error {
	var req = apihelper.URL(ctx, "/projects/"+
		url.QueryEscape(projectID)+
		"/services/"+
		url.QueryEscape(serviceID)+
		"/restart")

	apihelper.Auth(req)
	return apihelper.Validate(req, req.Post())
}

// Validate service
func Validate(ctx context.Context, projectID, serviceID string) (err error) {
	var req = apihelper.URL(ctx, "/validators/services/id")
	err = doValidate(projectID, serviceID, req)

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

func doValidate(projectID, serviceID string, req *wedeploy.WeDeploy) error {
	apihelper.Auth(req)

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