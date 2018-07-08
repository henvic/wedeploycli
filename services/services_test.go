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
	"time"

	"github.com/hashicorp/errwrap"
	"github.com/kylelemons/godebug/pretty"
	"github.com/wedeploy/cli/apihelper"
	"github.com/wedeploy/cli/config"
	"github.com/wedeploy/cli/defaults"
	"github.com/wedeploy/cli/servertest"
	"github.com/wedeploy/cli/tdata"
	"github.com/wedeploy/wedeploy-sdk-go/jsonlib"
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

func TestServiceCreatedAtTimeHelper(t *testing.T) {
	var s = Service{
		ServiceID: "abc",
		CreatedAt: 1517599604871,
	}

	var got = s.CreatedAtTime()
	var want = time.Date(2018, time.February, 2, 19, 26, 44, 0, time.UTC)

	if !got.Equal(want) {
		t.Errorf("Expected time didn't match: wanted %v, got %v", want, got)
	}
}

func TestServiceTypeHelper(t *testing.T) {
	var s = Service{
		ServiceID: "abc",
		Image:     "off",
	}

	var want = "off"
	var got = s.Type()

	if want != got {
		t.Errorf("Expected type to be shown as %v, got %v instead", want, got)
	}
}

func TestServiceTypeHelperHint(t *testing.T) {
	var s = Service{
		ServiceID: "abc",
		Image:     "off",
		ImageHint: "on",
	}

	var want = "on"
	var got = s.Type()

	if want != got {
		t.Errorf("Expected type to be shown as %v, got %v instead", want, got)
	}
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

	var wantLocations = []string{
		abs("mocks/app/email"),
		abs("mocks/app/landing"),
		abs("mocks/app/speaker"),
	}

	if !reflect.DeepEqual(services.GetLocations(), wantLocations) {
		t.Errorf("Want %v, got %v instead", wantLocations, services)
	}

	var want = ServiceInfoList{
		ServiceInfo{
			"exampleProject",
			"email",
			abs("mocks/app/email"),
		},
		ServiceInfo{
			"exampleProject",
			"landing",
			abs("mocks/app/landing"),
		},
		ServiceInfo{
			"exampleProject",
			"speaker",
			abs("mocks/app/speaker"),
		},
	}

	if !reflect.DeepEqual(services, want) {
		t.Errorf("Want %v, got %v instead", want, services)
	}

	speakerCI, speakerErr := services.Get("speaker")

	if speakerErr != nil {
		t.Errorf("Wanted speakerErr to be nil, got %v instead", speakerErr)
	}

	if speakerCI.Location != abs("mocks/app/speaker") || speakerCI.ServiceID != "speaker" {
		t.Errorf("speaker is not what was expected: %+v instead", speakerCI)
	}

	_, notFoundErr := services.Get("notfound")

	if notFoundErr == nil || notFoundErr.Error() != "found no service matching ID notfound locally" {
		t.Errorf("Expected not found error, got %v instead", err)
	}
}

func TestGetListFromDirectoryIgnoreNestedServiceOnRootLevel(t *testing.T) {
	var services, err = GetListFromDirectory("mocks/nest")

	if err != nil {
		t.Errorf("Expected %v, got %v instead", nil, err)
	}

	var wantIDs = []string{"hi"}

	if !reflect.DeepEqual(services.GetIDs(), wantIDs) {
		t.Errorf("Want %v, got %v instead", wantIDs, services)
	}

	var wantLocations = []string{
		abs("mocks/nest"),
	}

	if !reflect.DeepEqual(services.GetLocations(), wantLocations) {
		t.Errorf("Want %v, got %v instead", wantLocations, services)
	}

	var want = ServiceInfoList{
		ServiceInfo{
			"",
			"hi",
			abs("mocks/nest"),
		},
	}

	if !reflect.DeepEqual(services, want) {
		t.Errorf("Want %v, got %v instead", want, services)
	}
}

