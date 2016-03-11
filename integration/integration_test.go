package integration

import (
	"fmt"
	"math/rand"
	"strings"
	"testing"
)

func TestInvalidCommand(t *testing.T) {
	var originalBinary = binary

	defer func() {
		r := recover()

		if r != ErrExitCodeNotAvailable {
			t.Errorf("Expected panic with %v error, got %v instead", ErrExitCodeNotAvailable, r)
		}

		binary = originalBinary
	}()

	binary = fmt.Sprintf("invalid-command-%d", rand.Int())

	var cmd = &Command{}
	cmd.Run()
}

func TestInvalidArgument(t *testing.T) {
	var invalidArg = fmt.Sprintf("invalid-arg-%d", rand.Int())
	var cmd = &Command{
		Args: []string{invalidArg},
	}

	var e = &Expect{
		Stderr: fmt.Sprintf(`Error: unknown command "%v" for "launchpad"
Run 'launchpad --help' for usage.
`, invalidArg),
		ExitCode: 255,
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

func TestStdin(t *testing.T) {
	var originalBinary = binary

	binary = "cat"

	var cmd = &Command{
		Stdin: strings.NewReader("hello"),
	}

	var e = &Expect{
		Stdout:   "hello",
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

	binary = originalBinary
}
