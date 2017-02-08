package integration

import "testing"

func TestWho(t *testing.T) {
	defer Teardown()
	Setup()

	var cmd = &Command{
		Args: []string{"who"},
		Env:  []string{"WEDEPLOY_CUSTOM_HOME=" + GetLoginHome()},
		Dir:  "mocks/home/",
	}

	var e = &Expect{
		Stdout:   "foo",
		ExitCode: 0,
	}

	cmd.Run()
	e.Assert(t, cmd)
}

func TestWhoNotFound(t *testing.T) {
	defer Teardown()
	Setup()

	var cmd = &Command{
		Args: []string{"who"},
		Env:  []string{"WEDEPLOY_CUSTOM_HOME=" + GetLogoutHome()},
		Dir:  "mocks/home/",
	}

	var e = &Expect{
		Stderr:   "User is not available",
		ExitCode: 1,
	}

	cmd.Run()
	e.Assert(t, cmd)
}
