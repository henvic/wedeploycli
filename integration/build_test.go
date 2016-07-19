package integration

import (
	"runtime"
	"testing"

	"github.com/wedeploy/cli/tdata"
)

func TestBuild(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Not testing hooks.Build() on Windows")
	}

	Setup()

	var cmd = &Command{
		Args: []string{"build"},
		Env: []string{
			"WEDEPLOY_CUSTOM_HOME=" + GetLoginHome(),
		},
		Dir: "mocks/build-project/container",
	}

	var e = &Expect{
		ExitCode: 0,
		Stdout:   tdata.FromFile("mocks/build-project/container/expect"),
	}

	cmd.Run()
	e.Assert(t, cmd)

	Teardown()
}

func TestBuildError(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Not testing hooks.Build() on Windows")
	}

	Setup()

	var cmd = &Command{
		Args: []string{"build"},
		Env: []string{
			"WEDEPLOY_CUSTOM_HOME=" + GetLoginHome(),
		},
		Dir: "mocks/build-project/build-error-container",
	}

	var e = &Expect{
		ExitCode: 0,
		Stderr:   tdata.FromFile("mocks/build-project/build-error-container/expect_stderr"),
		Stdout:   tdata.FromFile("mocks/build-project/build-error-container/expect_stdout"),
	}

	cmd.Run()
	e.Assert(t, cmd)

	Teardown()
}

func TestBuildAllVerbose(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Not testing hooks.Build() on Windows")
	}

	Setup()

	var cmd = &Command{
		Args: []string{"build", "--verbose"},
		Env: []string{
			"WEDEPLOY_CUSTOM_HOME=" + GetLoginHome(),
		},
		Dir: "mocks/build-project",
	}

	var e = &Expect{
		ExitCode: 0,
		Stderr:   tdata.FromFile("mocks/build-project/expect_stderr"),
		Stdout:   tdata.FromFile("mocks/build-project/expect_stdout"),
	}

	cmd.Run()
	e.Assert(t, cmd)

	Teardown()
}
