package progress

import (
	"io/ioutil"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/wedeploy/cli/tdata"
)

func TestNew(t *testing.T) {
	// there is currently a hack that makes setting 100 => 99, see below
	var defaultOutStream = progressList.Out
	var tmp, err = ioutil.TempFile(os.TempDir(), "wedeploy-cli-test")

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

	assertProgress(t, 0, bar.Current())

	bar.Set(40)
	bar.Reset("copying", "eta 01:00")
	assertProgress(t, 0, bar.Current())

	wantPrepend := "copying"
	wantAppend := "eta 01:00"

	if bar.Prepend != wantPrepend {
		t.Errorf("Wanted prepend: %v, got %v instead", wantPrepend, bar.Prepend)
	}

	if bar.Append != wantAppend {
		t.Errorf("Wanted append: %v, got %v instead", wantAppend, bar.Append)
	}

	bar.Set(97)
	assertProgress(t, 97, bar.Current())

	bar.Flow()
	assertProgress(t, 98, bar.Current())

	bar.Flow()
	assertProgress(t, 99, bar.Current())

	bar.Flow()
	assertProgress(t, 0, bar.Current())

	// test hack as uiprogress show = as last character when 100%
	// and > (as in ===>) is desired
	bar.Set(100)
	assertProgress(t, 99, bar.Current())

	time.Sleep(50 * time.Millisecond)

	Stop()

	var want = tdata.FromFile("mocks/progress_output")
	var got = tdata.FromFile(tmp.Name())

	if !strings.Contains(got, want) {
		t.Error("Progress output doesn't contains any of the wanted progress")
	}

	os.Remove(tmp.Name())
	progressList.Out = defaultOutStream
}

func TestFail(t *testing.T) {
	if _, travis := os.LookupEnv("TRAVIS"); travis {
		t.Skip("Not testing on Travis due to weird issue. See issue #31.")
	}

	// there is currently a hack that makes setting 100 => 99, see below
	var defaultOutStream = progressList.Out
	var tmp, err = ioutil.TempFile(os.TempDir(), "wedeploy-cli-test")

	if err != nil {
		panic(err)
	}

	progressList.Out = tmp
	Start()

	bar := New("failure")
	bar.Set(47)
	assertProgress(t, 47, bar.Current())

	bar.Fail()

	time.Sleep(50 * time.Millisecond)

	Stop()

	var want = tdata.FromFile("mocks/progress_output_failure")
	var got = tdata.FromFile(tmp.Name())

	if !strings.Contains(got, want) {
		t.Error("Progress output doesn't contains any of the wanted progress")
	}

	os.Remove(tmp.Name())
	progressList.Out = defaultOutStream
}

func assertProgress(t *testing.T, want, got int) {
	if got != want {
		t.Errorf("Wanted bar to be %v, got %v instead", want, got)
	}
}
