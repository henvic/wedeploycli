package containers

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
	"github.com/wedeploy/cli/hooks"
	"github.com/wedeploy/cli/verbose"
	"github.com/wedeploy/cli/verbosereq"
)

// Containers list
type Containers []Container

// Get a container from the container list
func (cs Containers) Get(id string) (c Container, err error) {
	for _, c := range cs {
		if c.ServiceID == id {
			return c, nil
		}
	}

	return c, errors.New("No container found")
}

// Container structure
type Container struct {
	ServiceID string            `json:"serviceId,omitempty"`
	Health    string            `json:"health,omitempty"`
	Type      string            `json:"type,omitempty"`
	Hooks     *hooks.Hooks      `json:"hooks,omitempty"`
	Env       map[string]string `json:"env,omitempty"`
	Scale     int               `json:"scale,omitempty"`
	HealthUID string            `json:"healthUid,omitempty"`
}

// ContainerPackage is the structure for container.json
type ContainerPackage struct {
	ID    string            `json:"id,omitempty"`
	Type  string            `json:"type,omitempty"`
	Hooks *hooks.Hooks      `json:"hooks,omitempty"`
	Env   map[string]string `json:"env,omitempty"`
}

// Container returns a Container type created taking container.json as base
func (cp ContainerPackage) Container() *Container {
	return &Container{
		ServiceID: cp.ID,
		Type:      cp.Type,
		Hooks:     cp.Hooks,
		Env:       cp.Env,
	}
}

// Register for the container structure
type Register struct {
	ID          string `json:"id"`
	Type        string `json:"type"`
	Category    string `json:"category"`
	Description string `json:"description"`
}

var (
	// ErrContainerNotFound happens when a container.json is not found
	ErrContainerNotFound = errors.New("Container not found")

	// ErrContainerAlreadyExists happens when a container ID already exists
	ErrContainerAlreadyExists = errors.New("Container already exists")

	// ErrInvalidContainerID happens when a container ID is invalid
	ErrInvalidContainerID = errors.New("Invalid container ID")

	// ErrEmptyProjectID happens when trying to access a project, but providing an empty ID
	ErrEmptyProjectID = errors.New("Can not get project: Project ID is empty")

	// ErrEmptyContainerID happens when trying to access a container, but providing an empty ID
	ErrEmptyContainerID = errors.New("Can not get container: Container ID is empty")

	// ErrEmptyProjectAndContainerID happens when trying to access a container, but providing empty IDs
	ErrEmptyProjectAndContainerID = errors.New("Can not get container: Project and Container ID is empty")
)

// ContainerInfo is for a tuple of container ID and Location.
type ContainerInfo struct {
	ServiceID string
	Location  string
}

// ContainerInfoList is a list of ContainerInfo
type ContainerInfoList []ContainerInfo

// GetLocations returns the locations of a given ContainerInfoList.
func (c ContainerInfoList) GetLocations() []string {
	var locations = []string{}

	for _, ci := range c {
		locations = append(locations, ci.Location)
	}

	return locations
}

// GetIDs returns the containers ids of a given ContainerInfoList.
func (c ContainerInfoList) GetIDs() []string {
	var ids = []string{}

	for _, ci := range c {
		ids = append(ids, ci.ServiceID)
	}

	return ids
}

// GetListFromDirectory returns a list of containers on the given diretory
func GetListFromDirectory(root string) (ContainerInfoList, error) {
	return (&listFromDirectoryGetter{}).Walk(root)
}

type listFromDirectoryGetter struct {
	list     ContainerInfoList
	idDirMap map[string]string
	root     string
}

func (l *listFromDirectoryGetter) Walk(root string) (ContainerInfoList, error) {
	var err error

	if len(root) != 0 {
		if l.root, err = filepath.Abs(root); err != nil {
			return nil, err
		}
	}

	l.list = ContainerInfoList{}
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

	if info.Name() == ".nocontainer" {
		return filepath.SkipDir
	}

	if info.IsDir() || info.Name() != "container.json" {
		return nil
	}

	return l.readFunc(path)
}

func (l *listFromDirectoryGetter) readFunc(path string) error {
	var dir = strings.TrimSuffix(path, string(os.PathSeparator)+"container.json")
	var container, errRead = Read(dir)

	switch {
	case errRead == nil:
		return l.addFunc(container, dir)
	case errRead == ErrContainerNotFound:
		return nil
	default:
		return errwrap.Wrapf("Can not list containers: error reading "+
			path+
			": {{err}}", errRead)
	}
}

func (l *listFromDirectoryGetter) addFunc(cp *ContainerPackage, dir string) error {
	if cpv, ok := l.idDirMap[cp.ID]; ok {
		return fmt.Errorf(`Can not list containers: ID "%v" was found duplicated on containers %v and %v`,
			cp.ID,
			cpv,
			dir)
	}

	l.idDirMap[cp.ID] = dir
	l.list = append(l.list, ContainerInfo{
		ServiceID: cp.ID,
		Location:  strings.TrimPrefix(dir, l.root+string(os.PathSeparator)),
	})

	return filepath.SkipDir
}

