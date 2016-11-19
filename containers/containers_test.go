package containers

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/hashicorp/errwrap"
	"github.com/kylelemons/godebug/pretty"
	"github.com/wedeploy/api-go/jsonlib"
	"github.com/wedeploy/cli/apihelper"
	"github.com/wedeploy/cli/configmock"
	"github.com/wedeploy/cli/servertest"
	"github.com/wedeploy/cli/tdata"
)

func TestGetListFromDirectory(t *testing.T) {
	var containers, err = GetListFromDirectory("mocks/app")

	if err != nil {
		t.Errorf("Expected %v, got %v instead", nil, err)
	}

	var wantIDs = []string{"email", "landing", "speaker"}

	if !reflect.DeepEqual(containers.GetIDs(), wantIDs) {
		t.Errorf("Want %v, got %v instead", wantIDs, containers)
	}

	var wantLocations = []string{"email", "landing", "speaker"}

	if !reflect.DeepEqual(containers.GetLocations(), wantLocations) {
		t.Errorf("Want %v, got %v instead", wantLocations, containers)
	}

	var want = ContainerInfoList{
		ContainerInfo{
			"email",
			"email",
		},
		ContainerInfo{
			"landing",
			"landing",
		},
		ContainerInfo{
			"speaker",
			"speaker",
		},
	}

	if !reflect.DeepEqual(containers, want) {
		t.Errorf("Want %v, got %v instead", want, containers)
	}
}

func TestGetListFromDirectoryOnProjectWithContainersInsideSubdirectories(t *testing.T) {
	var containers, err = GetListFromDirectory("mocks/project-with-containers-inside-subdirs")

	if err != nil {
		t.Errorf("Expected %v, got %v instead", nil, err)
	}

	var wantIDs = []string{"level-1", "level-2", "level-2-2", "level-3"}

	if !reflect.DeepEqual(containers.GetIDs(), wantIDs) {
		t.Errorf("Want %v, got %v instead", wantIDs, containers)
	}

	for _, id := range containers.GetIDs() {
		if id == "unreachable-container" ||
			id == "skipped-unreachable-container" ||
			id == "same-dir-skipped-unreachable-container" {
			t.Errorf("%v should be unreachable", id)
		}
	}

	var wantLocations = []string{
		"container-level-1",
		"sub-dir/container-level-2",
		"sub-dir-2/container-level-2-2",
		"sub-dir-2/sub-dir/container-level-3",
	}

	if !reflect.DeepEqual(containers.GetLocations(), wantLocations) {
		t.Errorf("Want %v, got %v instead", wantLocations, containers)
	}

	var want = ContainerInfoList{
		ContainerInfo{
			"level-1",
			"container-level-1",
		},
		ContainerInfo{
			"level-2",
			"sub-dir/container-level-2",
		},
		ContainerInfo{
			"level-2-2",
			"sub-dir-2/container-level-2-2",
		},
		ContainerInfo{
			"level-3",
			"sub-dir-2/sub-dir/container-level-3",
		},
	}

	if !reflect.DeepEqual(containers, want) {
		t.Errorf("Want %v, got %v instead", want, containers)
	}

	// noWhat variable just to avoid a replace all messing things up
	// mocks/project-with-containers-inside-subdirs/sub-dir-2/same-dir-skipped-unreachable-container
	// also tries to fail if that happens
	var noWhat = "container"
	if _, err = os.Stat("mocks/project-with-containers-inside-subdirs/sub-dir-2/skip/.no" + noWhat); err != nil {
		t.Fatalf(`.nocontainer not found: filepath.Walk might fail given that we are counting on its lexical order walk`)
	}
}

func TestGetListFromDirectoryDuplicateID(t *testing.T) {
	var containers, err = GetListFromDirectory("mocks/project-with-duplicate-containers-ids")

	if len(containers) != 0 {
		t.Errorf("Expected containers length to be 0 on error.")
	}

	if err == nil {
		t.Errorf("Expected error, got %v instead.", err)
	}

	var wantErr = fmt.Sprintf(`Can't list containers: ID "email" was found duplicated on containers %v and %v`,
		abs("./mocks/project-with-duplicate-containers-ids/one"),
		abs("./mocks/project-with-duplicate-containers-ids/two"))

	if err.Error() != wantErr {
		t.Errorf("Expected error message to be %v, got %v instead", wantErr, err)
	}
}

func TestGetListFromDirectoryInvalid(t *testing.T) {
	var containers, err = GetListFromDirectory("mocks/app-with-invalid-container")

	if containers != nil {
		t.Errorf("Expected containers to be nil, got %v instead", containers)
	}

	if err == nil || os.IsNotExist(err) {
		t.Errorf("Expected error %v to be due to invalid config", err)
	}
}

