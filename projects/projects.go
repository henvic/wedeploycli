package projects

import (
	"fmt"
	"io"
	"os"

	"github.com/launchpad-project/cli/apihelper"
)

type Project struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Id          string   `json:"id"`
	OwnerId     string   `json:"ownerId"`
	Url         string   `json:"url"`
	Users       []string `json:"users"`
}

var outStream io.Writer = os.Stdout

func List() {
	var projects []Project
	var req = apihelper.URL("/api/projects")

	apihelper.Auth(req)
	apihelper.ValidateOrExit(req, req.Get())
	apihelper.DecodeJSON(req, &projects)

	for _, project := range projects {
		fmt.Fprintln(outStream, project.Name+" "+project.Id)
	}
}
