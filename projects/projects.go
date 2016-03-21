package projects

import (
	"fmt"
	"io"
	"os"

	"github.com/launchpad-project/cli/apihelper"
)

// Project structure
type Project struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	ID          string   `json:"id"`
	OwnerID     string   `json:"ownerId"`
	URL         string   `json:"url"`
	Users       []string `json:"users"`
}

var outStream io.Writer = os.Stdout

// GetStatus gets the status for the project
func GetStatus(id string) {
	var status string
	var req = apihelper.URL("/api/projects/" + id + "/state")

	apihelper.Auth(req)
	apihelper.ValidateOrExit(req, req.Get())
	apihelper.DecodeJSON(req, &status)
	fmt.Println(status + " (" + id + ")")
}

// List projects
func List() {
	var projects []Project
	var req = apihelper.URL("/api/projects")

	apihelper.Auth(req)
	apihelper.ValidateOrExit(req, req.Get())
	apihelper.DecodeJSON(req, &projects)

	for _, project := range projects {
		fmt.Fprintln(outStream, project.Name+" "+project.ID)
	}
}

// Restart restarts a project
func Restart(id string) {
	var req = apihelper.URL("/api/restart/project?projectId=" + id)

	apihelper.Auth(req)
	apihelper.ValidateOrExit(req, req.Post())
}
