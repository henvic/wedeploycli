package integration

import (
	"runtime"
	"testing"

	"github.com/wedeploy/cli/tdata"
)

func TestStartFromContainerDirectory(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Not testing hooks.Start() on Windows")
	}

	Setup()

	var cmd = &Command{
		Args: []string{"start"},
		Env: []string{
			"WEDEPLOY_CUSTOM_HOME=" + GetLoginHome(),
		},
		Dir: "mocks/start/project/container",
	}

	var e = &Expect{
		ExitCode: 0,
		Stdout:   tdata.FromFile("mocks/start/project/container/expect"),
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
		Dir: "mocks/start/container",
	}

	var e = &Expect{
		ExitCode: 0,
		Stdout:   tdata.FromFile("mocks/start/container/expect"),
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
		Dir: "mocks/start/project/start-error-container",
	}

	var e = &Expect{
		ExitCode: 1,
		Stderr:   tdata.FromFile("mocks/start/project/start-error-container/expect_stderr"),
		Stdout:   tdata.FromFile("mocks/start/project/start-error-container/expect_stdout"),
	}

	cmd.Run()
	e.Assert(t, cmd)

	Teardown()
}