func TestGetListFromDirectoryOnProjectWithServicesInsideSubdirectories(t *testing.T) {
	var services, err = GetListFromDirectory("mocks/project-with-services-inside-subdirs")

	if err != nil {
		t.Errorf("Expected %v, got %v instead", nil, err)
	}

	var wantIDs = []string{"level1"}

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
		abs("mocks/project-with-services-inside-subdirs/service-level-1"),
	}

	if !reflect.DeepEqual(services.GetLocations(), wantLocations) {
		t.Errorf("Want %v, got %v instead", wantLocations, services)
	}

	var want = ServiceInfoList{
		ServiceInfo{
			"exampleProject",
			"level1",
			abs("mocks/project-with-services-inside-subdirs/service-level-1"),
		},
	}

	if !reflect.DeepEqual(services, want) {
		t.Errorf("Want %v, got %v instead", want, services)
	}

	// noWhat variable just to avoid a replace all messing things up
	// mocks/project-with-services-inside-subdirs/sub-dir-2/same-dir-skipped-unreachable-service
	// also tries to fail if that happens
	var noWhat = "service"
	if _, err = os.Stat("mocks/project-with-services-inside-subdirs/skip/.no" + noWhat); err != nil {
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

	var wantErr = fmt.Sprintf(`found services with duplicated ID "email" on %v and %v`,
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
		CPU:       json.Number("0.5"),
		Memory:    json.Number("50.5"),
	}

	servertest.Mux.HandleFunc("/projects/images/services/search7606",
		tdata.ServerJSONFileHandler("mocks/service_response.json"))

	var got, err = client.Get(context.Background(), "images", "search7606")

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
	var _, err = client.Get(context.Background(), "images", "")

	if err != ErrEmptyServiceID {
		t.Errorf("Wanted error to be %v, got %v instead", ErrEmptyServiceID, err)
	}
}

func TestGetEmptyProjectID(t *testing.T) {
	var _, err = client.Get(context.Background(), "", "foo")

	if err != ErrEmptyProjectID {
		t.Errorf("Wanted error to be %v, got %v instead", ErrEmptyProjectID, err)
	}
}

func TestGetEmptyProjectAndServiceID(t *testing.T) {
	var _, err = client.Get(context.Background(), "", "")

	if err != ErrEmptyProjectAndServiceID {
		t.Errorf("Wanted error to be %v, got %v instead", ErrEmptyProjectAndServiceID, err)
	}
}

