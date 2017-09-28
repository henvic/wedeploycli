package integration

import (
	"testing"

	"github.com/kylelemons/godebug/diff"
	"github.com/wedeploy/cli/tdata"
)

func TestDiagnosticsHelpIssue321(t *testing.T) {
	// this would probably be better inside a fixedbugs package
	// https://github.com/wedeploy/cli/issues/321
	defer Teardown()
	Setup()

	var cmd = &Command{
		Args: []string{"diagnostics", "--help"},
		Env:  []string{"WEDEPLOY_CUSTOM_HOME=" + GetLoginHome()},
	}

	cmd.Run()

	if cmd.ExitCode != 0 {
		t.Errorf("Exit code for \"we diagnostics --help\" was %v, instead of expected 0", cmd.ExitCode)
	}

	if cmd.Stderr.String() != "" {
		t.Errorf("Wanted stderr to be empty, got %v instead", cmd.Stderr.String())
	}

	if update {
		tdata.ToFile("mocks/diagnostics_help", cmd.Stdout.String())
	}

	var want = tdata.FromFile("mocks/diagnostics_help")

	var strippedGot = stripSpaces(cmd.Stdout.String())
	var strippedWant = stripSpaces(want)

	if strippedGot != strippedWant {
		t.Errorf("Stdout does not match with expected value: %v", diff.Diff(want, cmd.Stdout.String()))
	}
}
