package projects

import (
	"context"
	"net/http"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/wedeploy/cli/apihelper"
	"github.com/wedeploy/cli/config"
	"github.com/wedeploy/cli/defaults"
	"github.com/wedeploy/cli/servertest"
	"github.com/wedeploy/cli/services"
	"github.com/wedeploy/cli/tdata"
)

var (
	wectx  config.Context
	client *Client
)

func TestMain(m *testing.M) {
	var err error
	wectx, err = config.Setup("mocks/.we")

	if err != nil {
		panic(err)
	}

	if err := wectx.SetEndpoint(defaults.CloudRemote); err != nil {
		panic(err)
	}

	client = New(wectx)
	os.Exit(m.Run())
}

func TestProjectCreatedAtTimeHelper(t *testing.T) {
	var s = Project{
		ProjectID: "abc",
		CreatedAt: 1517599604871,
	}

	var got = s.CreatedAtTime()
	var want = time.Date(2018, time.February, 2, 19, 26, 44, 0, time.UTC)

	if !got.Equal(want) {
		t.Errorf("Expected time didn't match: wanted %v, got %v", want, got)
	}
}

func TestCreate(t *testing.T) {
	servertest.Setup()

	servertest.Mux.HandleFunc("/projects",
		tdata.ServerJSONFileHandler("mocks/new_response.json"))

	var project, err = client.Create(context.Background(), Project{})

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

	var project, err = client.Create(context.Background(),
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

	var _, err = client.Create(context.Background(), Project{})

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

	var list, err = client.Get(context.Background(), "images")

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

func TestGetWithServices(t *testing.T) {
	servertest.Setup()

	servertest.Mux.HandleFunc(
		"/projects/images",
		tdata.ServerJSONFileHandler("mocks/project_get_response_with_services.json"))

	var list, err = client.Get(context.Background(), "images")

	var want = Project{
		ProjectID: "images",
		Health:    "on",
		Services: services.Services{
			services.Service{
				ServiceID: "hi",
			},
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

func TestGetEmpty(t *testing.T) {
	var _, err = client.Get(context.Background(), "")

	if err != ErrEmptyProjectID {
		t.Errorf("Wanted error to be %v, got %v instead", ErrEmptyProjectID, err)
	}
}

func TestList(t *testing.T) {
	servertest.Setup()

	servertest.Mux.HandleFunc(
		"/projects",
		tdata.ServerJSONFileHandler("mocks/projects_response.json"))

	var list, err = client.List(context.Background())

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

func TestListWithServices(t *testing.T) {
	servertest.Setup()

	servertest.Mux.HandleFunc(
		"/projects",
		tdata.ServerJSONFileHandler("mocks/projects_response_with_services.json"))

	var list, err = client.List(context.Background())

	var want = []Project{
		Project{
			ProjectID: "images",
			Health:    "on",
			Services: services.Services{
				services.Service{
					ServiceID: "hi",
				},
			},
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

func TestDelete(t *testing.T) {
	servertest.Setup()

	servertest.Mux.HandleFunc("/projects/foo", func(w http.ResponseWriter, r *http.Request) {
		var wantMethod = "DELETE"
		if r.Method != wantMethod {
			t.Errorf("Wanted method %v, got %v instead", wantMethod, r.Method)
		}
	})

	var err = client.Delete(context.Background(), "foo")

	if err != nil {
		t.Errorf("Expected no error, got %v instead", err)
	}

	servertest.Teardown()
}

func TestGetDeploymentOrder(t *testing.T) {
	servertest.Setup()

	var want = []string{"data", "auth", "hosting"}

	servertest.Mux.HandleFunc("/projects/foo/builds/order/xyz",
		tdata.ServerJSONFileHandler("mocks/deployment_order_response.json"))

	var got, err = client.GetDeploymentOrder(context.Background(), "foo", "xyz")

	if err != nil {
		t.Errorf("Expected no error, got %v instead", err)
	}

	if !reflect.DeepEqual(want, got) {
		t.Errorf("Wanted order to be %v, got %v instead", want, got)
	}

	servertest.Teardown()
}
