package containers

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/hashicorp/errwrap"
	"github.com/kylelemons/godebug/pretty"
	"github.com/wedeploy/api-go/jsonlib"
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

	config.SetEndpointContext(defaults.LocalRemote)
	os.Exit(m.Run())
}

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

	var wantErr = fmt.Sprintf(`Can not list containers: ID "email" was found duplicated on containers %v and %v`,
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

	var want = Container{
		ServiceID: "search7606",
		Health:    "on",
		Image:     "wedeploy/data",
		Scale:     7,
	}

	servertest.Mux.HandleFunc("/projects/images/services/search7606",
		tdata.ServerJSONFileHandler("mocks/container_response.json"))

	var got, err = Get(context.Background(), "images", "search7606")

	if err != nil {
		t.Errorf("Expected no error, got %v instead", err)
	}

	if !reflect.DeepEqual(want, got) {
		t.Errorf("Get does not match with wanted structure.")
		t.Errorf(pretty.Compare(want, got))
	}

	servertest.Teardown()
}

func TestGetEmptyContainerID(t *testing.T) {
	var _, err = Get(context.Background(), "images", "")

	if err != ErrEmptyContainerID {
		t.Errorf("Wanted error to be %v, got %v instead", ErrEmptyContainerID, err)
	}
}

func TestGetEmptyProjectID(t *testing.T) {
	var _, err = Get(context.Background(), "", "foo")

	if err != ErrEmptyProjectID {
		t.Errorf("Wanted error to be %v, got %v instead", ErrEmptyProjectID, err)
	}
}

func TestGetEmptyProjectAndContainerID(t *testing.T) {
	var _, err = Get(context.Background(), "", "")

	if err != ErrEmptyProjectAndContainerID {
		t.Errorf("Wanted error to be %v, got %v instead", ErrEmptyProjectAndContainerID, err)
	}
}

func TestList(t *testing.T) {
	servertest.Setup()

	var want = []Container{
		Container{
			ServiceID: "nodejs5143",
			Health:    "on",
			Scale:     5,
		},
		Container{
			ServiceID: "search7606",
			Health:    "on",
			Scale:     7,
		},
	}

	servertest.Mux.HandleFunc("/projects/images/services",
		tdata.ServerJSONFileHandler("mocks/containers_response.json"))

	var got, err = List(context.Background(), "images")

	if err != nil {
		t.Errorf("Expected no error, got %v instead", err)
	}

	// if !reflect.DeepEqual(want, got) {
	if fmt.Sprintf("%+v", got) != fmt.Sprintf("%+v", want) {
		t.Errorf("List does not match with wanted structure.")
		t.Errorf(pretty.Compare(want, got))
	}

	switch c, err := got.Get("search7606"); {
	case err != nil:
		t.Errorf("Expected no error filtering container, got %v instead", err)
	case c.ServiceID != "search7606":
		t.Errorf("Got wrong container ID: %+v", c)
	}

	switch c, err := got.Get("nodejs5143"); {
	case err != nil:
		t.Errorf("Expected no error filtering container, got %v instead", err)
	case c.ServiceID != "nodejs5143":
		t.Errorf("Got wrong container ID: %+v", c)
	}

	if _, err := got.Get("not_found"); err == nil {
		t.Errorf("Expected container not to be found, got %v instead", err)
	}

	servertest.Teardown()
}

func TestLink(t *testing.T) {
	servertest.Setup()

	servertest.Mux.HandleFunc(
		"/projects/sound/services",
		func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "POST" {
				t.Errorf("Expected install method to be PUT")
			}

			var body, err = ioutil.ReadAll(r.Body)

			if err != nil {
				t.Error(err)
			}

			var data map[string]json.RawMessage

			err = json.Unmarshal(body, &data)

			if err != nil {
				t.Error(err)
			}

			jsonlib.AssertJSONMarshal(t,
				`{"serviceId":"speaker", "scale": 1, "source": "mocks/app/speaker"}`,
				data)
		})

	var err = Link(context.Background(), "sound", Container{ServiceID: "speaker"}, "mocks/app/speaker")

	if err != nil {
		t.Errorf("Unexpected error on Install: %v", err)
	}

	servertest.Teardown()
}

