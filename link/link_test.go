package link

import (
	"bytes"
	"fmt"
	"net/http"
	"os"
	"testing"

	"github.com/wedeploy/cli/apihelper"
	"github.com/wedeploy/cli/configmock"
	"github.com/wedeploy/cli/containers"
	"github.com/wedeploy/cli/servertest"
	"github.com/wedeploy/cli/tdata"
)

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

func TestAll(t *testing.T) {
	servertest.Setup()
	configmock.Setup()

	var wantSource = "mocks/myproject/mycontainer"
	var gotSource string

	servertest.Mux.HandleFunc("/deploy",
		func(w http.ResponseWriter, r *http.Request) {
			gotSource = r.URL.Query().Get("source")
		})

	var m Machine
	var err = m.Setup([]string{"mocks/myproject/mycontainer"})

	if err != nil {
		t.Errorf("Unexpected error %v on linking", err)
	}

	m.Run()

	if len(m.Errors.List) != 0 {
		t.Errorf("Wanted list of errors to contain zero errors, got %v errors instead", m.Errors)
	}

	if wantSource != gotSource {
		t.Errorf("Wanted source %v, got %v instead", wantSource, gotSource)
	}

	configmock.Teardown()
	servertest.Teardown()
}

func TestAllQuiet(t *testing.T) {
	servertest.Setup()
	configmock.Setup()

	var wantSource = "mocks/myproject/mycontainer"
	var gotSource string

	servertest.Mux.HandleFunc("/deploy",
		func(w http.ResponseWriter, r *http.Request) {
			gotSource = r.URL.Query().Get("source")
		})

	var m Machine
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

	if wantSource != gotSource {
		t.Errorf("Wanted source %v, got %v instead", wantSource, gotSource)
	}

	configmock.Teardown()
	servertest.Teardown()
}

func TestAllOnlyNewError(t *testing.T) {
	servertest.Setup()
	configmock.Setup()

	var m Machine
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

	servertest.Mux.HandleFunc("/deploy",
		func(w http.ResponseWriter, r *http.Request) {})

	var m Machine
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

	servertest.Mux.HandleFunc("/deploy",
		func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(403)
		})

	var m Machine
	var err = m.Setup([]string{"mocks/myproject/mycontainer"})

	if err != nil {
		panic(err)
	}

	m.Run()

	var el = m.Errors.List
	var af = el[0].Error.(*apihelper.APIFault)

	if m.Errors == nil || af.Code != 403 {
		t.Errorf("Expected 403 Forbidden error, got %v instead", m.Errors)
	}

	configmock.Teardown()
	servertest.Teardown()
}
