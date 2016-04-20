package deploy

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/launchpad-project/cli/config"
)

func TestErrors(t *testing.T) {
	var e error = &Errors{
		List: map[string]error{
			"foo": os.ErrExist,
			"bar": os.ErrNotExist,
		},
	}

	var want = "Deploy error: map[foo:file already exists bar:file does not exist]"

	if fmt.Sprintf("%v", e) != want {
		t.Errorf("Wanted error %v, got %v instead.", e, want)
	}
}

func TestNew(t *testing.T) {
	var workingDir, _ = os.Getwd()

	if err := os.Chdir(filepath.Join(workingDir, "mocks/myproject")); err != nil {
		t.Error(err)
	}

	config.Setup()

	var _, err = New("mycontainer")

	if err != nil {
		t.Errorf("Expected New error to be null, got %v instead", err)
	}

	config.Teardown()
	os.Chdir(workingDir)
}

func TestNewErrorContainerNotFound(t *testing.T) {
	var workingDir, _ = os.Getwd()

	if err := os.Chdir(filepath.Join(workingDir, "mocks/myproject")); err != nil {
		t.Error(err)
	}

	config.Setup()

	var _, err = New("foo")

	if !os.IsNotExist(err) {
		t.Errorf("Expected container to be not found, got %v instead", err)
	}

	config.Teardown()
	os.Chdir(workingDir)
}
