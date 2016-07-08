package integration

import (
	"testing"

	"github.com/wedeploy/cli/servertest"
	"github.com/wedeploy/cli/tdata"
)

func TestContainers(t *testing.T) {
	defer Teardown()
	Setup()

	servertest.IntegrationMux.HandleFunc(
		"/projects/images/containers",
		tdata.ServerJSONFileHandler("../containers/mocks/containers_response.json"))

	var cmd = &Command{
		Args: []string{"containers", "images", "--local=false"},
		Env:  []string{"WEDEPLOY_CUSTOM_HOME=" + GetLoginHome()},
	}

	var e = &Expect{
		Stdout:   tdata.FromFile("../containers/mocks/want_containers"),
		ExitCode: 0,
	}

	cmd.Run()
	e.Assert(t, cmd)
}

func TestContainersFromProjectDirectory(t *testing.T) {
	defer Teardown()
	Setup()

	servertest.IntegrationMux.HandleFunc(
		"/projects/images/containers",
		tdata.ServerJSONFileHandler("../containers/mocks/containers_response.json"))

	var cmd = &Command{
		Args: []string{"containers", "--local=false"},
		Env:  []string{"WEDEPLOY_CUSTOM_HOME=" + GetLoginHome()},
		Dir:  "mocks/home/bucket/images",
	}

	var e = &Expect{
		Stdout:   tdata.FromFile("../containers/mocks/want_containers"),
		ExitCode: 0,
	}

	cmd.Run()
	e.Assert(t, cmd)
}

func TestContainersWithouProjectContext(t *testing.T) {
	defer Teardown()
	Setup()

	var cmd = &Command{
		Args: []string{"containers", "--local=false"},
		Env:  []string{"WEDEPLOY_CUSTOM_HOME=" + GetLoginHome()},
		Dir:  "mocks/home/bucket/",
	}

	var e = &Expect{
		Stderr:   tdata.FromFile("mocks/containers/no_context"),
		ExitCode: 1,
	}

	cmd.Run()
	e.Assert(t, cmd)
}