func TestRegistry(t *testing.T) {
	servertest.Setup()

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

func TestSetEnvironmentVariable(t *testing.T) {
	t.Skipf("Skipping until https://github.com/wedeploy/cli/issues/186 is closed")
	servertest.Setup()

	servertest.Mux.HandleFunc("/projects/foo/services/bar/env/xyz",
		func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPut {
				t.Errorf("Expected method %v, got %v instead", http.MethodPut, r.Method)
			}

			var body, err = ioutil.ReadAll(r.Body)

			if err != nil {
				t.Errorf("Error parsing response")
			}

			var wantBody = `"abc"`

			if string(body) != wantBody {
				t.Errorf("Wanted body to be %v, got %v instead", wantBody, string(body))
			}
		})

	if err := SetEnvironmentVariable(context.Background(), "foo", "bar", "xyz", "abc"); err != nil {
		t.Errorf("Expected no error when adding domains, got %v instead", err)
	}

	servertest.Teardown()
}

func TestUnsetEnvironmentVariable(t *testing.T) {
	t.Skipf("Skipping until https://github.com/wedeploy/cli/issues/186 is closed")
	servertest.Setup()

	servertest.Mux.HandleFunc("/projects/foo/services/bar/env/xyz",
		func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodDelete {
				t.Errorf("Expected method %v, got %v instead", http.MethodDelete, r.Method)
			}
		})

	if err := UnsetEnvironmentVariable(context.Background(), "foo", "bar", "xyz"); err != nil {
		t.Errorf("Expected no error when adding domains, got %v instead", err)
	}

	servertest.Teardown()
}

func TestRestart(t *testing.T) {
	servertest.Setup()

	servertest.Mux.HandleFunc("/projects/foo/services/bar/restart",
		func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintf(w, `"on"`)
		})

	if err := Restart(context.Background(), "foo", "bar"); err != nil {
		t.Errorf("Unexpected error on container restart: %v", err)
	}

	servertest.Teardown()
}

func TestUnlink(t *testing.T) {
	servertest.Setup()

	servertest.Mux.HandleFunc("/projects/foo/services/bar", func(w http.ResponseWriter, r *http.Request) {
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
}

func TestValidate(t *testing.T) {
	servertest.Setup()

	servertest.Mux.HandleFunc("/validators/services/id",
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
}

func TestValidateAlreadyExists(t *testing.T) {
	servertest.Setup()

	servertest.Mux.HandleFunc("/validators/services/id",
		func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-type", "application/json; charset=UTF-8")
			w.WriteHeader(404)
			fmt.Fprintf(w, tdata.FromFile("mocks/container_already_exists_response.json"))
		})

	if err := Validate(context.Background(), "foo", "bar"); err != ErrContainerAlreadyExists {
		t.Errorf("Wanted %v error, got %v instead", ErrContainerAlreadyExists, err)
	}

	servertest.Teardown()
}

func TestValidateInvalidID(t *testing.T) {
	servertest.Setup()

	servertest.Mux.HandleFunc("/validators/services/id",
		func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-type", "application/json; charset=UTF-8")
			w.WriteHeader(404)
			fmt.Fprintf(w, tdata.FromFile("mocks/container_invalid_id_response.json"))
		})

	if err := Validate(context.Background(), "foo", "bar"); err != ErrInvalidContainerID {
		t.Errorf("Wanted %v error, got %v instead", ErrInvalidContainerID, err)
	}

	servertest.Teardown()
}

func TestValidateError(t *testing.T) {
	servertest.Setup()

	servertest.Mux.HandleFunc("/validators/services/id",
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
}

func TestValidateInvalidError(t *testing.T) {
	servertest.Setup()

	servertest.Mux.HandleFunc("/validators/services/id",
		func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-type", "application/json; charset=UTF-8")
			w.WriteHeader(404)
		})

	var err = Validate(context.Background(), "foo", "bar")

	if err == nil || errwrap.Get(err, "unexpected end of JSON input") == nil {
		t.Errorf("Expected error didn't happen, got %v instead", err)
	}

	servertest.Teardown()
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