func TestCatalogItem(t *testing.T) {
	servertest.Setup()

	var want = Catalog{
		CatalogItem{
			Category:    "WeDeploy™",
			Description: "Simple authentication with email/password or third-party providers like GitHub and Google.",
			Image:       "wedeploy/auth",
			Name:        "WeDeploy™ Auth",
			State:       "active",
			Versions:    []string{"2.0.0"}},
		CatalogItem{
			Category:    "WeDeploy™",
			Description: "Scalable JSON database with search and realtime that makes building realtime apps dramatically easier.",
			Image:       "wedeploy/data",
			Name:        "WeDeploy™ Data",
			State:       "active",
			Versions:    []string{"2.0.0"},
		},
	}

	servertest.Mux.HandleFunc("/catalog/services",
		tdata.ServerJSONFileHandler("mocks/catalog_services.json"))

	var got, err = client.Catalog(context.Background())

	if err != nil {
		t.Errorf("Expected no error, got %v instead", err)
	}

	if !reflect.DeepEqual(want, got) {
		t.Error("Catalog doesn't have expected structure")
	}

	servertest.Teardown()
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

	var got, err = client.List(context.Background(), "images")

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

func TestGetEnvironmentVariable(t *testing.T) {
	servertest.Setup()

	var want = EnvironmentVariable{
		Name:  "MAX_APP_THREADS",
		Value: "10",
	}

	servertest.Mux.HandleFunc("/projects/henvic/services/pix/environment-variables/MAX_APP_THREADS",
		tdata.ServerJSONFileHandler("mocks/env_response.json"))

	var got, err = client.GetEnvironmentVariable(context.Background(), "henvic", "pix", "MAX_APP_THREADS")

	if err != nil {
		t.Errorf("Expected no error, got %v instead", err)
	}

	if !reflect.DeepEqual(want, got) {
		t.Errorf("Response does not match with wanted structure.")
		t.Errorf(pretty.Compare(want, got))
	}

	servertest.Teardown()
}

func TestGetEnvironmentVariables(t *testing.T) {
	servertest.Setup()

	var want = []EnvironmentVariable{
		EnvironmentVariable{
			Name:  "MAX_APP_THREADS",
			Value: "10",
		},
		EnvironmentVariable{
			Name:  "WEBSITE_NAMESPACE",
			Value: "web",
		},
	}

	servertest.Mux.HandleFunc("/projects/henvic/services/pix/environment-variables",
		tdata.ServerJSONFileHandler("mocks/envs_response.json"))

	var got, err = client.GetEnvironmentVariables(context.Background(), "henvic", "pix")

	if err != nil {
		t.Errorf("Expected no error, got %v instead", err)
	}

	if !reflect.DeepEqual(want, got) {
		t.Errorf("Response does not match with wanted structure.")
		t.Errorf(pretty.Compare(want, got))
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

func TestReadEmail(t *testing.T) {
	var s, err = Read("mocks/app-for/email")

	if s.ID != "email" {
		t.Errorf(`Expected email to be "email", got %v instead`, s.ID)
	}

	if s.Scale != 10 {
		t.Errorf("Expected scale to be 10, got %v instead", s.Scale)
	}

	if err != nil {
		t.Errorf("Expected err to be nil, got %v instead", err)
	}
}

func TestReadCorrupted(t *testing.T) {
	var _, err = Read("mocks/app-with-invalid-service/corrupted")

	var want = "error parsing wedeploy.json on mocks/app-with-invalid-service/corrupted:" +
		" invalid character 'I' looking for beginning of value"

	if err == nil || err.Error() != want {
		t.Errorf("Wanted err to be %v, got %v instead", want, err)
	}
}

func TestSetEnvironmentVariables(t *testing.T) {
	servertest.Setup()

	var envs = []EnvironmentVariable{
		EnvironmentVariable{
			Name:  "xyz",
			Value: "abc",
		},
	}

	var want = requestBodySetEnv{
		Env: envMap{
			"xyz": "abc",
		},
	}

	servertest.Mux.HandleFunc("/projects/foo/services/bar/environment-variables",
		func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPut {
				t.Errorf("Expected method %v, got %v instead", http.MethodPut, r.Method)
			}

			var body, err = ioutil.ReadAll(r.Body)

			if err != nil {
				t.Errorf("Error parsing response")
			}

			var got requestBodySetEnv

			err = json.Unmarshal(body, &got)

			if err != nil {
				t.Error(err)
			}

			if !reflect.DeepEqual(want, got) {
				t.Errorf("Expected received values for environment variables doesn't match.")
			}
		})

	if err := client.SetEnvironmentVariables(context.Background(), "foo", "bar", envs); err != nil {
		t.Errorf("Expected no error when setting environment variables, got %v instead", err)
	}

	servertest.Teardown()
}

func TestSetEnvironmentVariable(t *testing.T) {
	servertest.Setup()

	servertest.Mux.HandleFunc("/projects/foo/services/bar/environment-variables/xyz",
		func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPut {
				t.Errorf("Expected method %v, got %v instead", http.MethodPut, r.Method)
			}

			var body, err = ioutil.ReadAll(r.Body)

			if err != nil {
				t.Errorf("Error parsing response")
			}

			var want = map[string]string{
				"value": "abc",
			}

			var got map[string]string

			err = json.Unmarshal(body, &got)

			if err != nil {
				t.Error(err)
			}

			if !reflect.DeepEqual(want, got) {
				t.Errorf("Expected received values for environment variable doesn't match.")
			}
		})

	if err := client.SetEnvironmentVariable(context.Background(), "foo", "bar", "xyz", "abc"); err != nil {
		t.Errorf("Expected no error when adding domains, got %v instead", err)
	}

	servertest.Teardown()
}

func TestUnsetEnvironmentVariable(t *testing.T) {
	servertest.Setup()

	servertest.Mux.HandleFunc("/projects/foo/services/bar/environment-variables/xyz",
		func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodDelete {
				t.Errorf("Expected method %v, got %v instead", http.MethodDelete, r.Method)
			}
		})

	if err := client.UnsetEnvironmentVariable(context.Background(), "foo", "bar", "xyz"); err != nil {
		t.Errorf("Expected no error when unsetting environment variable, got %v instead", err)
	}

	servertest.Teardown()
}

