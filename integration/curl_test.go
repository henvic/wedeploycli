package integration

import (
	"testing"

	"github.com/wedeploy/cli/tdata"
)

func TestCURLEnableFirst(t *testing.T) {
	defer Teardown()
	Setup()

	var cmd = &Command{
		Args: []string{"curl", "--remote", "lcp"},
		Env:  []string{"WEDEPLOY_CUSTOM_HOME=" + GetLoginHome()},
		Dir:  "mocks/home/",
	}

	cmd.Run()

	if update {
		tdata.ToFile("mocks/curl/enable-first", cmd.Stderr.String())
	}

	var e = &Expect{
		Stderr:   tdata.FromFile("mocks/curl/enable-first"),
		ExitCode: 1,
	}

	e.Assert(t, cmd)
}

func TestCURLMissingPath(t *testing.T) {
	defer Teardown()
	defer disableCURL(t)
	Setup()
	enableCURL(t)

	var cmd = &Command{
		Args: []string{"curl", "--remote", "lcp"},
		Env:  []string{"WEDEPLOY_CUSTOM_HOME=" + GetLoginHome()},
		Dir:  "mocks/home/",
	}

	cmd.Run()

	if update {
		tdata.ToFile("mocks/curl/missing-path", cmd.Stderr.String())
	}

	var e = &Expect{
		Stderr:   tdata.FromFile("mocks/curl/missing-path"),
		ExitCode: 2,
	}

	e.Assert(t, cmd)
}

func enableCURL(t *testing.T) {
	var cmd = &Command{
		Args: []string{"curl", "enable"},
		Env:  []string{"WEDEPLOY_CUSTOM_HOME=" + GetLoginHome()},
		Dir:  "mocks/home/",
	}

	cmd.Run()

	var e = &Expect{
		ExitCode: 0,
	}

	e.Assert(t, cmd)
}

func disableCURL(t *testing.T) {
	var cmd = &Command{
		Args: []string{"curl", "disable"},
		Env:  []string{"WEDEPLOY_CUSTOM_HOME=" + GetLoginHome()},
		Dir:  "mocks/home/",
	}

	cmd.Run()

	var e = &Expect{
		ExitCode: 0,
	}

	e.Assert(t, cmd)
}
