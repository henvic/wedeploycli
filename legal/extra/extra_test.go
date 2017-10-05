package extra

import (
	"flag"
	"testing"

	"github.com/wedeploy/cli/tdata"
)

var update bool

func init() {
	flag.BoolVar(&update, "update", false, "update golden files")
}

func TestLicenseNotFound(t *testing.T) {
	l := License{
		LicensePath: "mocks/not-found",
	}

	var _, err = l.Get()

	if err == nil {
		t.Error("Expected error when trying to read mock license file")
	}
}

func TestLicense(t *testing.T) {
	l := License{
		Name:        "mock",
		Package:     "example.com/mock",
		Notes:       "",
		LicensePath: "mocks/in",
	}

	var got, err = l.Get()

	if err != nil {
		t.Errorf("Error reading mock license file: %v", err)
	}

	if update {
		tdata.ToFile("mocks/out", string(got))
	}

	var want = tdata.FromFile("mocks/out")

	if string(got) != want {
		t.Errorf("Wanted output to be %v, got %v instead", want, got)
	}
}

func TestLicenseWithNote(t *testing.T) {
	l := License{
		Name:        "mock",
		Package:     "example.com/mock",
		Notes:       "modified",
		LicensePath: "mocks/in",
	}

	var got, err = l.Get()

	if err != nil {
		t.Errorf("Error reading mock license file: %v", err)
	}

	if update {
		tdata.ToFile("mocks/out-with-note", string(got))
	}

	var want = tdata.FromFile("mocks/out-with-note")

	if string(got) != want {
		t.Errorf("Wanted output to be %v, got %v instead", want, got)
	}
}