// List containers of a given project
func List(ctx context.Context, projectID string) (Containers, error) {
	var cs Containers
	var err = apihelper.AuthGet(ctx, "/projects/"+url.QueryEscape(projectID)+"/services", &cs)
	sort.Slice(cs, func(i, j int) bool {
		return cs[i].ServiceID < cs[j].ServiceID
	})
	return cs, err
}

// Get container
func Get(ctx context.Context, projectID, containerID string) (c Container, err error) {
	switch {
	case projectID == "" && containerID == "":
		return c, ErrEmptyProjectAndContainerID
	case projectID == "":
		return c, ErrEmptyProjectID
	case containerID == "":
		return c, ErrEmptyContainerID
	}

	err = apihelper.AuthGet(ctx, "/projects/"+
		url.QueryEscape(projectID)+
		"/services/"+
		url.QueryEscape(containerID), &c)
	return c, err
}

// EnvironmentVariable of a container
type EnvironmentVariable struct {
	Name  string `json:"name,omitempty"`
	Value string `json:"value,omitempty"`
}

// GetEnvironmentVariables of a service
func GetEnvironmentVariables(ctx context.Context, projectID, containerID string) (envs []EnvironmentVariable, err error) {
	err = apihelper.AuthGet(ctx, "/projects/"+
		url.QueryEscape(projectID)+
		"/services/"+
		url.QueryEscape(containerID)+
		"/environment-variables", &envs)
	return envs, err
}

func extractType(container Container) (image, version string, err error) {
	var s = strings.SplitN(container.Type, ":", 2)

	if len(s) == 1 {
		return s[0], "", nil
	}

	if strings.Contains(s[1], ":") {
		return s[0], s[1], fmt.Errorf(`Invalid container type: "%s"`, container.Type)
	}

	return s[0], s[1], nil
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

// Link container to project
func Link(ctx context.Context, projectID string, container Container, source string) (err error) {
	var reqBody = linkRequestBody{
		ServiceID: container.ServiceID,
		Scale:     container.Scale,
		Env:       container.Env,
		Source:    source,
	}

	reqBody.Image, reqBody.Version, err = extractType(container)

	if err != nil {
		return err
	}

	verbose.Debug("Linking container " + container.ServiceID + " to project " + projectID)

	var req = apihelper.URL(ctx, "/projects", url.QueryEscape(projectID), "/services")
	apihelper.Auth(req)

	err = apihelper.SetBody(req, reqBody)

	if err != nil {
		return err
	}

	return apihelper.Validate(req, req.Post())
}

// Unlink container
func Unlink(ctx context.Context, projectID, containerID string) error {
	var req = apihelper.URL(ctx,
		"/projects",
		url.QueryEscape(projectID),
		"/services",
		url.QueryEscape(containerID))
	apihelper.Auth(req)

	return apihelper.Validate(req, req.Delete())
}

// GetRegistry gets a list of container images
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

// Read a container directory properties (defined by a container.json on it)
func Read(path string) (*ContainerPackage, error) {
	var content, err = ioutil.ReadFile(filepath.Join(path, "container.json"))
	var data ContainerPackage

	if err != nil {
		return nil, readValidate(data, err)
	}

	err = json.Unmarshal(content, &data)

	return &data, readValidate(data, err)
}

func readValidate(container ContainerPackage, err error) error {
	switch {
	case os.IsNotExist(err):
		return ErrContainerNotFound
	case err != nil:
		return err
	case container.ID == "":
		return ErrInvalidContainerID
	default:
		return err
	}
}

// SetEnvironmentVariable sets an environment variable
func SetEnvironmentVariable(ctx context.Context, projectID, containerID, key, value string) error {
	var req = apihelper.URL(ctx,
		"/projects",
		url.QueryEscape(projectID),
		"/services",
		url.QueryEscape(containerID),
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
func UnsetEnvironmentVariable(ctx context.Context, projectID, containerID, key string) error {
	var req = apihelper.URL(ctx,
		"/projects",
		url.QueryEscape(projectID),
		"/services",
		url.QueryEscape(containerID),
		"/environment-variables/"+
			url.QueryEscape(key))

	apihelper.Auth(req)

	return apihelper.Validate(req, req.Delete())
}

// Restart restarts a container inside a project
func Restart(ctx context.Context, projectID, serviceID string) error {
	var req = apihelper.URL(ctx, "/projects/"+
		url.QueryEscape(projectID)+
		"/services/"+
		url.QueryEscape(serviceID)+
		"/restart")

	apihelper.Auth(req)
	return apihelper.Validate(req, req.Post())
}

// Validate container
func Validate(ctx context.Context, projectID, containerID string) (err error) {
	var req = apihelper.URL(ctx, "/validators/services/id")
	err = doValidate(projectID, containerID, req)

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

func doValidate(projectID, containerID string, req *wedeploy.WeDeploy) error {
	apihelper.Auth(req)

	req.Param("projectId", projectID)
	req.Param("value", containerID)

	var err = req.Get()

	verbosereq.Feedback(req)
	return err
}

func getValidateAPIFaultError(errDoc apihelper.APIFault) error {
	switch {
	case errDoc.Has("invalidContainerId"):
		return ErrInvalidContainerID
	case errDoc.Has("containerAlreadyExists"):
		return ErrContainerAlreadyExists
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
