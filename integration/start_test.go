package integration

import (
	"runtime"
	"testing"

	"github.com/wedeploy/cli/tdata"
)

func TestStartFromServiceDirectory(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Not testing hooks.Start() on Windows")
	}

	Setup()

	var cmd = &Command{
		Args: []string{"start"},
		Env: []string{
			"WEDEPLOY_CUSTOM_HOME=" + GetLoginHome(),
		},
		Dir: "mocks/start/project/service",
	}

	var e = &Expect{
		ExitCode: 0,
		Stdout:   tdata.FromFile("mocks/start/project/service/expect"),
	}

	cmd.Run()
	e.Assert(t, cmd)

	Teardown()
}

func TestStartOutsideProject(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Not testing hooks.Start() on Windows")
	}

	Setup()

	var cmd = &Command{
		Args: []string{"start"},
		Env: []string{
			"WEDEPLOY_CUSTOM_HOME=" + GetLoginHome(),
		},
		Dir: "mocks/start/service",
	}

	var e = &Expect{
		ExitCode: 0,
		Stdout:   tdata.FromFile("mocks/start/service/expect"),
	}

	cmd.Run()
	e.Assert(t, cmd)

	Teardown()
}

func TestStartError(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Not testing hooks.Start() on Windows")
	}

	Setup()

	var cmd = &Command{
		Args: []string{"start"},
		Env: []string{
			"WEDEPLOY_CUSTOM_HOME=" + GetLoginHome(),
		},
		Dir: "mocks/start/project/start-error-service",
	}

	var e = &Expect{
		ExitCode: 1,
		Stderr:   tdata.FromFile("mocks/start/project/start-error-service/expect_stderr"),
		Stdout:   tdata.FromFile("mocks/start/project/start-error-service/expect_stdout"),
	}

	cmd.Run()
	e.Assert(t, cmd)

	Teardown()
}
