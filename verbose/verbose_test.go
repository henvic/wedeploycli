package verbose

import (
	"bytes"
	"os"
	"testing"

	"github.com/wedeploy/cli/color"
)

var bufErrStream bytes.Buffer

func TestMain(m *testing.M) {
	var defaultErrStream = ErrStream
	ErrStream = &bufErrStream
	color.NoColor = true
	defer func() {
		color.NoColor = false
		ErrStream = defaultErrStream
	}()
	ec := m.Run()
	os.Exit(ec)
}

func TestDebugOn(t *testing.T) {
	bufErrStream.Reset()
	Enabled = true
	Debug("Hello...", "World!")

	var want = "Hello... World!\n"
	if got := bufErrStream.String(); got != want {
		t.Errorf("Wanted %s, got %s instead", want, got)
	}
}

func TestDebugOff(t *testing.T) {
	bufErrStream.Reset()
	Enabled = false
	Debug("1, 2, 3")

	if got := bufErrStream.String(); len(got) != 0 {
		t.Errorf("Wanted no debug, got %s instead", got)
	}
}

func TestSafeEscape(t *testing.T) {
	var want = ` hidden value `

	if got := SafeEscape("moo"); got != want {
		t.Errorf("Wanted value %v, got %v instead", want, got)
	}
}

func TestSafeEscapeUnsafe(t *testing.T) {
	var defaultUnsafeVerbose = unsafeVerbose
	unsafeVerbose = true
	defer func() {
		unsafeVerbose = defaultUnsafeVerbose
	}()

	var want = `moo`

	if got := SafeEscape("moo"); got != want {
		t.Errorf("Wanted value %v, got %v instead", want, got)
	}
}

func TestSafeEscapeSlice(t *testing.T) {
	var want = ` 1 hidden value `

	if got := SafeEscapeSlice([]string{"moo"}); got != want {
		t.Errorf("Wanted value %v, got %v instead", want, got)
	}
}

func TestSafeEscapeSliceSeveral(t *testing.T) {
	var want = ` 3 hidden values `

	if got := SafeEscapeSlice([]string{"moo", "foo", "bar"}); got != want {
		t.Errorf("Wanted value %v, got %v instead", want, got)
	}
}

func TestSafeEscapeUnsafeSlice(t *testing.T) {
	var defaultUnsafeVerbose = unsafeVerbose
	if err := os.Setenv("WEDEPLOY_UNSAFE_VERBOSE", "true"); err != nil {
		t.Errorf("Error setting WEDEPLOY_UNSAFE_VERBOSE environment variable: %v", err)
	}

	unsafeVerbose = true
	defer func() {
		unsafeVerbose = defaultUnsafeVerbose
		if err := os.Unsetenv("WEDEPLOY_UNSAFE_VERBOSE"); err != nil {
			t.Errorf("Error unsetting WEDEPLOY_UNSAFE_VERBOSE environment variable: %v", err)
		}
	}()

	if !IsUnsafeMode() {
		t.Errorf("Expected unsafe mode to be on")
	}

	var want = `[moo foo bar]`

	if got := SafeEscapeSlice([]string{"moo", "foo", "bar"}); got != want {
		t.Errorf("Wanted value %v, got %v instead", want, got)
	}
}
