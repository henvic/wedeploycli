package link

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"testing"

	"github.com/wedeploy/api-go/jsonlib"
	"github.com/wedeploy/cli/apihelper"
	"github.com/wedeploy/cli/config"
	"github.com/wedeploy/cli/containers"
	"github.com/wedeploy/cli/hooks"
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

func TestNew(t *testing.T) {
	if _, err := New("mocks/myproject/mycontainer"); err != nil {
		t.Errorf("Expected New error to be null, got %v instead", err)
	}
}

func TestNewErrorContainerNotFound(t *testing.T) {
	if _, err := New("foo"); err != containers.ErrContainerNotFound {
		t.Errorf("Expected container to be not found, got %v instead", err)
	}
}

func TestErrors(t *testing.T) {
	var fooe = ContainerError{
		ContainerPath: "foo",
		Error:         os.ErrExist,
	}

	var bare = ContainerError{
		ContainerPath: "bar",
		Error:         os.ErrNotExist,
	}

	var e error = Errors{
		List: []ContainerError{fooe, bare},
	}

	var want = tdata.FromFile("mocks/test_errors")

	if fmt.Sprintf("%v", e) != want {
		t.Errorf("Wanted error %v, got %v instead.", want, e)
	}
}

func TestMissingProject(t *testing.T) {
	servertest.Setup()
	configmock.Setup()

	var m Machine
	var err = m.Setup([]string{"mocks/myproject/mycontainer"})

	if err != errMissingProjectID {
		t.Errorf("Expected error to be %v, got %v instead", errMissingProjectID, err)
	}

	configmock.Teardown()
	servertest.Teardown()
}

func TestAll(t *testing.T) {
	servertest.Setup()
	configmock.Setup()

	servertest.Mux.HandleFunc("/projects/foo/services",
		func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "POST" {
				t.Errorf("Expected method to be PUT, got %v instead", r.Method)
			}

			var body, err = ioutil.ReadAll(r.Body)

			if err != nil {
				t.Errorf("Expected no error, got %v isntead", err)
			}

			jsonlib.AssertJSONMarshal(t, string(body), containers.Container{
				ServiceID: "container",
				Source:    "mocks/myproject/mycontainer",
				Hooks:     &hooks.Hooks{},
			})
		})

	var m = &Machine{
		ProjectID: "foo",
	}

	var err = m.Setup([]string{"mocks/myproject/mycontainer"})

	if err != nil {
		t.Errorf("Unexpected error %v on linking", err)
	}

	m.Run()

	if len(m.Errors.List) != 0 {
		t.Errorf("Wanted list of errors to contain zero errors, got %v errors instead", m.Errors)
	}

	configmock.Teardown()
	servertest.Teardown()
}

func TestAllQuiet(t *testing.T) {
	servertest.Setup()
	configmock.Setup()

	servertest.Mux.HandleFunc("/projects/foo/services",
		func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "POST" {
				t.Errorf("Expected method to be POST, got %v instead", r.Method)
			}

			var body, err = ioutil.ReadAll(r.Body)

			if err != nil {
				t.Errorf("Expected no error, got %v isntead", err)
			}

			jsonlib.AssertJSONMarshal(t, string(body), containers.Container{
				ServiceID: "container",
				Source:    "mocks/myproject/mycontainer",
				Hooks:     &hooks.Hooks{},
			})
		})

	var m = &Machine{
		ProjectID: "foo",
	}

	var bufErrStream bytes.Buffer
	m.ErrStream = &bufErrStream

	var err = m.Setup([]string{"mocks/myproject/mycontainer"})

	if err != nil {
		t.Errorf("Unexpected error %v on linking", err)
	}

	m.Run()

	var wantContainerLinkedMessage = "Container container linked.\n"
	if bufErrStream.String() != wantContainerLinkedMessage {
		t.Errorf("Wanted container linked message not, got %v instead.", bufErrStream.String())
	}

	if len(m.Errors.List) != 0 {
		t.Errorf("Wanted list of errors to contain zero errors, got %v errors instead", m.Errors)
	}

	configmock.Teardown()
	servertest.Teardown()
}

func TestAllOnlyNewError(t *testing.T) {
	servertest.Setup()
	configmock.Setup()

	var m = &Machine{
		ProjectID: "foo",
	}

	var err = m.Setup([]string{"mocks/myproject/nil"})

	if err != nil {
		panic(err)
	}

	m.Run()

	var list = m.Errors.List

	if len(list) != 1 {
		t.Errorf("Expected 1 element on the list.")
	}

	var nilerr = list[0]

	if nilerr.ContainerPath != "mocks/myproject/nil" {
		t.Errorf("Expected container to be 'nil'")
	}

	if nilerr.Error != containers.ErrContainerNotFound {
		t.Errorf("Expected not exists error for container 'nil'")
	}

	configmock.Teardown()
	servertest.Teardown()
}

func TestAllMultipleWithOnlyNewError(t *testing.T) {
	servertest.Setup()
	configmock.Setup()

	servertest.Mux.HandleFunc("/projects/foo/services",
		func(w http.ResponseWriter, r *http.Request) {})

	var m = &Machine{
		ProjectID: "foo",
	}

	var err = m.Setup(
		[]string{"mocks/myproject/mycontainer", "mocks/myproject/nil", "mocks/myproject/nil2"})

	if err != nil {
		panic(err)
	}

	m.Run()

	var list = m.Errors.List

	if len(list) != 2 {
		t.Errorf("Expected error list of %v to have 2 items", err)
	}

	var find = map[string]bool{
		"mocks/myproject/nil":  true,
		"mocks/myproject/nil2": true,
	}

	for _, e := range list {
		if !find[e.ContainerPath] {
			t.Errorf("Unexpected %v on the error list %v",
				e.ContainerPath, list)
		}
	}

	configmock.Teardown()
	servertest.Teardown()
}

func TestAllInstallContainerError(t *testing.T) {
	servertest.Setup()
	configmock.Setup()

	servertest.Mux.HandleFunc("/projects/foo/services",
		func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(403)
		})

	var m = &Machine{
		ProjectID: "foo",
	}

	var err = m.Setup([]string{"mocks/myproject/mycontainer"})

	if err != nil {
		panic(err)
	}

	m.Run()

	var el = m.Errors.List
	var af = el[0].Error.(*apihelper.APIFault)

	if m.Errors == nil || af.Status != 403 {
		t.Errorf("Expected 403 Forbidden error, got %v instead", m.Errors)
	}

	configmock.Teardown()
	servertest.Teardown()
}
