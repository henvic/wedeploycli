package containers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
	"os"
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

var bufOutStream bytes.Buffer

func TestMain(m *testing.M) {
	var defaultOutStream = outStream
	outStream = &bufOutStream

	ec := m.Run()

	outStream = defaultOutStream
	os.Exit(ec)
}

func TestGetListFromDirectory(t *testing.T) {
	var containers, err = GetListFromDirectory("mocks/app")

	if err != nil {
		t.Errorf("Expected %v, got %v instead", nil, err)
	}

	var wantContainers = []string{"email", "landing"}

	if !reflect.DeepEqual(containers, wantContainers) {
		t.Errorf("Want %v, got %v instead", wantContainers, containers)
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
		ID:        "search7606",
		Name:      "Cloud Search",
		Health:    "on",
		Type:      "cloudsearch",
		Instances: 7,
	}

	servertest.Mux.HandleFunc("/projects/images/containers/search7606",
		tdata.ServerJSONFileHandler("mocks/container_response.json"))

	var got, err = Get("images", "search7606")

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
			ID:        "search7606",
			Name:      "Cloud Search",
			Health:    "on",
			Type:      "cloudsearch",
			Instances: 7,
		},
		"nodejs5143": &Container{
			ID:        "nodejs5143",
			Name:      "Node.js",
			Health:    "on",
			Type:      "nodejs",
			Instances: 5,
		},
	}

	servertest.Mux.HandleFunc("/projects/images/containers",
		tdata.ServerJSONFileHandler("mocks/containers_response.json"))

	var got, err = List("images")

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
	bufOutStream.Reset()

	servertest.Mux.HandleFunc(
		"/deploy",
		func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "PUT" {
				t.Errorf("Expected install method to be PUT")
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
				`{"id":"speaker", "name": "Speaker", "type": "nodejs"}`,
				data)
		})

	var c = &Container{
		ID:   "speaker",
		Name: "Speaker",
		Type: "nodejs",
	}

	var err = Link("sound", "", c)

	if err != nil {
		t.Errorf("Unexpected error on Install: %v", err)
	}

	servertest.Teardown()
	configmock.Teardown()
}

func TestRegistry(t *testing.T) {
	servertest.Setup()
	configmock.Setup()
	bufOutStream.Reset()

	servertest.Mux.HandleFunc(
		"/registry.json",
		tdata.ServerJSONFileHandler("mocks/registry.json"))

	var registry, err = GetRegistry()

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
	bufOutStream.Reset()

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

	Restart("foo", "bar")

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

	var err = Unlink("foo", "bar")

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

	if err := Validate("foo", "bar"); err != nil {
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

	if err := Validate("foo", "bar"); err != ErrContainerAlreadyExists {
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

	if err := Validate("foo", "bar"); err != ErrInvalidContainerID {
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

	var err = Validate("foo", "bar")

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

	var err = Validate("foo", "bar")

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
