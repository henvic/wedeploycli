package link

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/wedeploy/cli/apihelper"
	"github.com/wedeploy/cli/containers"
	"github.com/wedeploy/cli/globalconfigmock"
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
	var _, err = All("mocks/foo", []string{})

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
	servertest.Setup()
	globalconfigmock.Setup()

	servertest.Mux.HandleFunc("/projects",
		func(w http.ResponseWriter, r *http.Request) {})

	servertest.Mux.HandleFunc("/deploy/project",
		func(w http.ResponseWriter, r *http.Request) {})

	var success, err = All("mocks/myproject", []string{"mycontainer"})

	if err != nil {
		t.Errorf("Unexpected error %v on deploy", err)
	}

	var wantFeedback = tdata.FromFile("mocks/link_feedback")

	if !strings.Contains(wantFeedback, strings.Join(success, "\n")) {
		t.Errorf("Wanted feedback to contain %v, got %v instead", wantFeedback, success)
	}

	globalconfigmock.Teardown()
	servertest.Teardown()
}

func TestAllAuth(t *testing.T) {
	servertest.Setup()
	globalconfigmock.Setup()

	servertest.Mux.HandleFunc("/projects",
		func(w http.ResponseWriter, r *http.Request) {})

	servertest.Mux.HandleFunc("/deploy/project",
		func(w http.ResponseWriter, r *http.Request) {})

	servertest.Mux.HandleFunc("/projects/project/auth",
		func(w http.ResponseWriter, r *http.Request) {})

	var success, err = All("mocks/project-with-auth", []string{"mycontainer"})

	if err != nil {
		t.Errorf("Unexpected error %v on deploy", err)
	}

	var wantFeedback = tdata.FromFile("mocks/link_feedback")

	if !strings.Contains(wantFeedback, strings.Join(success, "\n")) {
		t.Errorf("Wanted feedback to contain %v, got %v instead", wantFeedback, success)
	}

	globalconfigmock.Teardown()
	servertest.Teardown()
}

func TestAllOnlyNewError(t *testing.T) {
	servertest.Setup()

	servertest.Mux.HandleFunc("/projects",
		func(w http.ResponseWriter, r *http.Request) {})

	globalconfigmock.Setup()

	var _, err = All("mocks/myproject", []string{"nil"})

	switch err.(type) {
	case *Errors:
		var list = err.(*Errors).List

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
	default:
		t.Errorf("Error is not of expected type: %v", err)
	}

	globalconfigmock.Teardown()
	servertest.Teardown()
}

func TestAllMultipleWithOnlyNewError(t *testing.T) {
	servertest.Setup()
	globalconfigmock.Setup()

	servertest.Mux.HandleFunc("/projects",
		func(w http.ResponseWriter, r *http.Request) {})

	servertest.Mux.HandleFunc("/deploy/project",
		func(w http.ResponseWriter, r *http.Request) {})

	var _, err = All("mocks/myproject", []string{"mycontainer", "nil", "nil2"})

	switch err.(type) {
	case *Errors:
		var list = err.(*Errors).List

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
	default:
		t.Errorf("Error is not of expected type: %v", err)
	}

	globalconfigmock.Teardown()
	servertest.Teardown()
}

func TestAllValidateOrCreateFailure(t *testing.T) {
	servertest.Setup()
	globalconfigmock.Setup()

	servertest.Mux.HandleFunc("/projects",
		func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(403)
		})

	var _, err = All("mocks/myproject", []string{"mycontainer"})

	if err == nil || err.(*apihelper.APIFault).Code != 403 {
		t.Errorf("Expected 403 Forbidden error, got %v instead", err)
	}

	globalconfigmock.Teardown()
	servertest.Teardown()
}

func TestAllInstallContainerError(t *testing.T) {
	servertest.Setup()
	globalconfigmock.Setup()

	servertest.Mux.HandleFunc("/projects",
		func(w http.ResponseWriter, r *http.Request) {})

	servertest.Mux.HandleFunc("/deploy/project",
		func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(403)
		})

	var _, err = All("mocks/myproject", []string{"mycontainer"})
	var el = err.(*Errors).List
	var af = el[0].Error.(*apihelper.APIFault)

	if err == nil || af.Code != 403 {
		t.Errorf("Expected 403 Forbidden error, got %v instead", err)
	}

	globalconfigmock.Teardown()
	servertest.Teardown()
}
