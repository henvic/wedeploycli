package integration

import (
	"testing"

	"github.com/wedeploy/cli/tdata"
)

func TestHelp(t *testing.T) {
	defer Teardown()
	Setup()

	var cmd = &Command{
		Args: []string{"help"},
		Env:  []string{"WEDEPLOY_CUSTOM_HOME=" + GetLoginHome()},
	}

	var e = &Expect{
		Stderr:   "",
		Stdout:   tdata.FromFile("mocks/help"),
		ExitCode: 0,
	}

	cmd.Run()
	e.Assert(t, cmd)
}
