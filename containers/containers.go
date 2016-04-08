package containers

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"

	"github.com/launchpad-project/cli/apihelper"
	"github.com/launchpad-project/cli/config"
	"github.com/launchpad-project/cli/configstore"
	"github.com/launchpad-project/cli/hooks"
)

// Containers map
type Containers map[string]*Container

// Container structure
type Container struct {
	ID           string       `json:"id"`
	Name         string       `json:"name"`
	Bootstrap    string       `json:"bootstrap"`
	Type         string       `json:"type,omitempty"`
	Hooks        *hooks.Hooks `json:"hooks,omitempty"`
	DeployIgnore []string     `json:"deploy_ignore,omitempty"`
}

var outStream io.Writer = os.Stdout
// Register for the container structure
type Register struct {
	Bootstrap   string `json:"bootstrap"`
	Category    string `json:"category"`
	Description string `json:"description"`
	ID          string `json:"id"`
	Name        string `json:"name"`
	Template    string `json:"template"`
}


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
		fmt.Fprintln(outStream, container.ID+" ("+container.Name+")")
	}

	fmt.Fprintln(outStream, "total", len(containers))
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
