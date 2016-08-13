package projects

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"reflect"
	"testing"

	"github.com/wedeploy/api-go/jsonlib"
	"github.com/wedeploy/cli/apihelper"
	"github.com/wedeploy/cli/configmock"
	"github.com/wedeploy/cli/servertest"
	"github.com/wedeploy/cli/tdata"
)

var bufOutStream bytes.Buffer

func TestMain(m *testing.M) {
	var defaultOutStream = outStream
	outStream = &bufOutStream

	ec := m.Run()

	outStream = defaultOutStream
	os.Exit(ec)
}

func TestCreate(t *testing.T) {
	defer servertest.Teardown()
	servertest.Setup()
	configmock.Setup()

	servertest.Mux.HandleFunc("/projects",
		func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "POST" {
				t.Errorf("Unexpected method %v", r.Method)
			}
		})

	err := Create("mocks/little/project.json")

	if err != nil {
		t.Errorf("Wanted err to be nil, got %v instead", err)
	}

	configmock.Teardown()
}

func TestCreateFailureNotFound(t *testing.T) {
	var err = Create(fmt.Sprintf("foo-%d.json", rand.Int()))

	if !os.IsNotExist(err) {
		t.Errorf("Wanted err to be due to file not found, got %v instead", err)
	}
}

func TestGet(t *testing.T) {
	servertest.Setup()
	configmock.Setup()

	servertest.Mux.HandleFunc(
		"/projects/images",
		tdata.ServerJSONFileHandler("mocks/project_get_response.json"))

	var list, err = Get("images")

	var want = Project{
		ID:     "images",
		Name:   "Image Server",
		Health: "on",
	}

	if !reflect.DeepEqual(want, list) {
		t.Errorf("Wanted %v, got %v instead", want, list)
	}

	if err != nil {
		t.Errorf("Wanted error to be nil, got %v instead", err)
	}

	servertest.Teardown()
	configmock.Teardown()
}

func TestList(t *testing.T) {
	servertest.Setup()
	configmock.Setup()

	servertest.Mux.HandleFunc(
		"/projects",
		tdata.ServerJSONFileHandler("mocks/projects_response.json"))

	var list, err = List()

	var want = []Project{
		Project{
			ID:     "images",
			Name:   "Image Server",
			Health: "on",
		},
	}

	if !reflect.DeepEqual(want, list) {
		t.Errorf("Wanted %v, got %v instead", want, list)
	}

	if err != nil {
		t.Errorf("Wanted error to be nil, got %v instead", err)
	}

	servertest.Teardown()
	configmock.Teardown()
}

func TestRead(t *testing.T) {
	var c, err = Read("mocks/little")

	if err != nil {
		t.Errorf("Expected no error, got %v instead", err)
	}

	jsonlib.AssertJSONMarshal(t, tdata.FromFile(
		"mocks/little/project_ref.json"),
		c)
}

func TestReadFileNotFound(t *testing.T) {
	var _, err = Read("mocks/unknown")

	if err != ErrProjectNotFound {
		t.Errorf("Expected %v, got %v instead", ErrProjectNotFound, err)
	}
}

func TestReadInvalidProjectID(t *testing.T) {
	var _, err = Read("mocks/missing-id")

	if err != ErrInvalidProjectID {
		t.Errorf("Expected %v, got %v instead", ErrInvalidProjectID, err)
	}
}

func TestReadCorrupted(t *testing.T) {
	var _, err = Read("mocks/corrupted")

	if _, ok := err.(*json.SyntaxError); !ok {
		t.Errorf("Wanted err to be *json.SyntaxError, got %v instead", err)
	}
}

func TestRestart(t *testing.T) {
	defer servertest.Teardown()
	servertest.Setup()
	configmock.Setup()
	bufOutStream.Reset()

	servertest.Mux.HandleFunc("/restart/project", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.RawQuery != "projectId=foo" {
			t.Error("Wrong query parameters for restart method")
		}

		fmt.Fprintf(w, `"on"`)
	})

	if err := Restart("foo"); err != nil {
		t.Errorf("Unexpected error on project restart: %v", err)
	}

	configmock.Teardown()
}

func TestSetAuth(t *testing.T) {
	defer servertest.Teardown()
	servertest.Setup()
	configmock.Setup()

	servertest.Mux.HandleFunc("/projects/little/auth",
		func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "PUT" {
				t.Errorf("Unexpected method %v", r.Method)
			}

			got, err := ioutil.ReadAll(r.Body)

			if err != nil {
				t.Error(err)
			}

			want, err := ioutil.ReadFile("mocks/little/auth.json")

			if err != nil {
				panic(err)
			}

			if !reflect.DeepEqual(want, got) {
				t.Errorf("Received object isn't sent object.")
			}
		})

	err := SetAuth("little", "mocks/little/auth.json")

	if err != nil {
		t.Errorf("Wanted err to be nil, got %v instead", err)
	}

	configmock.Teardown()
}

