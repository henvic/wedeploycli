package projects

import (
	"context"
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

func TestAddDomain(t *testing.T) {
	t.Skipf("Skipping until https://github.com/wedeploy/cli/issues/187 is closed")
	servertest.Setup()
	configmock.Setup()

	servertest.Mux.HandleFunc("/projects/foo/customDomains",
		func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPatch {
				t.Errorf("Expected method %v, got %v instead", http.MethodPatch, r.Method)
			}

			var body, err = ioutil.ReadAll(r.Body)

			if err != nil {
				t.Errorf("Error parsing response")
			}

			var wantBody = `"example.com"`

			if string(body) != wantBody {
				t.Errorf("Wanted body to be %v, got %v instead", wantBody, string(body))
			}
		})

	if err := AddDomain(context.Background(), "foo", "example.com"); err != nil {
		t.Errorf("Expected no error when adding domains, got %v instead", err)
	}

	servertest.Teardown()
	configmock.Teardown()
}

func TestRemoveDomain(t *testing.T) {
	t.Skipf("Skipping until https://github.com/wedeploy/cli/issues/187 is closed")
	servertest.Setup()
	configmock.Setup()

	servertest.Mux.HandleFunc("/projects/foo/customDomains/example.com",
		func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodDelete {
				t.Errorf("Expected method %v, got %v instead", http.MethodDelete, r.Method)
			}
		})

	if err := RemoveDomain(context.Background(), "foo", "example.com"); err != nil {
		t.Errorf("Expected no error when adding domains, got %v instead", err)
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

func TestGetEmpty(t *testing.T) {
	var _, err = Get(context.Background(), "")

	if err != ErrEmptyProjectID {
		t.Errorf("Wanted error to be %v, got %v instead", ErrEmptyProjectID, err)
	}
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
	var c, err = Read("mocks/little-legacy")

	jsonlib.AssertJSONMarshal(t, tdata.FromFile(
		"mocks/little-legacy/project_ref.json"),
		c)

	if err == nil {
		t.Fatalf("Expected error not to be null")
	}

	var wantErr = `CustomDomain string support was removed in favor of CustomDomains []string
Update your mocks/little-legacy/project.json file to use:
"customDomains": ["foo.com"] instead of "customDomain": "foo.com".`

	if err.Error() != wantErr {
		t.Errorf("Wanted err to be %v, got %v instead", wantErr, err)
	}
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
