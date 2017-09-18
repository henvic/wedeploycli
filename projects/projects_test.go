package projects

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"reflect"
	"testing"

	"github.com/wedeploy/cli/apihelper"
	"github.com/wedeploy/cli/config"
	"github.com/wedeploy/cli/defaults"
	"github.com/wedeploy/cli/servertest"
	"github.com/wedeploy/cli/tdata"
)

func TestMain(m *testing.M) {
	if err := config.Setup("mocks/.we"); err != nil {
		panic(err)
	}

	if err := config.SetEndpointContext(defaults.CloudRemote); err != nil {
		panic(err)
	}

	os.Exit(m.Run())
}

func TestCreate(t *testing.T) {
	servertest.Setup()

	servertest.Mux.HandleFunc("/projects",
		tdata.ServerJSONFileHandler("mocks/new_response.json"))

	var project, err = Create(context.Background(), Project{})

	if project.ProjectID != "tesla36" {
		t.Errorf("Wanted project ID to be tesla36, got %v instead", project.ProjectID)
	}

	if project.Health != "on" {
		t.Errorf("Wanted project Health to be on, got %v instead", project.Health)
	}

	if err != nil {
		t.Errorf("Wanted error to be nil, got %v instead", err)
	}

	servertest.Teardown()
}

func TestCreateNamed(t *testing.T) {
	servertest.Setup()

	servertest.Mux.HandleFunc("/projects",
		tdata.ServerJSONFileHandler("mocks/new_named_response.json"))

	var project, err = Create(context.Background(),
		Project{
			ProjectID: "banach30",
		})

	if project.ProjectID != "banach30" {
		t.Errorf("Wanted project ID to be banach30, got %v instead", project.ProjectID)
	}

	if err != nil {
		t.Errorf("Wanted error to be nil, got %v instead", err)
	}

	servertest.Teardown()
}

func TestCreateError(t *testing.T) {
	servertest.Setup()

	var _, err = Create(context.Background(), Project{})

	switch err.(type) {
	case *apihelper.APIFault:
	default:
		t.Errorf("Wanted APIFault error, got %v instead", err)
	}

	servertest.Teardown()
}

func TestGet(t *testing.T) {
	servertest.Setup()

	servertest.Mux.HandleFunc(
		"/projects/images",
		tdata.ServerJSONFileHandler("mocks/project_get_response.json"))

	var list, err = Get(context.Background(), "images")

	var want = Project{
		ProjectID: "images",
		Health:    "on",
	}

	if !reflect.DeepEqual(want, list) {
		t.Errorf("Wanted %v, got %v instead", want, list)
	}

	if err != nil {
		t.Errorf("Wanted error to be nil, got %v instead", err)
	}

	servertest.Teardown()
}

func TestGetEmpty(t *testing.T) {
	var _, err = Get(context.Background(), "")

	if err != ErrEmptyProjectID {
		t.Errorf("Wanted error to be %v, got %v instead", ErrEmptyProjectID, err)
	}
}

func TestList(t *testing.T) {
	servertest.Setup()

	servertest.Mux.HandleFunc(
		"/projects",
		tdata.ServerJSONFileHandler("mocks/projects_response.json"))

	var list, err = List(context.Background())

	var want = []Project{
		Project{
			ProjectID: "images",
			Health:    "on",
		},
	}

	if !reflect.DeepEqual(want, list) {
		t.Errorf("Wanted %v, got %v instead", want, list)
	}

	if err != nil {
		t.Errorf("Wanted error to be nil, got %v instead", err)
	}

	servertest.Teardown()
}

func TestRestart(t *testing.T) {
	defer servertest.Teardown()
	servertest.Setup()

	servertest.Mux.HandleFunc("/projects/foo/restart", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, `"on"`)
	})

	if err := Restart(context.Background(), "foo"); err != nil {
		t.Errorf("Unexpected error on project restart: %v", err)
	}
}

func TestUnlink(t *testing.T) {
	servertest.Setup()

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
}
