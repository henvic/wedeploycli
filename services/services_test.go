package services

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

	if err := config.SetEndpointContext(defaults.LocalRemote); err != nil {
		panic(err)
	}

	os.Exit(m.Run())
}

func TestGetListFromDirectory(t *testing.T) {
	var services, err = GetListFromDirectory("mocks/app")

	if err != nil {
		t.Errorf("Expected %v, got %v instead", nil, err)
	}

	var wantIDs = []string{"email", "landing", "speaker"}

	if !reflect.DeepEqual(services.GetIDs(), wantIDs) {
		t.Errorf("Want %v, got %v instead", wantIDs, services)
	}

	var wantLocations = []string{"email", "landing", "speaker"}

	if !reflect.DeepEqual(services.GetLocations(), wantLocations) {
		t.Errorf("Want %v, got %v instead", wantLocations, services)
	}

	var want = ServiceInfoList{
		ServiceInfo{
			"email",
			"email",
		},
		ServiceInfo{
			"landing",
			"landing",
		},
		ServiceInfo{
			"speaker",
			"speaker",
		},
	}

	if !reflect.DeepEqual(services, want) {
		t.Errorf("Want %v, got %v instead", want, services)
	}

	speakerCI, speakerErr := services.Get("speaker")

	if speakerErr != nil {
		t.Errorf("Wanted speakerErr to be nil, got %v instead", speakerErr)
	}

	if speakerCI.Location != "speaker" || speakerCI.ServiceID != "speaker" {
		t.Errorf("speakerCI is not what was expected: %+v instead", speakerCI)
	}

	_, notFoundErr := services.Get("notfound")

	if notFoundErr == nil || notFoundErr.Error() != "found no service matching ID notfound locally" {
		t.Errorf("Expected not found error, got %v instead", err)
	}
}

func TestGetListFromDirectoryOnProjectWithServicesInsideSubdirectories(t *testing.T) {
	var services, err = GetListFromDirectory("mocks/project-with-services-inside-subdirs")

	if err != nil {
		t.Errorf("Expected %v, got %v instead", nil, err)
	}

	var wantIDs = []string{"level-1", "level-2", "level-2-2", "level-3"}

	if !reflect.DeepEqual(services.GetIDs(), wantIDs) {
		t.Errorf("Want %v, got %v instead", wantIDs, services)
	}

	for _, id := range services.GetIDs() {
		if id == "unreachable-service" ||
			id == "skipped-unreachable-service" ||
			id == "same-dir-skipped-unreachable-service" {
			t.Errorf("%v should be unreachable", id)
		}
	}

	var wantLocations = []string{
		"service-level-1",
		"sub-dir/service-level-2",
		"sub-dir-2/service-level-2-2",
		"sub-dir-2/sub-dir/service-level-3",
	}

	if !reflect.DeepEqual(services.GetLocations(), wantLocations) {
		t.Errorf("Want %v, got %v instead", wantLocations, services)
	}

	var want = ServiceInfoList{
		ServiceInfo{
			"level-1",
			"service-level-1",
		},
		ServiceInfo{
			"level-2",
			"sub-dir/service-level-2",
		},
		ServiceInfo{
			"level-2-2",
			"sub-dir-2/service-level-2-2",
		},
		ServiceInfo{
			"level-3",
			"sub-dir-2/sub-dir/service-level-3",
		},
	}

	if !reflect.DeepEqual(services, want) {
		t.Errorf("Want %v, got %v instead", want, services)
	}

	// noWhat variable just to avoid a replace all messing things up
	// mocks/project-with-services-inside-subdirs/sub-dir-2/same-dir-skipped-unreachable-service
	// also tries to fail if that happens
	var noWhat = "service"
	if _, err = os.Stat("mocks/project-with-services-inside-subdirs/sub-dir-2/skip/.no" + noWhat); err != nil {
		t.Fatalf(`.noservice not found: filepath.Walk might fail given that we are counting on its lexical order walk`)
	}
}

func TestGetListFromDirectoryDuplicateID(t *testing.T) {
	var services, err = GetListFromDirectory("mocks/project-with-duplicate-services-ids")

	if len(services) != 0 {
		t.Errorf("Expected services length to be 0 on error.")
	}

	if err == nil {
		t.Errorf("Expected error, got %v instead.", err)
	}

	var wantErr = fmt.Sprintf(`Can not list services: ID "email" was found duplicated on services %v and %v`,
		abs("./mocks/project-with-duplicate-services-ids/one"),
		abs("./mocks/project-with-duplicate-services-ids/two"))

	if err.Error() != wantErr {
		t.Errorf("Expected error message to be %v, got %v instead", wantErr, err)
	}
}

