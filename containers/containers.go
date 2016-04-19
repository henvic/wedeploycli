package containers

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
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
	ID           string       `json:"id"`
	Name         string       `json:"name"`
	Bootstrap    string       `json:"bootstrap"`
	State        string       `json:"state,omitempty"`
	Template     string       `json:"template"`
	Type         string       `json:"type"`
	Hooks        *hooks.Hooks `json:"hooks,omitempty"`
	DeployIgnore []string     `json:"deploy_ignore,omitempty"`
}

// Register for the container structure
type Register struct {
	Bootstrap   string `json:"bootstrap"`
	Category    string `json:"category"`
	Description string `json:"description"`
	ID          string `json:"id"`
	Name        string `json:"name"`
	Template    string `json:"template"`
	Type        string `json:"type"`
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
	var req = apihelper.URL("/api/projects/" + projectID + "/containers/" + containerID + "/state")

	apihelper.Auth(req)
	apihelper.ValidateOrExit(req, req.Get())
	apihelper.DecodeJSONOrExit(req, &status)
	fmt.Fprintln(outStream, status+" ("+projectID+" "+containerID+")")
}

// List of containers of a given project
func List(projectID string) {
	var containers Containers
	var req = apihelper.URL("/api/projects/" + projectID + "/containers")

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

// Install container to project
func Install(projectID string, c *Container) error {
	var req = apihelper.URL(path.Join("/api/projects", projectID, "containers", c.ID))
	apihelper.Auth(req)

	var reader, err = apihelper.EncodeJSON(map[string]string{
		"id":        c.ID,
		"bootstrap": c.Bootstrap,
		"name":      c.Name,
		"template":  c.Template,
	})

	if err == nil {
		req.Body(reader)
		verbose.Debug("Installing container")
		err = apihelper.Validate(req, req.Put())
	}

	return err
}

// GetRegistry gets a list of container images
func GetRegistry() (registry []Register) {
	var req = apihelper.URL("/api/registry")

	apihelper.Auth(req)
	apihelper.ValidateOrExit(req, req.Get())
	apihelper.DecodeJSONOrExit(req, &registry)

	return registry
}

// Restart restarts a container inside a project
func Restart(projectID, containerID string) {
	var req = apihelper.URL("/api/restart/container?projectId=" + projectID + "&containerId=" + containerID)

	apihelper.Auth(req)
	apihelper.ValidateOrExit(req, req.Post())
}

// Validate container
func Validate(projectID, containerID string) (err error) {
	var req = apihelper.URL("/api/validators/containers/id")

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

// ValidateOrCreate container
func ValidateOrCreate(projectID string, c *Container) (bool, error) {
	var created bool
	var err = Validate(projectID, c.ID)

	if err == ErrContainerAlreadyExists {
		return false, nil
	}

	if err != nil {
		return false, err
	}

	err = Install(projectID, c)

	if err == nil {
		created = true
	}

	return created, err
}