func TestSetAuthFailureNotFound(t *testing.T) {
	var err = SetAuth("foo", fmt.Sprintf("foo-%d.json", rand.Int()))

	if !os.IsNotExist(err) {
		t.Errorf("Wanted err to be due to file not found, got %v instead", err)
	}
}

func TestUnlink(t *testing.T) {
	servertest.Setup()
	configmock.Setup()

	servertest.Mux.HandleFunc("/deploy/foo", func(w http.ResponseWriter, r *http.Request) {
		var wantMethod = "DELETE"
		if r.Method != wantMethod {
			t.Errorf("Wanted method %v, got %v instead", wantMethod, r.Method)
		}
	})

	var err = Unlink("foo")

	if err != nil {
		t.Errorf("Expected no error, got %v instead", err)
	}

	servertest.Teardown()
	configmock.Teardown()
}

func TestValidate(t *testing.T) {
	servertest.Setup()
	configmock.Setup()

	servertest.Mux.HandleFunc("/validators/project/id", func(w http.ResponseWriter, r *http.Request) {
		if r.FormValue("value") != "foo" {
			t.Errorf("Wrong value form value")
		}
	})

	if err := Validate("foo"); err != nil {
		t.Errorf("Wanted null error, got %v instead", err)
	}

	servertest.Teardown()
	configmock.Teardown()
}

func TestValidateAlreadyExists(t *testing.T) {
	servertest.Setup()
	configmock.Setup()

	servertest.Mux.HandleFunc("/validators/project/id",
		func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-type", "application/json; charset=UTF-8")
			w.WriteHeader(404)
			fmt.Fprintf(w, tdata.FromFile("mocks/project_already_exists_response.json"))
		})

	if err := Validate("foo"); err != ErrProjectAlreadyExists {
		t.Errorf("Wanted %v error, got %v instead", ErrProjectAlreadyExists, err)
	}

	servertest.Teardown()
	configmock.Teardown()
}

func TestValidateInvalidID(t *testing.T) {
	servertest.Setup()
	configmock.Setup()

	servertest.Mux.HandleFunc("/validators/project/id",
		func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-type", "application/json; charset=UTF-8")
			w.WriteHeader(404)
			fmt.Fprintf(w, tdata.FromFile("mocks/project_invalid_id_response.json"))
		})

	if err := Validate("foo"); err != ErrInvalidProjectID {
		t.Errorf("Wanted %v error, got %v instead", ErrInvalidProjectID, err)
	}

	servertest.Teardown()
	configmock.Teardown()
}

func TestValidateError(t *testing.T) {
	servertest.Setup()
	configmock.Setup()

	servertest.Mux.HandleFunc("/validators/project/id",
		func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-type", "application/json; charset=UTF-8")
			w.WriteHeader(400)
			fmt.Fprintf(w, tdata.FromFile("../apihelper/mocks/unknown_error_api_response.json"))
		})

	var err = Validate("foo")

	switch err.(type) {
	case apihelper.APIFault:
	default:
		t.Errorf("Wanted error to be apihelper.APIFault, got %v instead", err)
	}

	servertest.Teardown()
	configmock.Teardown()
}

func TestValidateInvalidError(t *testing.T) {
	servertest.Setup()
	configmock.Setup()

	servertest.Mux.HandleFunc("/validators/project/id",
		func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(404)
		})

	var err = Validate("foo")

	if err != apihelper.ErrInvalidContentType {
		t.Errorf("Expected content-type error didn't happen")
	}

	servertest.Teardown()
	configmock.Teardown()
}

func TestValidateOrCreateAlreadyExists(t *testing.T) {
	servertest.Setup()
	configmock.Setup()

	servertest.Mux.HandleFunc("/projects",
		func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-type", "application/json; charset=UTF-8")
			w.WriteHeader(400)
			fmt.Fprintf(w, tdata.FromFile("mocks/create_already_exists_response.json"))
		})

	if ok, err := ValidateOrCreate("mocks/little/project.json"); ok != false || err != nil {
		t.Errorf("Wanted (%v, %v), got (%v, %v) instead", false, nil, ok, err)
	}

	servertest.Teardown()
	configmock.Teardown()
}

func TestValidateOrCreateNotExists(t *testing.T) {
	servertest.Setup()
	configmock.Setup()
	bufOutStream.Reset()

	servertest.Mux.HandleFunc("/projects",
		func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "POST" {
				t.Errorf("Unexpected method %v", r.Method)
			}
		})

	var ok, err = ValidateOrCreate("mocks/little/project.json")

	if ok != true || err != nil {
		t.Errorf("Unexpected error on Install: (%v, %v)", ok, err)
	}

	servertest.Teardown()
	configmock.Teardown()
}

func TestValidateOrCreateInvalidError(t *testing.T) {
	servertest.Setup()
	configmock.Setup()

	var _, err = ValidateOrCreate("mocks/little/project.json")

	if err == nil {
		t.Errorf("Expected error, got %v instead", err)
	}

	servertest.Teardown()
	configmock.Teardown()
}
