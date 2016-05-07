package projects

import (
	"bytes"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"testing"

	"github.com/launchpad-project/cli/apihelper"
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

func TestCreateFromDefinition(t *testing.T) {
	defer servertest.Teardown()
	servertest.Setup()
	globalconfigmock.Setup()

	servertest.Mux.HandleFunc("/projects",
		func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "POST" {
				t.Errorf("Unexpected method %v", r.Method)
			}
		})

	err := CreateFromDefinition("mocks/project.json")

	if err != nil {
		t.Errorf("Wanted err to be nil, got %v instead", err)
	}

	globalconfigmock.Teardown()
}

func TestCreateFromDefinitionFailureNotFound(t *testing.T) {
	var err = CreateFromDefinition(fmt.Sprintf("foo-%d.json", rand.Int()))

	if !os.IsNotExist(err) {
		t.Errorf("Wanted err to be due to file not found, got %v instead", err)
	}
}

func TestGetStatus(t *testing.T) {
	defer servertest.Teardown()
	servertest.Setup()
	globalconfigmock.Setup()
	bufOutStream.Reset()

	var want = "on (foo)\n"

	servertest.Mux.HandleFunc(
		"/projects/foo/state", tdata.ServerJSONHandler(`"on"`))

	GetStatus("foo")

	if bufOutStream.String() != want {
		t.Errorf("Wanted %v, got %v instead", want, bufOutStream.String())
	}

	globalconfigmock.Teardown()
}

func TestList(t *testing.T) {
	defer servertest.Teardown()
	servertest.Setup()
	globalconfigmock.Setup()
	bufOutStream.Reset()

	var want = tdata.FromFile("mocks/want_projects")

	servertest.Mux.HandleFunc(
		"/projects",
		tdata.ServerJSONFileHandler("mocks/projects_response.json"))

	List()

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

	servertest.Mux.HandleFunc("/restart/project", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.RawQuery != "projectId=foo" {
			t.Error("Wrong query parameters for restart method")
		}

		fmt.Fprintf(w, `"on"`)
	})

	Restart("foo")

	globalconfigmock.Teardown()
}

func TestValidate(t *testing.T) {
	servertest.Setup()
	globalconfigmock.Setup()

	servertest.Mux.HandleFunc("/validators/project/id", func(w http.ResponseWriter, r *http.Request) {
		if r.FormValue("value") != "foo" {
			t.Errorf("Wrong value form value")
		}
	})

	if err := Validate("foo"); err != nil {
		t.Errorf("Wanted null error, got %v instead", err)
	}

	servertest.Teardown()
	globalconfigmock.Teardown()
}

func TestValidateAlreadyExists(t *testing.T) {
	servertest.Setup()
	globalconfigmock.Setup()

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
	globalconfigmock.Teardown()
}

func TestValidateInvalidID(t *testing.T) {
	servertest.Setup()
	globalconfigmock.Setup()

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
	globalconfigmock.Teardown()
}

func TestValidateError(t *testing.T) {
	servertest.Setup()
	globalconfigmock.Setup()

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
	globalconfigmock.Teardown()
}

func TestValidateInvalidError(t *testing.T) {
	servertest.Setup()
	globalconfigmock.Setup()

	servertest.Mux.HandleFunc("/validators/project/id",
		func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(404)
		})

	var err = Validate("foo")

	if err != apihelper.ErrInvalidContentType {
		t.Errorf("Expected content-type error didn't happen")
	}

	servertest.Teardown()
	globalconfigmock.Teardown()
}

func TestValidateOrCreateAlreadyExists(t *testing.T) {
	servertest.Setup()
	globalconfigmock.Setup()

	servertest.Mux.HandleFunc("/projects",
		func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-type", "application/json; charset=UTF-8")
			w.WriteHeader(400)
			fmt.Fprintf(w, tdata.FromFile("mocks/create_already_exists_response.json"))
		})

	if ok, err := ValidateOrCreate("mocks/project.json"); ok != false || err != nil {
		t.Errorf("Wanted (%v, %v), got (%v, %v) instead", false, nil, ok, err)
	}

	servertest.Teardown()
	globalconfigmock.Teardown()
}

func TestValidateOrCreateNotExists(t *testing.T) {
	servertest.Setup()
	globalconfigmock.Setup()
	bufOutStream.Reset()

	servertest.Mux.HandleFunc("/projects",
		func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "POST" {
				t.Errorf("Unexpected method %v", r.Method)
			}
		})

	var ok, err = ValidateOrCreate("mocks/project.json")

	if ok != true || err != nil {
		t.Errorf("Unexpected error on Install: (%v, %v)", ok, err)
	}

	servertest.Teardown()
	globalconfigmock.Teardown()
}

func TestValidateOrCreateInvalidError(t *testing.T) {
	servertest.Setup()
	globalconfigmock.Setup()

	var _, err = ValidateOrCreate("mocks/project.json")

	if err == nil {
		t.Errorf("Expected error, got %v instead", err)
	}

	servertest.Teardown()
	globalconfigmock.Teardown()
}
