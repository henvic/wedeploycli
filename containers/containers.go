package containers

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"

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
	State        string       `json:"state"`
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
func List(id string) {
	var containers Containers
	var req = apihelper.URL("/api/projects/" + id + "/containers")

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

type InstallParams struct {
	ID        string `json:"id"`
	Bootstrap string `json:"bootstrap"`
	Name      string `json:"name"`
	Template  string `json:"template"`
}

func Install(projectID string, c Container) (err error) {
	var req = apihelper.URL(path.Join("/api/projects", projectID, "containers", c.ID))
	verbose.Debug("Installing container")

	var params = &InstallParams{
		ID:        c.ID,
		Bootstrap: c.Bootstrap,
		Name:      c.Name,
		Template:  c.Template,
	}

	var b []byte

	b, err = json.Marshal(params)

	if err != nil {
		return err
	}

	apihelper.Auth(req)

	req.Body(bytes.NewReader(b))

	err = apihelper.Validate(req, req.Put())

	return err
}

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

type ValidateParams struct {
	ProjectID string `json:"projectId"`
	Value     string `json:"value"`
}

func Validate(projectID, containerID string) (err error) {
	var req = apihelper.URL("/api/validators/containers/id")

	var params = &ValidateParams{
		ProjectID: projectID,
		Value:     containerID,
	}

	apihelper.Auth(req)
	apihelper.ParamsFromJSON(req, params)

	err = req.Get()

	if err == nil {
		return nil
	}

	// @Everything here is to be refactored, this is a hack
	if err == launchpad.ErrUnexpectedResponse {
		body, err := ioutil.ReadAll(req.Response.Body)

		if err != nil {
			return err
		}

		b := string(body)

		if strings.Contains(b, "invalidContainerId") {
			return ErrInvalidContainerID
		}

		if strings.Contains(b, "containerAlreadyExists") {
			return ErrContainerAlreadyExists
		}
	}

	return err
}
