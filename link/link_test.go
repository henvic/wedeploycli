package link

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"os"
	"testing"

	"github.com/wedeploy/cli/apihelper"
	"github.com/wedeploy/cli/config"
	"github.com/wedeploy/cli/services"
	"github.com/wedeploy/cli/defaults"
	"github.com/wedeploy/cli/projects"
	"github.com/wedeploy/cli/servertest"
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
	if _, err := New("mocks/myproject/myservice"); err != nil {
		t.Errorf("Expected New error to be null, got %v instead", err)
	}
}

func TestNewErrorServiceNotFound(t *testing.T) {
	if _, err := New("foo"); err != services.ErrServiceNotFound {
		t.Errorf("Expected service to be not found, got %v instead", err)
	}
}

func TestErrors(t *testing.T) {
	var fooe = ServiceError{
		ServicePath: "foo",
		Error:         os.ErrExist,
	}

	var bare = ServiceError{
		ServicePath: "bar",
		Error:         os.ErrNotExist,
	}

	var e error = Errors{
		List: []ServiceError{fooe, bare},
	}

	var want = `Local deployment errors:
foo: file already exists
bar: file does not exist`

	if fmt.Sprintf("%v", e) != want {
		t.Errorf("Wanted error %v, got %v instead.", want, e)
	}
}

func TestMissingProject(t *testing.T) {
	servertest.Setup()

	var m Machine
	var err = m.Setup([]string{"mocks/myproject/myservice"})

	if err != errMissingProjectID {
		t.Errorf("Expected error to be %v, got %v instead", errMissingProjectID, err)
	}

	servertest.Teardown()
}

func TestAll(t *testing.T) {
	servertest.Setup()

	servertest.Mux.HandleFunc("/projects/foo/services",
		func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-type", "application/json; charset=UTF-8")
			fmt.Fprintf(w, "[]")
		})

	var m = &Machine{
		Project: projects.Project{
			ProjectID: "foo",
		},
	}

	var err = m.Setup([]string{"mocks/myproject/myservice"})

	if err != nil {
		t.Errorf("Unexpected error %v on linking", err)
	}

	var ctx, cancel = context.WithCancel(context.Background())
	m.Run(cancel)
	<-ctx.Done()

	if len(m.Errors.List) != 0 {
		t.Errorf("Wanted list of errors to contain zero errors, got %v errors instead", m.Errors)
	}

	servertest.Teardown()
}

func TestAllQuiet(t *testing.T) {
	servertest.Setup()

	servertest.Mux.HandleFunc("/projects/foo/services",
		func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-type", "application/json; charset=UTF-8")
			fmt.Fprintf(w, "[]")
		})

	var m = &Machine{
		Project: projects.Project{
			ProjectID: "foo",
		},
	}

	var bufErrStream bytes.Buffer
	m.ErrStream = &bufErrStream

	var err = m.Setup([]string{"mocks/myproject/myservice"})

	if err != nil {
		t.Errorf("Unexpected error %v on linking", err)
	}

	var ctx, cancel = context.WithCancel(context.Background())
	m.Run(cancel)
	<-ctx.Done()

	var wantServiceLinkedMessage = "Service service deployed locally.\n"
	if bufErrStream.String() != wantServiceLinkedMessage {
		t.Errorf("Wanted service deployed locally message, got %v instead.", bufErrStream.String())
	}

	if len(m.Errors.List) != 0 {
		t.Errorf("Wanted list of errors to contain zero errors, got %v errors instead", m.Errors)
	}

	servertest.Teardown()
}

func TestAllMultipleWithOnlyNewError(t *testing.T) {
	servertest.Setup()

	servertest.Mux.HandleFunc("/projects/foo/services",
		func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprintf(w, `[]`)
		})

	var m = &Machine{
		Project: projects.Project{
			ProjectID: "foo",
		},
	}

	var err = m.Setup(
		[]string{"mocks/myproject/myservice", "mocks/myproject/nil", "mocks/myproject/nil2"})

	if err != nil {
		panic(err)
	}

	var ctx, cancel = context.WithCancel(context.Background())
	m.Run(cancel)
	<-ctx.Done()

	var list = m.Errors.List

	if len(list) != 2 {
		t.Errorf("Expected error list of %v to have 2 items", err)
	}

	var find = map[string]bool{
		"mocks/myproject/nil":  true,
		"mocks/myproject/nil2": true,
	}

	for _, e := range list {
		if !find[e.ServicePath] {
			t.Errorf("Unexpected %v on the error list %v",
				e.ServicePath, list)
		}
	}

	servertest.Teardown()
}

func TestAllInstallServiceError(t *testing.T) {
	servertest.Setup()

	servertest.Mux.HandleFunc("/projects/foo/services",
		func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(403)
		})

	var m = &Machine{
		Project: projects.Project{
			ProjectID: "foo",
		},
	}

	var err = m.Setup([]string{"mocks/myproject/myservice"})
	var af = err.(*apihelper.APIFault)

	if err == nil || af.Status != 403 {
		t.Errorf("Expected 403 Forbidden error, got %v instead", m.Errors)
	}

	servertest.Teardown()
}
