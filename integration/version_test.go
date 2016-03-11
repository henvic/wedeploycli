package integration

import (
	"fmt"
	"runtime"
	"testing"

	"github.com/launchpad-project/cli/defaults"
)

func TestVersion(t *testing.T) {
	var cmd = &Command{
		Args: []string{"version"},
	}

	var os = runtime.GOOS
	var arch = runtime.GOARCH
	var version = fmt.Sprintf(
		"Launchpad CLI version %s %s/%s\n",
		defaults.Version,
		os,
		arch)

	var e = &Expect{
		Stdout:   version,
		ExitCode: 0,
	}

	cmd.Run()

	if cmd.ExitCode != e.ExitCode {
		t.Errorf("Wanted exit code %v, got %v instead", e.ExitCode, cmd.ExitCode)
	}

	errString := cmd.Stderr.String()
	outString := cmd.Stdout.String()

	if errString != e.Stderr {
		t.Errorf("Wanted Stderr %v, got %v instead", errString, e.Stderr)
	}

	if outString != e.Stdout {
		t.Errorf("Wanted Stdout %v, got %v instead", outString, e.Stdout)
	}
}
