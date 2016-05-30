package containers

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"

	"github.com/launchpad-project/api.go"
	"github.com/launchpad-project/cli/apihelper"
	"github.com/launchpad-project/cli/config"
	"github.com/launchpad-project/cli/hooks"
	"github.com/launchpad-project/cli/verbose"
)

// Containers map
type Containers map[string]*Container

// Container structure
type Container struct {
	ID           string            `json:"id"`
	Name         string            `json:"name"`
	Port         int               `json:"port,omitempty"`
	State        string            `json:"state,omitempty"`
	Type         string            `json:"type,omitempty"`
	Hooks        *hooks.Hooks      `json:"hooks,omitempty"`
	DeployIgnore []string          `json:"deploy_ignore,omitempty"`
	Env          map[string]string `json:"env,omitempty"`
	Instances    int               `json:"instances,omitempty"`
}

// Register for the container structure
type Register struct {
	Category         string            `json:"category"`
	ContainerDefault RegisterContainer `json:"containerDefault"`
	Description      string            `json:"description"`
}

// RegisterContainer structure for container on register
type RegisterContainer struct {
	ID   string            `json:"id"`
	Name string            `json:"name"`
	Type string            `json:"type"`
	Env  map[string]string `json:"env"`
}

var (
	// ErrContainerAlreadyExists happens when a container ID already exists
	ErrContainerAlreadyExists = errors.New("Container already exists")

	// ErrInvalidContainerID happens when a container ID is invalid
	ErrInvalidContainerID = errors.New("Invalid container ID")

	outStream io.Writer = os.Stdout
)

// GetConfig reads the container configuration file in a given directory
func GetConfig(dir string, c *Container) error {
	content, err := ioutil.ReadFile(filepath.Join(dir, "container.json"))

	if err == nil {
		err = json.Unmarshal(content, &c)
	}

	return err
}

// GetListFromScope returns a list of containers on the current context
// actually, directories...
func GetListFromScope() ([]string, error) {
	var projectRoot = config.Context.ProjectRoot

	if config.Context.ContainerRoot != "" {
		_, container := filepath.Split(config.Context.ContainerRoot)
		return []string{container}, nil
	}

	files, err := ioutil.ReadDir(projectRoot)

	if err != nil {
		return nil, err
	}

	return getContainersFromScope(files)
}

// GetStatus gets the status for a container
func GetStatus(projectID, containerID string) string {
	var status string
	apihelper.AuthGetOrExit(
		"/projects/"+projectID+"/containers/"+containerID+"/state",
		&status)
	return status
}

// List of containers of a given project
func List(projectID string) {
	var cs Containers
	apihelper.AuthGetOrExit("/projects/"+projectID+"/containers", &cs)
	var keys = make([]string, 0, len(cs))
	for k := range cs {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	for _, k := range keys {
		container := cs[k]
		fmt.Fprintf(outStream,
			"%s\t%s.%s.liferay.io (%s) %s\n",
			container.ID,
			container.ID,
			projectID,
			container.Name,
			container.State)
	}

	fmt.Fprintln(outStream, "total", len(cs))
}

// InstallFromDefinition container to project
func InstallFromDefinition(projectID, containerPath string, container *Container) error {
	verbose.Debug("Installing container from definition")

	var req = apihelper.URL("/containers")
	apihelper.Auth(req)

	req.Param("projectId", projectID)
	maybeSetLocalContainerPath(containerPath, req)

	var err = apihelper.SetBody(req, &container)

	if err != nil {
		return err
	}

	return apihelper.Validate(req, req.Put())
}

func maybeSetLocalContainerPath(containerPath string,
	req *launchpad.Launchpad) {
	if config.Stores["global"].Get("local") == "true" {
		req.Param("source", containerPath)
	}
}

// GetRegistry gets a list of container images
func GetRegistry() (registry []Register) {
	apihelper.AuthGetOrExit("/registry", &registry)
	return registry
}

// Restart restarts a container inside a project
func Restart(projectID, containerID string) {
	var req = apihelper.URL("/restart/container?projectId=" + projectID + "&containerId=" + containerID)

	apihelper.Auth(req)
	apihelper.ValidateOrExit(req, req.Post())
}

// Validate container
func Validate(projectID, containerID string) (err error) {
	var req = apihelper.URL("/validators/containers/id")
	err = doValidate(projectID, containerID, req)

	if err == nil || err != launchpad.ErrUnexpectedResponse {
		return err
	}

	var errDoc apihelper.APIFault

	err = apihelper.DecodeJSON(req, &errDoc)

	if err != nil {
		return err
	}

	return getValidateAPIFaultError(errDoc)
}

func doValidate(projectID, containerID string, req *launchpad.Launchpad) error {
	apihelper.Auth(req)

	req.Param("projectId", projectID)
	req.Param("value", containerID)

	var err = req.Get()

	apihelper.RequestVerboseFeedback(req)
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

func getContainersFromScope(files []os.FileInfo) ([]string, error) {
	var projectRoot = config.Context.ProjectRoot
	var list []string

	for _, file := range files {
		if !file.IsDir() {
			continue
		}

		var err = GetConfig(filepath.Join(projectRoot, file.Name()), nil)

		if err == nil {
			list = append(list, file.Name())
			continue
		}

		if !os.IsNotExist(err) {
			return nil, err
		}
	}

	return list, nil
}
