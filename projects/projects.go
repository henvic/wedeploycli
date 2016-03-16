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
