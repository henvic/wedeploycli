package containers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/launchpad-project/api.go/jsonlib"
	"github.com/launchpad-project/cli/apihelper"
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

	jsonlib.AssertJSONMarshal(t, tdata.FromFile(
		filepath.Join(workingDir, "mocks/app/email/container_ref.json")),
		c)

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
	servertest.Setup()
	globalconfigmock.Setup()
	bufOutStream.Reset()

	var want = tdata.FromFile("mocks/want_containers")

	servertest.Mux.HandleFunc("/projects/images/containers",
		tdata.ServerJSONFileHandler("mocks/containers_response.json"))

	List("images")

	if bufOutStream.String() != want {
		t.Errorf("Wanted %v, got %v instead", want, bufOutStream.String())
	}

	servertest.Teardown()
	globalconfigmock.Teardown()
}

func TestInstallFromDefinition(t *testing.T) {
	servertest.Setup()
	globalconfigmock.Setup()
	bufOutStream.Reset()

	servertest.Mux.HandleFunc(
		"/containers",
		func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "PUT" {
				t.Errorf("Expected install method to be POST")
			}

			var body, err = ioutil.ReadAll(r.Body)

			if err != nil {
				t.Error(err)
			}

			var data map[string]string

			err = json.Unmarshal(body, &data)

			if err != nil {
				t.Error(err)
			}

			jsonlib.AssertJSONMarshal(t,
				`{"id":"speaker", "name": "Speaker", "type": "nodejs"}`,
				data)
		})

	var c = &Container{
		ID:   "speaker",
		Name: "Speaker",
		Type: "nodejs",
	}

	var err = InstallFromDefinition("sound", c)

	if err != nil {
		t.Errorf("Unexpected error on Install: %v", err)
	}

	servertest.Teardown()
	globalconfigmock.Teardown()
}

func TestGetStatus(t *testing.T) {
	t.SkipNow()
	servertest.Setup()
	globalconfigmock.Setup()
	bufOutStream.Reset()

	var want = "on (foo bar)\n"
	servertest.Mux.HandleFunc("/projects/foo/containers/bar/state",
		tdata.ServerJSONHandler(`"on"`))

	GetStatus("foo", "bar")

	if bufOutStream.String() != want {
		t.Errorf("Wanted %v, got %v instead", want, bufOutStream.String())
	}

	servertest.Teardown()
	globalconfigmock.Teardown()
}

func TestRegistry(t *testing.T) {
	servertest.Setup()
	globalconfigmock.Setup()
	bufOutStream.Reset()

	servertest.Mux.HandleFunc(
		"/registry",
		tdata.ServerJSONFileHandler("mocks/registry.json"))

	var registry = GetRegistry()

	if len(registry) != 7 {
		t.Errorf("Expected registry to have 7 images")
	}

	servertest.Teardown()
	globalconfigmock.Teardown()
}

func TestRestart(t *testing.T) {
	servertest.Setup()
	globalconfigmock.Setup()
	bufOutStream.Reset()

	servertest.Mux.HandleFunc("/restart/container",
		func(w http.ResponseWriter, r *http.Request) {
			if r.URL.RawQuery != "projectId=foo&containerId=bar" {
				t.Error("Wrong query parameters for restart method")
			}

			fmt.Fprintf(w, `"on"`)
		})

	Restart("foo", "bar")

	servertest.Teardown()
	globalconfigmock.Teardown()
}

func TestValidate(t *testing.T) {
	servertest.Setup()
	globalconfigmock.Setup()

	servertest.Mux.HandleFunc("/validators/containers/id",
		func(w http.ResponseWriter, r *http.Request) {
			if r.FormValue("projectId") != "foo" {
				t.Errorf("Wrong projectId form value")
			}

			if r.FormValue("value") != "bar" {
				t.Errorf("Wrong containerId form value")
			}
		})

	if err := Validate("foo", "bar"); err != nil {
		t.Errorf("Wanted null error, got %v instead", err)
	}

	servertest.Teardown()
	globalconfigmock.Teardown()
}

func TestValidateAlreadyExists(t *testing.T) {
	servertest.Setup()
	globalconfigmock.Setup()

	servertest.Mux.HandleFunc("/validators/containers/id",
		func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-type", "application/json; charset=UTF-8")
			w.WriteHeader(404)
			fmt.Fprintf(w, tdata.FromFile("mocks/container_already_exists_response.json"))
		})

	if err := Validate("foo", "bar"); err != ErrContainerAlreadyExists {
		t.Errorf("Wanted %v error, got %v instead", ErrContainerAlreadyExists, err)
	}

	servertest.Teardown()
	globalconfigmock.Teardown()
}

func TestValidateInvalidID(t *testing.T) {
	servertest.Setup()
	globalconfigmock.Setup()

	servertest.Mux.HandleFunc("/validators/containers/id",
		func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-type", "application/json; charset=UTF-8")
			w.WriteHeader(404)
			fmt.Fprintf(w, tdata.FromFile("mocks/container_invalid_id_response.json"))
		})

	if err := Validate("foo", "bar"); err != ErrInvalidContainerID {
		t.Errorf("Wanted %v error, got %v instead", ErrInvalidContainerID, err)
	}

	servertest.Teardown()
	globalconfigmock.Teardown()
}

func TestValidateError(t *testing.T) {
	servertest.Setup()
	globalconfigmock.Setup()

	servertest.Mux.HandleFunc("/validators/containers/id",
		func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-type", "application/json; charset=UTF-8")
			w.WriteHeader(400)
			fmt.Fprintf(w, tdata.FromFile("../apihelper/mocks/unknown_error_api_response.json"))
		})

	var err = Validate("foo", "bar")

	switch err.(type) {
	case apihelper.APIFault:
	default:
		t.Errorf("Wanted error to be apihelper.APIFault, got %v instead", err)
	}

	servertest.Teardown()
	globalconfigmock.Teardown()
}

func TestValidateInvalidError(t *testing.T) {
	servertest.Setup()
	globalconfigmock.Setup()

	servertest.Mux.HandleFunc("/validators/containers/id",
		func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-type", "application/json; charset=UTF-8")
			w.WriteHeader(404)
		})

	var err = Validate("foo", "bar")

	if err == nil || err.Error() != "unexpected end of JSON input" {
		t.Errorf("Expected error didn't happen")
	}

	servertest.Teardown()
	globalconfigmock.Teardown()
}
