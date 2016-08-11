package hooks

import (
	"bytes"
	"crypto/md5"
	"errors"
	"fmt"
	"io"
	"os"
	"reflect"
	"runtime"
	"strings"
	"testing"

	"github.com/wedeploy/cli/tdata"
)

type HooksProvider struct {
	Type        string
	Hook        *Hooks
	WantOutput  string
	WantErr     string
	WantErrType bool
	WantError   error
}

var HooksCases = []HooksProvider{
	HooksProvider{
		Type: "build",
		Hook: &Hooks{
			BeforeBuild: "echo before build",
			Build:       "echo during build",
			AfterBuild:  "echo after build",
		},
		WantOutput: "before build\nduring build\nafter build\n",
		WantError:  nil,
	},
	HooksProvider{
		Type: Build,
		Hook: &Hooks{
			BeforeBuild: "echo a",
			AfterBuild:  "echo b",
		},
		WantOutput: "a\nb\n",
		WantErr:    "Error: no build hook main action\n",
		WantError:  nil,
	},
	HooksProvider{
		Type:      "not implemented",
		WantError: ErrMissingHook,
	},
	HooksProvider{
		Type: Build,
		Hook: &Hooks{
			Build: "test",
		},
		WantError: HookError{
			Command: "test",
			Err:     errors.New("exit status 1"),
		},
		WantErrType: true,
	},
}

var (
	bufErrStream bytes.Buffer
	bufOutStream bytes.Buffer
)

func TestMain(m *testing.M) {
	var defaultErrStream = errStream
	var defaultOutStream = outStream
	errStream = &bufErrStream
	outStream = &bufOutStream
	ec := m.Run()
	errStream = defaultErrStream
	outStream = defaultOutStream
	os.Exit(ec)
}

func TestHookError(t *testing.T) {
	var he = HookError{
		Command: "foo",
		Err:     errors.New("bar"),
	}

	var want = "Command foo failure: bar"
	var got = he.Error()

	if want != got {
		t.Errorf("Wanted hook error %v, got %v instead", want, got)
	}
}

func TestRunHooks(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Not testing hooks.Build() on Windows")
	}

	for _, c := range HooksCases {
		bufErrStream.Reset()
		bufOutStream.Reset()

		err := c.Hook.Run(c.Type, "")

		// get the error message + err type or the actual error
		switch c.WantErrType {
		case true:
			if reflect.TypeOf(c.WantError) != reflect.TypeOf(err) {
				t.Errorf("Different error type, expected %v, got %v instead",
					reflect.TypeOf(c.WantError),
					reflect.TypeOf(err))
			}

			if (err != nil || c.WantError != nil) && err.Error() != c.WantError.Error() {
				t.Errorf("Expected %v, got %v instead", c.WantError, err)
			}
		default:
			if err != c.WantError {
				t.Errorf("Expected %v, got %v instead", c.WantError, err)
			}
		}

		var gotOutStream = bufOutStream.String()
		var gotErrStream = bufErrStream.String()

		if gotErrStream != c.WantErr {
			t.Errorf("Expected %v, got %v instead", c.WantErr, gotErrStream)
		}

		if gotOutStream != c.WantOutput {
			t.Errorf("Expected %v, got %v instead", c.WantOutput, gotOutStream)
		}
	}
}

func TestRun(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Not testing hooks.Run() on Windows")
	}

	bufErrStream.Reset()
	bufOutStream.Reset()

	if err := Run("openssl md5 hooks.go"); err != nil {
		t.Errorf("Unexpected error %v when running md5 hooks.go", err)
	}

	h := md5.New()

	if _, err := io.WriteString(h, tdata.FromFile("./hooks.go")); err != nil {
		panic(err)
	}

	if !strings.Contains(bufOutStream.String(), fmt.Sprintf("%x", h.Sum(nil))) {
		t.Errorf("Expected Run() test to contain md5 output similar to crypto.md5")
	}

	if bufErrStream.Len() != 0 {
		t.Errorf("Unexpected err output")
	}
}
