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
	"github.com/launchpad-project/cli/configstore"
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
	var list []string

	if config.Context.ContainerRoot != "" {
		_, container := filepath.Split(config.Context.ContainerRoot)
		list = append(list, container)
		return list, nil
	}

	files, err := ioutil.ReadDir(projectRoot)

	if err != nil {
		return nil, err
	}

	for _, file := range files {
		if !file.IsDir() {
			continue
		}

		var cs = configstore.Store{
			Name: file.Name(),
			Path: filepath.Join(projectRoot, file.Name(), "container.json"),
		}

		err = cs.Load()

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

// GetStatus gets the status for a container
func GetStatus(projectID, containerID string) {
	var status string
	var req = apihelper.URL("/projects/" + projectID + "/containers/" + containerID + "/state")

	apihelper.Auth(req)
	apihelper.ValidateOrExit(req, req.Get())
	apihelper.DecodeJSONOrExit(req, &status)
	fmt.Fprintln(outStream, status+" ("+projectID+" "+containerID+")")
}

// List of containers of a given project
func List(projectID string) {
	var containers Containers
	var req = apihelper.URL("/projects/" + projectID + "/containers")

	apihelper.Auth(req)
	apihelper.ValidateOrExit(req, req.Get())
	apihelper.DecodeJSONOrExit(req, &containers)

	keys := make([]string, 0, len(containers))
	for k := range containers {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	for _, k := range keys {
		container := containers[k]
		fmt.Fprintln(outStream, container.ID+"\t"+container.ID+"."+projectID+".liferay.io ("+container.Name+") "+container.State)
	}

	fmt.Fprintln(outStream, "total", len(containers))
}

// InstallFromDefinition container to project
func InstallFromDefinition(projectID, containerPath string, container *Container) error {
	verbose.Debug("Installing container from definition")

	var req = apihelper.URL("/containers")
	apihelper.Auth(req)

	req.Param("projectId", projectID)

	if config.Stores["global"].Get("local") == "true" {
		req.Param("source", containerPath)
	}

	var r, err = apihelper.EncodeJSON(&container)

	if err != nil {
		return err
	}

	req.Body(r)

	return apihelper.Validate(req, req.Put())
}

// GetRegistry gets a list of container images
func GetRegistry() (registry []Register) {
	var req = apihelper.URL("/registry")

	apihelper.Auth(req)
	apihelper.ValidateOrExit(req, req.Get())
	apihelper.DecodeJSONOrExit(req, &registry)

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

	apihelper.Auth(req)

	req.Param("projectId", projectID)
	req.Param("value", containerID)

	err = req.Get()

	apihelper.RequestVerboseFeedback(req)

	if err == nil || err != launchpad.ErrUnexpectedResponse {
		return err
	}

	var errDoc apihelper.APIFault

	err = apihelper.DecodeJSON(req, &errDoc)

	if err != nil {
		return err
	}

	for _, ed := range errDoc.Errors {
		switch ed.Reason {
		case "invalidContainerId":
			return ErrInvalidContainerID
		case "containerAlreadyExists":
			return ErrContainerAlreadyExists
		}
	}

	return errDoc
}
