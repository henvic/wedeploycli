package progress

import (
	"io/ioutil"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/launchpad-project/cli/tdata"
)

func TestNew(t *testing.T) {
	var defaultOutStream = progressList.Out
	var tmp, err = ioutil.TempFile(os.TempDir(), "launchpad-cli-test")

	if err != nil {
		panic(err)
	}

	progressList.Out = tmp
	Start()

	bar := New("foo")
	wantName := "foo"

	if bar.Name != wantName {
		t.Errorf("Wanted name to be %v, got %v instead", wantName, bar.Name)
	}

	wantCurrent := 0
	current := bar.Current()

	if current != wantCurrent {
		t.Errorf("Wanted bar to be %v, got %v instead", wantCurrent, current)
	}

	bar.Incr()
	wantCurrent = 1
	current = bar.Current()

	if current != wantCurrent {
		t.Errorf("Wanted bar to be %v, got %v instead", wantCurrent, current)
	}

	bar.Reset("copying", "eta 01:00")

	wantCurrent = 0
	current = bar.Current()

	if current != wantCurrent {
		t.Errorf("Wanted bar to be %v, got %v instead", wantCurrent, current)
	}

	wantPrepend := "copying"
	wantAppend := "eta 01:00"

	if bar.Prepend != wantPrepend {
		t.Errorf("Wanted prepend: %v, got %v instead", wantPrepend, bar.Prepend)
	}

	if bar.Append != wantAppend {
		t.Errorf("Wanted append: %v, got %v instead", wantAppend, bar.Append)
	}

	bar.Set(100)
	wantCurrent = 100
	current = bar.Current()

	if current != wantCurrent {
		t.Errorf("Wanted bar to be %v, got %v instead", wantCurrent, current)
	}

	time.Sleep(1 * time.Second)

	Stop()

	var want = tdata.FromFile("mocks/progress_output")
	var got = tdata.FromFile(tmp.Name())

	if !strings.Contains(got, want) {
		t.Error("Progress output doesn't contains any of the wanted progress")
	}

	os.Remove(tmp.Name())
	progressList.Out = defaultOutStream
}