func TestGetListFromDirectoryInvalid(t *testing.T) {
	var services, err = GetListFromDirectory("mocks/app-with-invalid-service")

	if services != nil {
		t.Errorf("Expected services to be nil, got %v instead", services)
	}

	if err == nil || os.IsNotExist(err) {
		t.Errorf("Expected error %v to be due to invalid config", err)
	}
}

func TestGetListFromDirectoryNotExists(t *testing.T) {
	var services, err = GetListFromDirectory(fmt.Sprintf("not-found-%d", rand.Int()))

	if services != nil {
		t.Errorf("Expected services to be nil, got %v instead", services)
	}

	if !os.IsNotExist(err) {
		t.Errorf("Expected error %v to be due to file not found", err)
	}
}

func TestGet(t *testing.T) {
	servertest.Setup()

	var want = Service{
		ServiceID: "search7606",
		Health:    "on",
		Image:     "wedeploy/data",
		Scale:     7,
	}

	servertest.Mux.HandleFunc("/projects/images/services/search7606",
		tdata.ServerJSONFileHandler("mocks/service_response.json"))

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

func TestGetEmptyServiceID(t *testing.T) {
	var _, err = Get(context.Background(), "images", "")

	if err != ErrEmptyServiceID {
		t.Errorf("Wanted error to be %v, got %v instead", ErrEmptyServiceID, err)
	}
}

func TestGetEmptyProjectID(t *testing.T) {
	var _, err = Get(context.Background(), "", "foo")

	if err != ErrEmptyProjectID {
		t.Errorf("Wanted error to be %v, got %v instead", ErrEmptyProjectID, err)
	}
}

func TestGetEmptyProjectAndServiceID(t *testing.T) {
	var _, err = Get(context.Background(), "", "")

	if err != ErrEmptyProjectAndServiceID {
		t.Errorf("Wanted error to be %v, got %v instead", ErrEmptyProjectAndServiceID, err)
	}
}

func TestList(t *testing.T) {
	servertest.Setup()

	var want = []Service{
		Service{
			ServiceID: "nodejs5143",
			Health:    "on",
			Scale:     5,
		},
		Service{
			ServiceID: "search7606",
			Health:    "on",
			Scale:     7,
		},
	}

	servertest.Mux.HandleFunc("/projects/images/services",
		tdata.ServerJSONFileHandler("mocks/services_response.json"))

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
		t.Errorf("Expected no error filtering service, got %v instead", err)
	case c.ServiceID != "search7606":
		t.Errorf("Got wrong service ID: %+v", c)
	}

	switch c, err := got.Get("nodejs5143"); {
	case err != nil:
		t.Errorf("Expected no error filtering service, got %v instead", err)
	case c.ServiceID != "nodejs5143":
		t.Errorf("Got wrong service ID: %+v", c)
	}

	if _, err := got.Get("not_found"); err == nil {
		t.Errorf("Expected service not to be found, got %v instead", err)
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

	var err = Link(context.Background(), "sound", Service{ServiceID: "speaker"}, "mocks/app/speaker")

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
		"mocks/app/email/service_ref.json"),
		c)
}

func TestReadFileNotFound(t *testing.T) {
	var _, err = Read("mocks/app/unknown")

	if err != ErrServiceNotFound {
		t.Errorf("Expected %v, got %v instead", ErrServiceNotFound, err)
	}
}

func TestReadInvalidServiceID(t *testing.T) {
	var _, err = Read("mocks/app-for/missing-email-id")

	if err != ErrInvalidServiceID {
		t.Errorf("Expected %v, got %v instead", ErrInvalidServiceID, err)
	}
}

func TestReadCorrupted(t *testing.T) {
	var _, err = Read("mocks/app-with-invalid-service/corrupted")

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
		t.Errorf("Unexpected error on service restart: %v", err)
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
				t.Errorf("Wrong serviceId form value")
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
			fmt.Fprintf(w, tdata.FromFile("mocks/service_already_exists_response.json"))
		})

	if err := Validate(context.Background(), "foo", "bar"); err != ErrServiceAlreadyExists {
		t.Errorf("Wanted %v error, got %v instead", ErrServiceAlreadyExists, err)
	}

	servertest.Teardown()
}

func TestValidateInvalidID(t *testing.T) {
	servertest.Setup()

	servertest.Mux.HandleFunc("/validators/services/id",
		func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-type", "application/json; charset=UTF-8")
			w.WriteHeader(404)
			fmt.Fprintf(w, tdata.FromFile("mocks/service_invalid_id_response.json"))
		})

	if err := Validate(context.Background(), "foo", "bar"); err != ErrInvalidServiceID {
		t.Errorf("Wanted %v error, got %v instead", ErrInvalidServiceID, err)
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