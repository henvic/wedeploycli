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

	"github.com/wedeploy/api-go/jsonlib"
	"github.com/wedeploy/cli/apihelper"
	"github.com/wedeploy/cli/globalconfigmock"
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

func TestList(t *testing.T) {
	servertest.Setup()
	globalconfigmock.Setup()
	bufOutStream.Reset()

	var want = tdata.FromFile("mocks/want_containers")

	servertest.Mux.HandleFunc("/projects/images/containers",
		tdata.ServerJSONFileHandler("mocks/containers_response.json"))

	List("images")

	if bufOutStream.String() != want {
		t.Errorf("Wanted %v, got %v instead", want, bufOutStream.String())
	}

	servertest.Teardown()
	globalconfigmock.Teardown()
}

func TestLink(t *testing.T) {
	servertest.Setup()
	globalconfigmock.Setup()
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
	globalconfigmock.Teardown()
}

func TestGetStatus(t *testing.T) {
	servertest.Setup()
	globalconfigmock.Setup()

	servertest.Mux.HandleFunc("/projects/foo/containers/bar/state",
		tdata.ServerJSONHandler(`"on"`))

	var want = "on"
	var got = GetStatus("foo", "bar")

	if got != want {
		t.Errorf("Wanted %v, got %v instead", want, got)
	}

	servertest.Teardown()
	globalconfigmock.Teardown()
}

func TestRegistry(t *testing.T) {
	servertest.Setup()
	globalconfigmock.Setup()
	bufOutStream.Reset()

	servertest.Mux.HandleFunc(
		"/registry",
		tdata.ServerJSONFileHandler("mocks/registry.json"))

	var registry = GetRegistry()

	if len(registry) != 7 {
		t.Errorf("Expected registry to have 7 images")
	}

	servertest.Teardown()
	globalconfigmock.Teardown()
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
	globalconfigmock.Setup()
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
	globalconfigmock.Teardown()
}

func TestUnlink(t *testing.T) {
	servertest.Setup()
	globalconfigmock.Setup()

	servertest.Mux.HandleFunc("/deploy", func(w http.ResponseWriter, r *http.Request) {
		var wantMethod = "DELETE"
		if r.Method != wantMethod {
			t.Errorf("Wanted method %v, got %v instead", wantMethod, r.Method)
		}

		var p, err = url.ParseQuery(r.URL.RawQuery)

		if err != nil {
			panic(err)
		}

		if p.Get("projectId") != "foo" || p.Get("containerId") != "bar" {
			t.Errorf("Wrong query parameters, got %v", r.URL.RawQuery)
		}
	})

	var err = Unlink("foo", "bar")

	if err != nil {
		t.Errorf("Expected no error, got %v instead", err)
	}

	servertest.Teardown()
	globalconfigmock.Teardown()
}

func TestValidate(t *testing.T) {
	servertest.Setup()
	globalconfigmock.Setup()

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
	globalconfigmock.Teardown()
}

func TestValidateAlreadyExists(t *testing.T) {
	servertest.Setup()
	globalconfigmock.Setup()

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
	globalconfigmock.Teardown()
}

func TestValidateInvalidID(t *testing.T) {
	servertest.Setup()
	globalconfigmock.Setup()

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
	globalconfigmock.Teardown()
}

func TestValidateError(t *testing.T) {
	servertest.Setup()
	globalconfigmock.Setup()

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
	globalconfigmock.Teardown()
}

func TestValidateInvalidError(t *testing.T) {
	servertest.Setup()
	globalconfigmock.Setup()

	servertest.Mux.HandleFunc("/validators/containers/id",
		func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-type", "application/json; charset=UTF-8")
			w.WriteHeader(404)
		})

	var err = Validate("foo", "bar")

	if err == nil || err.Error() != "unexpected end of JSON input" {
		t.Errorf("Expected error didn't happen")
	}

	servertest.Teardown()
	globalconfigmock.Teardown()
}