func TestGetListFromDirectoryNotExists(t *testing.T) {
	var containers, err = GetListFromDirectory(fmt.Sprintf("not-found-%d", rand.Int()))

	if containers != nil {
		t.Errorf("Expected containers to be nil, got %v instead", containers)
	}

	if !os.IsNotExist(err) {
		t.Errorf("Expected error %v to be due to file not found", err)
	}
}

func TestGet(t *testing.T) {
	servertest.Setup()
	configmock.Setup()

	var want = Container{
		ID:     "search7606",
		Health: "on",
		Type:   "cloudsearch",
		Scale:  7,
	}

	servertest.Mux.HandleFunc("/projects/images/containers/search7606",
		tdata.ServerJSONFileHandler("mocks/container_response.json"))

	var got, err = Get(context.Background(), "images", "search7606")

	if err != nil {
		t.Errorf("Expected no error, got %v instead", err)
	}

	if !reflect.DeepEqual(want, got) {
		t.Errorf("Get doesn't match with wanted structure.")
		t.Errorf(pretty.Compare(want, got))
	}

	servertest.Teardown()
	configmock.Teardown()
}

func TestList(t *testing.T) {
	servertest.Setup()
	configmock.Setup()

	var want = Containers{
		"search7606": &Container{
			ID:     "search7606",
			Health: "on",
			Type:   "cloudsearch",
			Scale:  7,
		},
		"nodejs5143": &Container{
			ID:     "nodejs5143",
			Health: "on",
			Type:   "nodejs",
			Scale:  5,
		},
	}

	servertest.Mux.HandleFunc("/projects/images/containers",
		tdata.ServerJSONFileHandler("mocks/containers_response.json"))

	var got, err = List(context.Background(), "images")

	if err != nil {
		t.Errorf("Expected no error, got %v instead", err)
	}

	if !reflect.DeepEqual(want, got) {
		t.Errorf("List doesn't match with wanted structure.")
		t.Errorf(pretty.Compare(want, got))
	}

	servertest.Teardown()
	configmock.Teardown()
}

func TestLink(t *testing.T) {
	servertest.Setup()
	configmock.Setup()

	servertest.Mux.HandleFunc(
		"/deploy",
		func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "PUT" {
				t.Errorf("Expected install method to be PUT")
			}

			var qp = r.URL.Query()

			if qp.Get("projectId") != "sound" {
				t.Errorf("Missing expected value for projectId query param")
			}

			if qp.Get("containerId") != "speaker" {
				t.Errorf("Missing expected value for containerId query param")
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
				`{"id":"speaker", "type": "nodejs"}`,
				data)
		})

	var err = Link(context.Background(), "sound", "speaker", "mocks/app/speaker")

	if err != nil {
		t.Errorf("Unexpected error on Install: %v", err)
	}

	servertest.Teardown()
	configmock.Teardown()
}

func TestRegistry(t *testing.T) {
	servertest.Setup()
	configmock.Setup()

	servertest.Mux.HandleFunc(
		"/registry",
		tdata.ServerJSONFileHandler("mocks/registry.json"))

	var registry, err = GetRegistry(context.Background())

	if len(registry) != 7 {
		t.Errorf("Expected registry to have 7 images")
	}

	if err != nil {
		t.Errorf("Expected no error, got %v instead", err)
	}

	servertest.Teardown()
	configmock.Teardown()
}

func TestRead(t *testing.T) {
	var c, err = Read("mocks/app/email")

	if err != nil {
		t.Errorf("Expected no error, got %v instead", err)
	}

	jsonlib.AssertJSONMarshal(t, tdata.FromFile(
		"mocks/app/email/container_ref.json"),
		c)
}

func TestReadFileNotFound(t *testing.T) {
	var _, err = Read("mocks/app/unknown")

	if err != ErrContainerNotFound {
		t.Errorf("Expected %v, got %v instead", ErrContainerNotFound, err)
	}
}

func TestReadInvalidContainerID(t *testing.T) {
	var _, err = Read("mocks/app-for/missing-email-id")

	if err != ErrInvalidContainerID {
		t.Errorf("Expected %v, got %v instead", ErrInvalidContainerID, err)
	}
}

func TestReadCorrupted(t *testing.T) {
	var _, err = Read("mocks/app-with-invalid-container/corrupted")

	if _, ok := err.(*json.SyntaxError); !ok {
		t.Errorf("Wanted err to be *json.SyntaxError, got %v instead", err)
	}
}

func TestRestart(t *testing.T) {
	servertest.Setup()
	configmock.Setup()

	servertest.Mux.HandleFunc("/restart/container",
		func(w http.ResponseWriter, r *http.Request) {
			var p, err = url.ParseQuery(r.URL.RawQuery)

			if err != nil {
				panic(err)
			}

			if p.Get("projectId") != "foo" || p.Get("containerId") != "bar" {
				t.Errorf("Wrong query parameters, got %v", r.URL.RawQuery)
			}

			fmt.Fprintf(w, `"on"`)
		})

	if err := Restart(context.Background(), "foo", "bar"); err != nil {
		t.Errorf("Unexpected error on container restart: %v", err)
	}

	servertest.Teardown()
	configmock.Teardown()
}

