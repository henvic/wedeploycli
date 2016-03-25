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
	e.AssertExact(t, cmd)
}
