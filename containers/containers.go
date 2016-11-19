package containers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
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
	var files, err = ioutil.ReadDir(root)

	if err != nil {
		return nil, err
	}

	return getListFromDirectory(root, files)
}

// List containers of a given project
func List(ctx context.Context, projectID string) (Containers, error) {
	var cs Containers
	var err = apihelper.AuthGet(ctx, "/projects/"+projectID+"/containers", &cs)
	return cs, err
}

// Get container
func Get(ctx context.Context, projectID, containerID string) (Container, error) {
	var c Container
	var err = apihelper.AuthGet(ctx, "/projects/"+projectID+"/containers/"+containerID, &c)
	return c, err
}

// Link container to project
func Link(ctx context.Context, projectID, containerID, containerPath string) error {
	verbose.Debug("Installing container from definition")

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
	var req = apihelper.URL(ctx, "/restart/container?projectId="+projectID+"&containerId="+containerID)

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

func getListFromDirectory(dir string, files []os.FileInfo) (ContainerInfoList, error) {
	var list = ContainerInfoList{}
	var idToPathMap = map[string]string{}

	for _, file := range files {
		if !file.IsDir() {
			continue
		}

		var container, err = Read(filepath.Join(dir, file.Name()))

		if err == nil {
			if cp, ok := idToPathMap[container.ID]; ok {
				return nil, fmt.Errorf(`Can't list containers: ID "%v" was found duplicated on containers %v and %v`,
					container.ID,
					"./"+cp,
					"./"+file.Name())
			}

			idToPathMap[container.ID] = file.Name()
			list = append(list, ContainerInfo{
				ID:       container.ID,
				Location: file.Name(),
			})
			continue
		}

		if err != ErrContainerNotFound {
			return nil, errwrap.Wrapf("Can't list containers: error reading "+
				filepath.Join(dir, file.Name(), "container.json")+
				": {{err}}", err)
		}
	}

	return list, nil
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
