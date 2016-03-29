package integration

import (
	"testing"

	"github.com/launchpad-project/cli/servertest"
	"github.com/launchpad-project/cli/tdata"
)

func TestStatusProject(t *testing.T) {
	defer Teardown()
	Setup()

	servertest.IntegrationMux.HandleFunc(
		"/api/projects/foo/state", tdata.ServerHandler(`"on"`))

	var cmd = &Command{
		Args: []string{"status", "foo"},
		Env:  []string{"LAUNCHPAD_CUSTOM_HOME=" + GetLoginHome()},
	}

	var e = &Expect{
		Stdout:   "on (foo)\n",
		ExitCode: 0,
	}

	cmd.Run()
	e.AssertExact(t, cmd)
}

func TestStatusContainer(t *testing.T) {
	defer Teardown()
	Setup()

	servertest.IntegrationMux.HandleFunc(
		"/api/projects/foo/containers/bar/state", tdata.ServerHandler(`"on"`))

	var cmd = &Command{
		Args: []string{"status", "foo", "bar"},
		Env:  []string{"LAUNCHPAD_CUSTOM_HOME=" + GetLoginHome()},
	}

	var e = &Expect{
		Stdout:   "on (foo bar)\n",
		ExitCode: 0,
	}

	cmd.Run()
	e.AssertExact(t, cmd)
}
