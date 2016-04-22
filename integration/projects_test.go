package integration

import (
	"testing"

	"github.com/launchpad-project/cli/servertest"
	"github.com/launchpad-project/cli/tdata"
)

func TestProjects(t *testing.T) {
	defer Teardown()
	Setup()

	servertest.IntegrationMux.HandleFunc(
		"/api/projects",
		tdata.ServerJSONFileHandler("../projects/mocks/projects_response.json"))

	var cmd = &Command{
		Args: []string{"projects"},
		Env:  []string{"LAUNCHPAD_CUSTOM_HOME=" + GetLoginHome()},
	}

	var e = &Expect{
		Stdout:   tdata.FromFile("../projects/mocks/want_projects"),
		ExitCode: 0,
	}

	cmd.Run()
	e.Assert(t, cmd)
}
