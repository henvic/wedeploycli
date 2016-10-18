package projects

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/wedeploy/api-go/jsonlib"
	"github.com/wedeploy/cli/apihelper"
	"github.com/wedeploy/cli/configmock"
	"github.com/wedeploy/cli/servertest"
	"github.com/wedeploy/cli/tdata"
)

var defaultErrStream = errStream

func TestCreateFromJSON(t *testing.T) {
	defer servertest.Teardown()
	servertest.Setup()
	configmock.Setup()

	servertest.Mux.HandleFunc("/projects",
		func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "POST" {
				t.Errorf("Unexpected method %v", r.Method)
			}
		})

	err := CreateFromJSON(context.Background(), "mocks/little/project.json")

	if err != nil {
		t.Errorf("Wanted err to be nil, got %v instead", err)
	}

	configmock.Teardown()
}

func TestCreateFromJSONFailureNotFound(t *testing.T) {
	var err = CreateFromJSON(context.Background(),
		fmt.Sprintf("foo-%d.json", rand.Int()))

	if !os.IsNotExist(err) {
		t.Errorf("Wanted err to be due to file not found, got %v instead", err)
	}
}

func TestCreate(t *testing.T) {
	servertest.Setup()
	configmock.Setup()

	servertest.Mux.HandleFunc("/projects",
		tdata.ServerJSONFileHandler("mocks/new_response.json"))

	var project, err = Create(context.Background(), "")

	if project.ID != "tesla36" {
		t.Errorf("Wanted project ID to be tesla36, got %v instead", project.ID)
	}

	if project.Health != "on" {
		t.Errorf("Wanted project Health to be on, got %v instead", project.Health)
	}

	if err != nil {
		t.Errorf("Wanted error to be nil, got %v instead", err)
	}

	servertest.Teardown()
	configmock.Teardown()
}

func TestCreateNamed(t *testing.T) {
	servertest.Setup()
	configmock.Setup()

	servertest.Mux.HandleFunc("/projects",
		tdata.ServerJSONFileHandler("mocks/new_named_response.json"))

	var project, err = Create(context.Background(), "banach30")

	if project.ID != "banach30" {
		t.Errorf("Wanted project ID to be banach30, got %v instead", project.ID)
	}

	if err != nil {
		t.Errorf("Wanted error to be nil, got %v instead", err)
	}

	servertest.Teardown()
	configmock.Teardown()
}

func TestCreateError(t *testing.T) {
	servertest.Setup()
	configmock.Setup()

	var project, err = Create(context.Background(), "")

	if project != nil {
		t.Errorf("Wanted project to be nil, got %v instead", project)
	}

	switch err.(type) {
	case *apihelper.APIFault:
	default:
		t.Errorf("Wanted APIFault error, got %v instead", err)
	}

	servertest.Teardown()
	configmock.Teardown()
}

