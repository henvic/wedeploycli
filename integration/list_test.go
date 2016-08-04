package integration

import (
	"testing"

	"github.com/wedeploy/cli/servertest"
	"github.com/wedeploy/cli/tdata"
)

func TestList(t *testing.T) {
	defer Teardown()
	Setup()

	servertest.IntegrationMux.HandleFunc(
		"/projects",
		tdata.ServerJSONFileHandler("./mocks/list/projects_response.json"))

	var cmd = &Command{
		Args: []string{"list", "--no-color"},
		Env:  []string{"WEDEPLOY_CUSTOM_HOME=" + GetLoginHome()},
	}

	var e = &Expect{
		Stdout:   tdata.FromFile("mocks/list/want"),
		ExitCode: 0,
	}

	cmd.Run()
	e.Assert(t, cmd)
}