func TestUnlink(t *testing.T) {
	servertest.Setup()
	configmock.Setup()

	servertest.Mux.HandleFunc("/deploy/foo/bar", func(w http.ResponseWriter, r *http.Request) {
		var wantMethod = "DELETE"
		if r.Method != wantMethod {
			t.Errorf("Wanted method %v, got %v instead", wantMethod, r.Method)
		}
	})

	var err = Unlink(context.Background(), "foo", "bar")

	if err != nil {
		t.Errorf("Expected no error, got %v instead", err)
	}

	servertest.Teardown()
	configmock.Teardown()
}

func TestValidate(t *testing.T) {
	servertest.Setup()
	configmock.Setup()

	servertest.Mux.HandleFunc("/validators/containers/id",
		func(w http.ResponseWriter, r *http.Request) {
			if r.FormValue("projectId") != "foo" {
				t.Errorf("Wrong projectId form value")
			}

			if r.FormValue("value") != "bar" {
				t.Errorf("Wrong containerId form value")
			}
		})

	if err := Validate(context.Background(), "foo", "bar"); err != nil {
		t.Errorf("Wanted null error, got %v instead", err)
	}

	servertest.Teardown()
	configmock.Teardown()
}

func TestValidateAlreadyExists(t *testing.T) {
	servertest.Setup()
	configmock.Setup()

	servertest.Mux.HandleFunc("/validators/containers/id",
		func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-type", "application/json; charset=UTF-8")
			w.WriteHeader(404)
			fmt.Fprintf(w, tdata.FromFile("mocks/container_already_exists_response.json"))
		})

	if err := Validate(context.Background(), "foo", "bar"); err != ErrContainerAlreadyExists {
		t.Errorf("Wanted %v error, got %v instead", ErrContainerAlreadyExists, err)
	}

	servertest.Teardown()
	configmock.Teardown()
}

func TestValidateInvalidID(t *testing.T) {
	servertest.Setup()
	configmock.Setup()

	servertest.Mux.HandleFunc("/validators/containers/id",
		func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-type", "application/json; charset=UTF-8")
			w.WriteHeader(404)
			fmt.Fprintf(w, tdata.FromFile("mocks/container_invalid_id_response.json"))
		})

	if err := Validate(context.Background(), "foo", "bar"); err != ErrInvalidContainerID {
		t.Errorf("Wanted %v error, got %v instead", ErrInvalidContainerID, err)
	}

	servertest.Teardown()
	configmock.Teardown()
}

func TestValidateError(t *testing.T) {
	servertest.Setup()
	configmock.Setup()

	servertest.Mux.HandleFunc("/validators/containers/id",
		func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-type", "application/json; charset=UTF-8")
			w.WriteHeader(400)
			fmt.Fprintf(w, tdata.FromFile("../apihelper/mocks/unknown_error_api_response.json"))
		})

	var err = Validate(context.Background(), "foo", "bar")

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

	servertest.Mux.HandleFunc("/validators/containers/id",
		func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-type", "application/json; charset=UTF-8")
			w.WriteHeader(404)
		})

	var err = Validate(context.Background(), "foo", "bar")

	if err == nil || errwrap.Get(err, "unexpected end of JSON input") == nil {
		t.Errorf("Expected error didn't happen, got %v instead", err)
	}

	servertest.Teardown()
	configmock.Teardown()
}

type TestNormalizePathToUnixProvider struct {
	in   string
	want string
}

var TestNormalizePathToUnixCases = []TestNormalizePathToUnixProvider{
	TestNormalizePathToUnixProvider{`C:\`, `/C`},
	TestNormalizePathToUnixProvider{`C:\foobar`, `/C/foobar`},
	TestNormalizePathToUnixProvider{`/`, `/`},
	TestNormalizePathToUnixProvider{`/home`, `/home`},
	TestNormalizePathToUnixProvider{`/home/user`, `/home/user`},
	TestNormalizePathToUnixProvider{`/Users/user`, `/Users/user`},
	TestNormalizePathToUnixProvider{`Z:\foo\bar`, `/Z/foo/bar`},
}

func TestNormalizePathToUnix(t *testing.T) {
	for _, c := range TestNormalizePathToUnixCases {
		if got := normalizePathToUnix(c.in); got != c.want {
			t.Errorf("Expected %v, got %v instead", c.want, got)
		}
	}
}

func abs(path string) string {
	var abs, err = filepath.Abs(path)

	if err != nil {
		panic(err)
	}

	return abs
}
