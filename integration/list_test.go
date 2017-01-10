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

func TestListIncompatibleUseContainerRequiresProject(t *testing.T) {
	defer Teardown()
	Setup()

	var cmd = &Command{
		Args: []string{"list", "--container", "container", "--no-color"},
		Env:  []string{"WEDEPLOY_CUSTOM_HOME=" + GetLoginHome()},
	}

	var e = &Expect{
		Stderr:   tdata.FromFile("mocks/list/incompatible-use"),
		ExitCode: 1,
	}

	cmd.Run()
	e.Assert(t, cmd)
}

func TestListContainerFromInsideProject(t *testing.T) {
	defer Teardown()
	Setup()

	servertest.IntegrationMux.HandleFunc(
		"/projects/app",
		tdata.ServerJSONFileHandler("./mocks/home/bucket/project/container/container_list.json"))

	var cmd = &Command{
		Args: []string{"list", "--container", "container", "--no-color"},
		Env:  []string{"WEDEPLOY_CUSTOM_HOME=" + GetLoginHome()},
		Dir:  "mocks/home/bucket/project",
	}

	var e = &Expect{
		Stdout:   tdata.FromFile("mocks/home/bucket/project/container/container_list_want"),
		ExitCode: 0,
	}

	cmd.Run()
	e.Assert(t, cmd)
}
