package integration

import (
	"testing"

	"github.com/launchpad-project/cli/servertest"
	"github.com/launchpad-project/cli/tdata"
)

func TestContainers(t *testing.T) {
	defer Teardown()
	Setup()

	servertest.IntegrationMux.HandleFunc(
		"/api/projects/images/containers",
		tdata.ServerFileHandler("../containers/mocks/containers_response.json"))

	var cmd = &Command{
		Args: []string{"containers", "images"},
		Env:  []string{"LAUNCHPAD_CUSTOM_HOME=" + GetLoginHome()},
	}

	var e = &Expect{
		Stdout:   tdata.FromFile("../containers/mocks/want_containers"),
		ExitCode: 0,
	}

	cmd.Run()
	e.AssertExact(t, cmd)
}

func TestContainersFromProjectDirectory(t *testing.T) {
	defer Teardown()
	Setup()

	servertest.IntegrationMux.HandleFunc(
		"/api/projects/images/containers",
		tdata.ServerFileHandler("../containers/mocks/containers_response.json"))

	var cmd = &Command{
		Args: []string{"containers"},
		Env:  []string{"LAUNCHPAD_CUSTOM_HOME=" + GetLoginHome()},
		Dir:  "mocks/home/bucket/images",
	}

	var e = &Expect{
		Stdout:   tdata.FromFile("../containers/mocks/want_containers"),
		ExitCode: 0,
	}

	cmd.Run()
	e.AssertExact(t, cmd)
}

func TestContainersWithouProjectContext(t *testing.T) {
	defer Teardown()
	Setup()

	var cmd = &Command{
		Args: []string{"containers"},
		Env:  []string{"LAUNCHPAD_CUSTOM_HOME=" + GetLoginHome()},
		Dir:  "mocks/home/bucket/",
	}

	var e = &Expect{
		Stderr:   tdata.FromFile("mocks/containers/no_context"),
		ExitCode: 1,
	}

	cmd.Run()
	e.AssertExact(t, cmd)
}
