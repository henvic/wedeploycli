package projects

import (
	"fmt"
	"io"
	"os"

	"github.com/launchpad-project/cli/apihelper"
)

// Project structure
type Project struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Domain      string `json:"domain"`
	Description string `json:"description,omitempty"`
}

var outStream io.Writer = os.Stdout

// GetStatus gets the status for the project
func GetStatus(id string) {
	var status string
	var req = apihelper.URL("/api/projects/" + id + "/state")

	apihelper.Auth(req)
	apihelper.ValidateOrExit(req, req.Get())
	apihelper.DecodeJSONOrExit(req, &status)
	fmt.Fprintln(outStream, status+" ("+id+")")
}

// List projects
func List() {
	var projects []Project
	var req = apihelper.URL("/api/projects")

	apihelper.Auth(req)
	apihelper.ValidateOrExit(req, req.Get())
	apihelper.DecodeJSONOrExit(req, &projects)

	for _, project := range projects {
		fmt.Fprintln(outStream, project.ID+" ("+project.Name+")")
	}

	fmt.Fprintln(outStream, "total", len(projects))
}

// Restart restarts a project
func Restart(id string) {
	var req = apihelper.URL("/api/restart/project?projectId=" + id)

	apihelper.Auth(req)
	apihelper.ValidateOrExit(req, req.Post())
}
