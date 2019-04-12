package integration

import (
	"fmt"
	"runtime"
	"testing"

	"strings"

	"github.com/wedeploy/cli/defaults"
)

func TestVersion(t *testing.T) {
	var cmd = &Command{
		Args: []string{"version"},
	}

	var os = runtime.GOOS
	var arch = runtime.GOARCH
	var version = fmt.Sprintf(
		"Liferay CLI version %s %s/%s\n",
		defaults.Version,
		os,
		arch)

	cmd.Run()

	if cmd.ExitCode != 0 {
		t.Errorf("Wanted exit code 0, got %v instead", cmd.ExitCode)
	}

	if !strings.Contains(cmd.Stdout.String(), version) {
		t.Errorf("Wanted version message doesn't contain %v, got %v instead", version, cmd.Stdout)
	}
}
