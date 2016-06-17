package integration

import (
	"testing"

	"github.com/wedeploy/cli/servertest"
	"github.com/wedeploy/cli/tdata"
)

func TestProjects(t *testing.T) {
	defer Teardown()
	Setup()

	servertest.IntegrationMux.HandleFunc(
		"/projects",
		tdata.ServerJSONFileHandler("mocks/projects_response.json"))

	var cmd = &Command{
		Args: []string{"projects"},
		Env:  []string{"WEDEPLOY_CUSTOM_HOME=" + GetLoginHome()},
	}

	var e = &Expect{
		Stdout:   tdata.FromFile("mocks/want_projects"),
		ExitCode: 0,
	}

	cmd.Run()
	e.Assert(t, cmd)
}
