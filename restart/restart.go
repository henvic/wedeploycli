package restart

import (
	"fmt"
	"io"
	"os"

	"github.com/launchpad-project/cli/apihelper"
)

var (
	errStream io.Writer = os.Stderr
	outStream io.Writer = os.Stdout
)

// Project restarts a project
func Project(id string) {
	var req = apihelper.URL("/api/restart/project?projectId=" + id)

	apihelper.Auth(req)
	apihelper.ValidateOrExit(req, req.Post())

	if req.Response.StatusCode != 200 {
		fmt.Fprintln(outStream, "Can't restart:", req.Response.Status)
	}
}

// Container restarts a container inside a project
func Container(projectID string, containerID string) {
	var req = apihelper.URL("/api/restart/container?projectId=" + projectID + "&containerId=" + containerID)

	apihelper.Auth(req)
	apihelper.ValidateOrExit(req, req.Post())

	if req.Response.StatusCode != 200 {
		fmt.Fprintln(errStream, "Can't restart:", req.Response.Status)
	}
}
