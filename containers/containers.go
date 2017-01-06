package containers

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
	"strings"

	"github.com/hashicorp/errwrap"
	"github.com/wedeploy/api-go"
	"github.com/wedeploy/cli/apihelper"
	"github.com/wedeploy/cli/defaults"
	"github.com/wedeploy/cli/hooks"
	"github.com/wedeploy/cli/verbose"
	"github.com/wedeploy/cli/verbosereq"
)

// Containers map
type Containers map[string]*Container

// Container structure
type Container struct {
	ID     string            `json:"id"`
	Health string            `json:"health,omitempty"`
	Type   string            `json:"type,omitempty"`
	Hooks  *hooks.Hooks      `json:"hooks,omitempty"`
	Env    map[string]string `json:"env,omitempty"`
	Scale  int               `json:"scale,omitempty"`
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
)

// ContainerInfo is for a tuple of container ID and Location.
type ContainerInfo struct {
	ID       string
	Location string
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
		ids = append(ids, ci.ID)
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

func (l *listFromDirectoryGetter) addFunc(container *Container, dir string) error {
	if cp, ok := l.idDirMap[container.ID]; ok {
		return fmt.Errorf(`Can not list containers: ID "%v" was found duplicated on containers %v and %v`,
			container.ID,
			cp,
			dir)
	}

	l.idDirMap[container.ID] = dir
	l.list = append(l.list, ContainerInfo{
		ID:       container.ID,
		Location: strings.TrimPrefix(dir, l.root+string(os.PathSeparator)),
	})

	return filepath.SkipDir
}

// List containers of a given project
func List(ctx context.Context, projectID string) (Containers, error) {
	var cs Containers
	var err = apihelper.AuthGet(ctx, "/projects/"+url.QueryEscape(projectID)+"/containers", &cs)
	return cs, err
}

// Get container
func Get(ctx context.Context, projectID, containerID string) (Container, error) {
	var c Container
	var err = apihelper.AuthGet(ctx, "/projects/"+
		url.QueryEscape(projectID)+
		"/containers/"+
		url.QueryEscape(containerID), &c)
	return c, err
}

// Link container to project
func Link(ctx context.Context, projectID, containerID, containerPath string) error {
	verbose.Debug("Linking container " + containerID + " to project " + projectID)

	var req = apihelper.URL(ctx, "/deploy")
	apihelper.Auth(req)

	req.Param("projectId", projectID)
	req.Param("containerId", containerID)
	req.Param("source", normalizePath(containerPath))

	var c, err = ioutil.ReadFile(filepath.Join(containerPath, "container.json"))

	if err != nil {
		return err
	}

	req.Body(bytes.NewReader(c))

	return apihelper.Validate(req, req.Put())
}

// Unlink container
func Unlink(ctx context.Context, projectID, containerID string) error {
	var req = apihelper.URL(ctx, "/deploy", projectID, containerID)
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
func Read(path string) (*Container, error) {
	var content, err = ioutil.ReadFile(filepath.Join(path, "container.json"))
	var data Container

	if err != nil {
		return nil, readValidate(data, err)
	}

	err = json.Unmarshal(content, &data)

	return &data, readValidate(data, err)
}

func readValidate(container Container, err error) error {
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

// Restart restarts a container inside a project
func Restart(ctx context.Context, projectID, containerID string) error {
	var req = apihelper.URL(ctx, "/restart/container?projectId="+
		url.QueryEscape(projectID)+
		"&containerId="+
		url.QueryEscape(containerID))

	apihelper.Auth(req)
	return apihelper.Validate(req, req.Post())
}

// Validate container
func Validate(ctx context.Context, projectID, containerID string) (err error) {
	var req = apihelper.URL(ctx, "/validators/containers/id")
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
