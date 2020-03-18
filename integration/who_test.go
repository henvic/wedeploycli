package integration

import (
	"testing"

	"github.com/henvic/wedeploycli/tdata"
)

func TestWho(t *testing.T) {
	defer Teardown()
	Setup()

	var cmd = &Command{
		Args: []string{"who"},
		Env:  []string{"WEDEPLOY_CUSTOM_HOME=" + GetLoginHome()},
		Dir:  "mocks/home/",
	}

	cmd.Run()

	if update {
		tdata.ToFile("mocks/who/found", cmd.Stdout.String())
	}

	var e = &Expect{
		Stdout:   tdata.FromFile("mocks/who/found"),
		ExitCode: 0,
	}

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

	cmd.Run()

	if update {
		tdata.ToFile("mocks/who/not-found", cmd.Stderr.String())
	}

	var e = &Expect{
		Stderr:   tdata.FromFile("mocks/who/not-found"),
		ExitCode: 1,
	}

	e.Assert(t, cmd)
}
