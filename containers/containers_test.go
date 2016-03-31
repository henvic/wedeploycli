package containers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/launchpad-project/api.go/jsonlib"
	"github.com/launchpad-project/cli/config"
	"github.com/launchpad-project/cli/globalconfigmock"
	"github.com/launchpad-project/cli/servertest"
	"github.com/launchpad-project/cli/tdata"
)

var bufOutStream bytes.Buffer

func TestMain(m *testing.M) {
	var defaultOutStream = outStream
	outStream = &bufOutStream

	ec := m.Run()

	outStream = defaultOutStream
	os.Exit(ec)
}

func TestGetConfig(t *testing.T) {
	var workingDir, _ = os.Getwd()

	if err := os.Chdir(filepath.Join(workingDir, "mocks/app")); err != nil {
		t.Error(err)
	}

	var c Container

	if err := GetConfig("email", &c); err != nil {
		t.Errorf("Wanted err to be nil, got %v instead", err)
	}

	jsonlib.AssertJSONMarshal(t, `{
		"deploy_ignore": ["xoo", "foo"],
		"id": "email",
		"image": "",
		"name": "",
		"template": ""}`, c)

	os.Chdir(workingDir)
}

func TestGetConfigFileNotFound(t *testing.T) {
	var workingDir, _ = os.Getwd()

	if err := os.Chdir(filepath.Join(workingDir, "mocks/app")); err != nil {
		t.Error(err)
	}

	var c Container

	if err := GetConfig("unknown", &c); !os.IsNotExist(err) {
		t.Errorf("Wanted file to not exist, got %v error instead", err)
	}

	os.Chdir(workingDir)
}

func TestGetConfigCorrupted(t *testing.T) {
	var workingDir, _ = os.Getwd()

	if err := os.Chdir(filepath.Join(workingDir, "mocks/app-with-invalid-container")); err != nil {
		t.Error(err)
	}

	var c Container

	var err = GetConfig("corrupted", &c)

	if _, ok := err.(*json.SyntaxError); !ok {
		t.Errorf("Wanted err to be *json.SyntaxError, got %v instead", err)
	}

	os.Chdir(workingDir)
}

func TestGetListFromScopeFromProject(t *testing.T) {
	var workingDir, _ = os.Getwd()

	if err := os.Chdir(filepath.Join(workingDir, "mocks/app")); err != nil {
		t.Error(err)
	}

	config.Setup()

	var containers, err = GetListFromScope()

	if err != nil {
		t.Errorf("Expected %v, got %v instead", nil, err)
	}

	var wantContainers = []string{"email", "landing"}

	if !reflect.DeepEqual(containers, wantContainers) {
		t.Errorf("Want %v, got %v instead", wantContainers, containers)
	}

	os.Chdir(workingDir)
	config.Teardown()
}

func TestGetListFromScopeFromContainer(t *testing.T) {
	var workingDir, _ = os.Getwd()

	if err := os.Chdir(filepath.Join(workingDir, "mocks/app/email")); err != nil {
		t.Error(err)
	}

	config.Setup()

	var containers, err = GetListFromScope()

	if err != nil {
		t.Errorf("Expected %v, got %v instead", nil, err)
	}

	var wantContainers = []string{"email"}

	if !reflect.DeepEqual(containers, wantContainers) {
		t.Errorf("Want %v, got %v instead", wantContainers, containers)
	}

	os.Chdir(workingDir)
	config.Teardown()
}

func TestGetListFromScopeInvalid(t *testing.T) {
	var workingDir, _ = os.Getwd()

	if err := os.Chdir(filepath.Join(workingDir, "mocks/app-with-invalid-container")); err != nil {
		t.Error(err)
	}

	config.Setup()

	var containers, err = GetListFromScope()

	if containers != nil {
		t.Errorf("Expected containers to be nil, got %v instead", containers)
	}

	if err == nil || os.IsNotExist(err) {
		t.Errorf("Expected error %v to be due to invalid config", err)
	}

	os.Chdir(workingDir)
	config.Teardown()
}

func TestGetListFromScopeDirectoryNotExists(t *testing.T) {
	var workingDir, _ = os.Getwd()

	config.Setup()

	config.Context.ProjectRoot = filepath.Join(".", "not-found", fmt.Sprintf("%d", rand.Int()))

	var containers, err = GetListFromScope()

	if containers != nil {
		t.Errorf("Expected containers to be nil, got %v instead", containers)
	}

	if !os.IsNotExist(err) {
		t.Errorf("Expected error %v to be due to file not found", err)
	}

	os.Chdir(workingDir)
	config.Teardown()
}

func TestList(t *testing.T) {
	defer servertest.Teardown()
	servertest.Setup()
	globalconfigmock.Setup()
	bufOutStream.Reset()

	var want = tdata.FromFile("mocks/want_containers")

	servertest.Mux.HandleFunc("/api/projects/123/containers",
		tdata.ServerFileHandler("mocks/containers_response.json"))

	List("123")

	if bufOutStream.String() != want {
		t.Errorf("Wanted %v, got %v instead", want, bufOutStream.String())
	}

	globalconfigmock.Teardown()
}

func TestGetStatus(t *testing.T) {
	defer servertest.Teardown()
	servertest.Setup()
	globalconfigmock.Setup()
	bufOutStream.Reset()

	var want = "on (foo bar)\n"
	servertest.Mux.HandleFunc("/api/projects/foo/containers/bar/state", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, `"on"`)
	})

	GetStatus("foo", "bar")

	if bufOutStream.String() != want {
		t.Errorf("Wanted %v, got %v instead", want, bufOutStream.String())
	}

	globalconfigmock.Teardown()
}

func TestRestart(t *testing.T) {
	defer servertest.Teardown()
	servertest.Setup()
	globalconfigmock.Setup()
	bufOutStream.Reset()

	servertest.Mux.HandleFunc("/api/restart/container", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.RawQuery != "projectId=foo&containerId=bar" {
			t.Error("Wrong query parameters for restart method")
		}

		fmt.Fprintf(w, `"on"`)
	})

	Restart("foo", "bar")

	globalconfigmock.Teardown()
}
