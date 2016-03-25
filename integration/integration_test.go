package integration

import (
	"bytes"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type HomesProvider struct {
	in   string
	path string
}

type AssertExactProvider struct {
	Command *Command
	Expect  *Expect
	Pass    bool
}

var HomesCases = []HomesProvider{
	{
		"home",
		GetRegularHome(),
	},
	{
		"login",
		GetLoginHome(),
	},
	{
		"logout",
		GetLogoutHome(),
	},
}

var AssertExactCases = []AssertExactProvider{
	{
		&Command{
			Stderr:   bytes.NewBufferString("foo"),
			Stdout:   bytes.NewBufferString("bar"),
			ExitCode: 2,
		},
		&Expect{
			Stderr:   "foo",
			Stdout:   "bar",
			ExitCode: 2,
		},
		true,
	},
	{
		&Command{
			Stderr:   bytes.NewBufferString("foo"),
			Stdout:   bytes.NewBufferString("bar"),
			ExitCode: 3,
		},
		&Expect{
			Stderr:   "nonfoo",
			Stdout:   "bar",
			ExitCode: 3,
		},
		false,
	},
	{
		&Command{
			Stderr:   bytes.NewBufferString("foo"),
			Stdout:   bytes.NewBufferString("bar"),
			ExitCode: 4,
		},
		&Expect{
			Stderr:   "foo",
			Stdout:   "nonbar",
			ExitCode: 4,
		},
		false,
	},
	{
		&Command{
			Stderr:   bytes.NewBufferString("foo"),
			Stdout:   bytes.NewBufferString("bar"),
			ExitCode: 2,
		},
		&Expect{
			Stderr:   "foo",
			Stdout:   "bar",
			ExitCode: 3,
		},
		false,
	},
}

func TestAssertExact(t *testing.T) {
	for _, c := range AssertExactCases {
		var mockTest = &testing.T{}
		c.Expect.AssertExact(mockTest, c.Command)

		if mockTest.Failed() == c.Pass {
			t.Errorf("Mock test did not meet passing status = %v assertion", c.Pass)
		}
	}
}

func TestGetHomes(t *testing.T) {
	var base, err = os.Getwd()

	if err != nil {
		panic(err)
	}

	for _, c := range HomesCases {
		var want = filepath.Join(base, "mocks", c.in)

		if want != c.path {
			t.Errorf("Wanted home path %v, got %v instead", want, c.path)
		}
	}
}

func TestInvalidDir(t *testing.T) {
	var defaultErrStream = errStream
	var bufErrStream bytes.Buffer

	defer func() {
		r := recover()

		if r != ErrExitCodeNotAvailable {
			t.Errorf("Expected panic with %v error, got %v instead", ErrExitCodeNotAvailable, r)
		}

		if !strings.Contains(bufErrStream.String(), "no such file or directory") {
			t.Error("Expected missing 'no such file or directory' message")
		}

		errStream = defaultErrStream
	}()

	errStream = &bufErrStream

	var cmd = &Command{
		Dir: "foo",
	}
	cmd.Run()
}

func TestInvalidCommand(t *testing.T) {
	var originalBinary = binary
	var defaultErrStream = errStream
	var bufErrStream bytes.Buffer

	defer func() {
		r := recover()

		if r != ErrExitCodeNotAvailable {
			t.Errorf("Expected panic with %v error, got %v instead", ErrExitCodeNotAvailable, r)
		}

		if !strings.Contains(bufErrStream.String(), "executable file not found") {
			t.Error("Expected missing 'executable file not found error' message")
		}

		binary = originalBinary
		errStream = defaultErrStream
	}()

	binary = fmt.Sprintf("invalid-command-%d", rand.Int())
	errStream = &bufErrStream

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
		t.Errorf("Wanted Stderr %v, got %v instead", e.Stderr, errString)
	}

	if outString != e.Stdout {
		t.Errorf("Wanted Stdout %v, got %v instead", e.Stdout, outString)
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
		t.Errorf("Wanted Stderr %v, got %v instead", e.Stderr, errString)
	}

	if outString != e.Stdout {
		t.Errorf("Wanted Stdout %v, got %v instead", e.Stdout, outString)
	}

	binary = originalBinary
}
