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
	"github.com/wedeploy/cli/projects"
	"github.com/wedeploy/cli/servertest"
	"github.com/wedeploy/cli/tdata"
)

func TestNew(t *testing.T) {
	var project, err = projects.Read("mocks/myproject")

	if err != nil {
		panic(err)
	}

	_, err = New(project, "mocks/myproject/mycontainer")

	if err != nil {
		t.Errorf("Expected New error to be null, got %v instead", err)
	}
}

func TestNewErrorProjectNotFound(t *testing.T) {
	var m Machine
	var err = m.Setup("mocks/foo", []string{})

	if err != projects.ErrProjectNotFound {
		t.Errorf("Expected project to be not found, got %v instead", err)
	}
}

func TestNewErrorContainerNotFound(t *testing.T) {
	var project, err = projects.Read("mocks/myproject")

	if err != nil {
		panic(err)
	}

	_, err = New(project, "foo")

	if err != containers.ErrContainerNotFound {
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
	var defaultOutStream = outStream
	var b = &bytes.Buffer{}
	outStream = b
	servertest.Setup()
	configmock.Setup()

	servertest.Mux.HandleFunc("/projects",
		func(w http.ResponseWriter, r *http.Request) {})

	servertest.Mux.HandleFunc("/deploy",
		func(w http.ResponseWriter, r *http.Request) {})

	var m Machine
	var err = m.Setup("mocks/myproject", []string{"mycontainer"})

	if err != nil {
		t.Errorf("Unexpected error %v on linking", err)
	}

	m.Run()

	if b.String() != "New project project created.\n" {
		t.Errorf("Wanted new project message not found.")
	}

	configmock.Teardown()
	servertest.Teardown()
	outStream = defaultOutStream
}

func TestAllAuth(t *testing.T) {
	var defaultOutStream = outStream
	var b = &bytes.Buffer{}
	outStream = b
	servertest.Setup()
	configmock.Setup()

	servertest.Mux.HandleFunc("/projects",
		func(w http.ResponseWriter, r *http.Request) {})

	servertest.Mux.HandleFunc("/deploy",
		func(w http.ResponseWriter, r *http.Request) {})

	var m Machine
	var err = m.Setup("mocks/project-with-auth", []string{"mycontainer"})

	if err != nil {
		t.Errorf("Unexpected error %v on linking", err)
	}

	m.Run()

	if b.String() != "New project project created.\n" {
		t.Errorf("Wanted new project message not found.")
	}

	configmock.Teardown()
	servertest.Teardown()
	outStream = defaultOutStream
}

func TestAllOnlyNewError(t *testing.T) {
	var defaultOutStream = outStream
	var b = &bytes.Buffer{}
	outStream = b
	servertest.Setup()

	servertest.Mux.HandleFunc("/projects",
		func(w http.ResponseWriter, r *http.Request) {})

	configmock.Setup()

	var m Machine
	var err = m.Setup("mocks/myproject", []string{"nil"})

	if err != nil {
		panic(err)
	}

	m.Run()

	var list = m.Errors.List

	if len(list) != 1 {
		t.Errorf("Expected 1 element on the list.")
	}

	var nilerr = list[0]

	if nilerr.ContainerPath != "nil" {
		t.Errorf("Expected container to be 'nil'")
	}

	if nilerr.Error != containers.ErrContainerNotFound {
		t.Errorf("Expected not exists error for container 'nil'")
	}

	if b.String() != "New project project created.\n" {
		t.Errorf("Wanted new project message not found.")
	}

	configmock.Teardown()
	servertest.Teardown()
	outStream = defaultOutStream
}

func TestAllMultipleWithOnlyNewError(t *testing.T) {
	var defaultOutStream = outStream
	var b = &bytes.Buffer{}
	outStream = b
	servertest.Setup()
	configmock.Setup()

	servertest.Mux.HandleFunc("/projects",
		func(w http.ResponseWriter, r *http.Request) {})

	servertest.Mux.HandleFunc("/deploy",
		func(w http.ResponseWriter, r *http.Request) {})

	var m Machine
	var err = m.Setup(
		"mocks/myproject",
		[]string{"mycontainer", "nil", "nil2"})

	if err != nil {
		panic(err)
	}

	m.Run()

	var list = m.Errors.List

	if len(list) != 2 {
		t.Errorf("Expected error list of %v to have 2 items", err)
	}

	var find = map[string]bool{
		"nil":  true,
		"nil2": true,
	}

	for _, e := range list {
		if !find[e.ContainerPath] {
			t.Errorf("Unexpected %v on the error list %v",
				e.ContainerPath, list)
		}
	}

	configmock.Teardown()
	servertest.Teardown()
	outStream = defaultOutStream
}

func TestAllValidateOrCreateFailure(t *testing.T) {
	servertest.Setup()
	configmock.Setup()

	servertest.Mux.HandleFunc("/projects",
		func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(403)
		})

	var m Machine
	var err = m.Setup("mocks/myproject", []string{})

	if err == nil || err.(*apihelper.APIFault).Code != 403 {
		t.Errorf("Expected 403 Forbidden error, got %v instead", err)
	}

	configmock.Teardown()
	servertest.Teardown()
}

func TestAllInstallContainerError(t *testing.T) {
	var defaultOutStream = outStream
	var b = &bytes.Buffer{}
	outStream = b
	servertest.Setup()
	configmock.Setup()

	servertest.Mux.HandleFunc("/projects",
		func(w http.ResponseWriter, r *http.Request) {})

	servertest.Mux.HandleFunc("/deploy",
		func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(403)
		})

	var m Machine
	var err = m.Setup("mocks/myproject", []string{"mycontainer"})

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
	outStream = defaultOutStream
}