func TestScale(t *testing.T) {
	servertest.Setup()

	servertest.Mux.HandleFunc("/projects/foo/services/bar/scale",
		func(w http.ResponseWriter, r *http.Request) {
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
				`{"value": 10}`,
				data)

			w.WriteHeader(http.StatusNoContent)
		})

	var scale = Scale{
		Current: 10,
	}

	if err := client.Scale(context.Background(), "foo", "bar", scale); err != nil {
		t.Errorf("Unexpected error on setting service scale: %v", err)
	}

	servertest.Teardown()
}

func TestRestart(t *testing.T) {
	servertest.Setup()

	servertest.Mux.HandleFunc("/projects/foo/services/bar/restart",
		func(w http.ResponseWriter, r *http.Request) {
			_, _ = fmt.Fprintf(w, `"on"`)
		})

	if err := client.Restart(context.Background(), "foo", "bar"); err != nil {
		t.Errorf("Unexpected error on service restart: %v", err)
	}

	servertest.Teardown()
}

func TestDelete(t *testing.T) {
	servertest.Setup()

	servertest.Mux.HandleFunc("/projects/foo/services/bar", func(w http.ResponseWriter, r *http.Request) {
		var wantMethod = "DELETE"
		if r.Method != wantMethod {
			t.Errorf("Wanted method %v, got %v instead", wantMethod, r.Method)
		}
	})

	var err = client.Delete(context.Background(), "foo", "bar")

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

	if err := client.Validate(context.Background(), "foo", "bar"); err != nil {
		t.Errorf("Wanted null error, got %v instead", err)
	}

	servertest.Teardown()
}

func TestValidateAlreadyExists(t *testing.T) {
	servertest.Setup()

	servertest.Mux.HandleFunc("/validators/services/id",
		func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json; charset=UTF-8")
			w.WriteHeader(404)
			_, _ = fmt.Fprintf(w, tdata.FromFile("mocks/service_already_exists_response.json"))
		})

	if err := client.Validate(context.Background(), "foo", "bar"); err != ErrServiceAlreadyExists {
		t.Errorf("Wanted %v error, got %v instead", ErrServiceAlreadyExists, err)
	}

	servertest.Teardown()
}

func TestValidateInvalidID(t *testing.T) {
	servertest.Setup()

	servertest.Mux.HandleFunc("/validators/services/id",
		func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json; charset=UTF-8")
			w.WriteHeader(404)
			_, _ = fmt.Fprintf(w, tdata.FromFile("mocks/service_invalid_id_response.json"))
		})

	if err := client.Validate(context.Background(), "foo", "bar"); err != ErrInvalidServiceID {
		t.Errorf("Wanted %v error, got %v instead", ErrInvalidServiceID, err)
	}

	servertest.Teardown()
}

func TestValidateError(t *testing.T) {
	servertest.Setup()

	servertest.Mux.HandleFunc("/validators/services/id",
		func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json; charset=UTF-8")
			w.WriteHeader(400)
			_, _ = fmt.Fprintf(w, tdata.FromFile("../apihelper/mocks/unknown_error_api_response.json"))
		})

	var err = client.Validate(context.Background(), "foo", "bar")

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
			w.Header().Set("Content-Type", "application/json; charset=UTF-8")
			w.WriteHeader(404)
		})

	var err = client.Validate(context.Background(), "foo", "bar")

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
