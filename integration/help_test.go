package integration

import (
	"strings"
	"testing"

	"github.com/kylelemons/godebug/diff"
	"github.com/wedeploy/cli/tdata"
)

func TestHelp(t *testing.T) {
	defer Teardown()
	Setup()

	var cmd = &Command{
		Args: []string{"help"},
		Env:  []string{"WEDEPLOY_CUSTOM_HOME=" + GetLoginHome()},
	}

	cmd.Run()

	if cmd.ExitCode != 0 {
		t.Errorf("Exit code for \"lcp help\" was %v, instead of expected 0", cmd.ExitCode)
	}

	if cmd.Stderr.String() != "" {
		t.Errorf("Wanted stderr to be empty, got %v instead", cmd.Stderr.String())
	}

	if update {
		tdata.ToFile("mocks/help", cmd.Stdout.String())
	}

	var want = tdata.FromFile("mocks/help")

	var strippedGot = stripSpaces(cmd.Stdout.String())
	var strippedWant = stripSpaces(want)

	if strippedGot != strippedWant {
		t.Errorf("Stdout does not match with expected value: %v", diff.Diff(want, cmd.Stdout.String()))
	}
}

func stripSpaces(s string) string {
	return strings.TrimSpace(strings.Replace(s, " ", "", -1))
}
