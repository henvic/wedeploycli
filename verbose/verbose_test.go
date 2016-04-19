package verbose

import (
	"bytes"
	"os"
	"testing"
)

var bufErrStream bytes.Buffer

func TestMain(m *testing.M) {
	var defaultErrStream = ErrStream
	ErrStream = &bufErrStream
	ec := m.Run()
	ErrStream = defaultErrStream
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