func TestGet(t *testing.T) {
	servertest.Setup()
	configmock.Setup()

	servertest.Mux.HandleFunc(
		"/projects/images",
		tdata.ServerJSONFileHandler("mocks/project_get_response.json"))

	var list, err = Get(context.Background(), "images")

	var want = Project{
		ID:     "images",
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

	var list, err = List(context.Background())

	var want = []Project{
		Project{
			ID:     "images",
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

func TestReadLegacyCustomDomain(t *testing.T) {
	var bufErrStream bytes.Buffer
	errStream = &bufErrStream

	var c, err = Read("mocks/little-legacy")

	if err != nil {
		t.Errorf("Expected no error, got %v instead", err)
	}

	jsonlib.AssertJSONMarshal(t, tdata.FromFile(
		"mocks/little-legacy/project_ref.json"),
		c)

	var wantDeprecation = `DEPRECATED: CustomDomain string is now CustomDomains []string
Update your project.json to use:
"customDomains": ["foo.com"] instead of "customDomain": "foo.com".`

	if !strings.Contains(bufErrStream.String(), wantDeprecation) {
		t.Errorf("Wanted deprecation info not available")
	}

	errStream = defaultErrStream
}

func TestReadLegacyCustomDomainError(t *testing.T) {
	var bufErrStream bytes.Buffer
	errStream = &bufErrStream

	var _, err = Read("mocks/little-legacy-issue")

	var wantErr = "Can't use both customDomains and deprecated customDomain on project.json"

	if err == nil || err.Error() != wantErr {
		t.Errorf("Expected error %v, got %v instead", wantErr, err)
	}

	var wantDeprecation = `DEPRECATED: CustomDomain string is now CustomDomains []string
Update your project.json to use:
"customDomains": ["foo.com"] instead of "customDomain": "foo.com".`

	if !strings.Contains(bufErrStream.String(), wantDeprecation) {
		t.Errorf("Wanted deprecation info not available")
	}

	errStream = defaultErrStream
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

	servertest.Mux.HandleFunc("/restart/project", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.RawQuery != "projectId=foo" {
			t.Error("Wrong query parameters for restart method")
		}

		fmt.Fprintf(w, `"on"`)
	})

	if err := Restart(context.Background(), "foo"); err != nil {
		t.Errorf("Unexpected error on project restart: %v", err)
	}

	configmock.Teardown()
}

func TestUnlink(t *testing.T) {
	servertest.Setup()
	configmock.Setup()

	servertest.Mux.HandleFunc("/projects/foo", func(w http.ResponseWriter, r *http.Request) {
		var wantMethod = "DELETE"
		if r.Method != wantMethod {
			t.Errorf("Wanted method %v, got %v instead", wantMethod, r.Method)
		}
	})

	var err = Unlink(context.Background(), "foo")

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

	if err := Validate(context.Background(), "foo"); err != nil {
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

	if err := Validate(context.Background(), "foo"); err != ErrProjectAlreadyExists {
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

	if err := Validate(context.Background(), "foo"); err != ErrInvalidProjectID {
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

	var err = Validate(context.Background(), "foo")

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

	var err = Validate(context.Background(), "foo")

	if err != apihelper.ErrInvalidContentType {
		t.Errorf("Expected content-type error didn't happen")
	}

	servertest.Teardown()
	configmock.Teardown()
}

func TestValidateOrCreate(t *testing.T) {
	servertest.Setup()
	configmock.Setup()

	servertest.Mux.HandleFunc("/projects",
		func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-type", "application/json; charset=UTF-8")
			fmt.Fprintf(w, tdata.FromFile("mocks/new_response.json"))
		})

	if fid, err := ValidateOrCreate("tesla36"); fid != "tesla36" || err != nil {
		t.Errorf("Wanted (%v, %v), got (%v, %v) instead", "tesla36", nil, fid, err)
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

	if fid, err := ValidateOrCreate("little"); fid != "little" || err != nil {
		t.Errorf("Wanted (%v, %v), got (%v, %v) instead", "", nil, fid, err)
	}

	servertest.Teardown()
	configmock.Teardown()
}

func TestValidateOrCreateInvalidError(t *testing.T) {
	servertest.Setup()
	configmock.Setup()

	var _, err = ValidateOrCreate("little")

	if err == nil {
		t.Errorf("Expected error, got %v instead", err)
	}

	servertest.Teardown()
	configmock.Teardown()
}

func TestValidateOrCreateFromJSONAlreadyExists(t *testing.T) {
	servertest.Setup()
	configmock.Setup()

	servertest.Mux.HandleFunc("/projects",
		func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-type", "application/json; charset=UTF-8")
			w.WriteHeader(400)
			fmt.Fprintf(w, tdata.FromFile("mocks/create_already_exists_response.json"))
		})

	if ok, err := ValidateOrCreateFromJSON("mocks/little/project.json"); ok != false || err != nil {
		t.Errorf("Wanted (%v, %v), got (%v, %v) instead", false, nil, ok, err)
	}

	servertest.Teardown()
	configmock.Teardown()
}

func TestValidateOrCreateFromJSONNotExists(t *testing.T) {
	servertest.Setup()
	configmock.Setup()

	servertest.Mux.HandleFunc("/projects",
		func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "POST" {
				t.Errorf("Unexpected method %v", r.Method)
			}
		})

	var ok, err = ValidateOrCreateFromJSON("mocks/little/project.json")

	if ok != true || err != nil {
		t.Errorf("Unexpected error on Install: (%v, %v)", ok, err)
	}

	servertest.Teardown()
	configmock.Teardown()
}

func TestValidateOrCreateFromJSONInvalidError(t *testing.T) {
	servertest.Setup()
	configmock.Setup()

	var _, err = ValidateOrCreateFromJSON("mocks/little/project.json")

	if err == nil {
		t.Errorf("Expected error, got %v instead", err)
	}

	servertest.Teardown()
	configmock.Teardown()
}
