package containers

import (
	"fmt"
	"io"
	"os"

	"github.com/launchpad-project/cli/apihelper"
)

// Containers map
type Containers map[string]*Container

// Container structure
type Container struct {
	Template string `json:"template"`
	Name     string `json:"name"`
	Image    string `json:"image"`
	ID       string `json:"id"`
}

var outStream io.Writer = os.Stdout

// GetStatus gets the status for a container
func GetStatus(projectID, containerID string) {
	var status string
	var req = apihelper.URL("/api/projects/" + projectID + "/containers/" + containerID + "/state")

	apihelper.Auth(req)
	apihelper.ValidateOrExit(req, req.Get())
	apihelper.DecodeJSON(req, &status)
	fmt.Println(status + " (" + projectID + " " + containerID + ")")
}

// List of containers of a given project
func List(id string) {
	var containers Containers
	var req = apihelper.URL("/api/projects/" + id + "/containers")

	apihelper.Auth(req)
	apihelper.ValidateOrExit(req, req.Get())
	apihelper.DecodeJSON(req, &containers)

	for _, container := range containers {
		fmt.Fprintln(outStream, container.Name+" "+container.ID+" ("+container.Image+")")
	}
}

// Restart restarts a container inside a project
func Restart(projectID string, containerID string) {
	var req = apihelper.URL("/api/restart/container?projectId=" + projectID + "&containerId=" + containerID)

	apihelper.Auth(req)
	apihelper.ValidateOrExit(req, req.Post())
}
