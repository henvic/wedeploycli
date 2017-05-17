package integration

import (
	"strings"
	"testing"
)

func TestCorruptConfig(t *testing.T) {
	var cmd = &Command{
		Args: []string{"list", "-v"},
		Env:  []string{"WEDEPLOY_CUSTOM_HOME=" + GetBrokenHome()},
	}

	cmd.Run()

	if cmd.ExitCode != 1 {
		t.Errorf("Expected exit code to be 1")
	}

	if !strings.Contains(
		cmd.Stderr.String(),
		"Error reading configuration file: key-value delimiter not found: }") {
		t.Errorf("Expected error configuration message not found")
	}
}

func TestLoggedOut(t *testing.T) {
	var cmd = &Command{
		Args: []string{"list", "-v", "--remote", "foo"},
		Env:  []string{"WEDEPLOY_CUSTOM_HOME=" + GetLogoutHome()},
	}

	var e = &Expect{
		Stderr:   "Error: Remote foo not found\n",
		ExitCode: 1,
	}

	cmd.Run()
	e.Assert(t, cmd)
}
