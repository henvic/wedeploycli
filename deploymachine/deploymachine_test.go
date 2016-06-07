package deploymachine

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/wedeploy/cli/apihelper"
	"github.com/wedeploy/cli/config"
	"github.com/wedeploy/cli/containers"
	"github.com/wedeploy/cli/deploy"
	"github.com/wedeploy/cli/globalconfigmock"
	"github.com/wedeploy/cli/servertest"
	"github.com/wedeploy/cli/tdata"
)

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
	var workingDir, _ = os.Getwd()

	if err := os.Chdir(filepath.Join(workingDir, "mocks/myproject")); err != nil {
		t.Error(err)
	}

	config.Setup()
	globalconfigmock.Setup()

	servertest.Mux.HandleFunc("/projects",
		func(w http.ResponseWriter, r *http.Request) {})

	servertest.Mux.HandleFunc("/containers",
		func(w http.ResponseWriter, r *http.Request) {})

	servertest.Mux.HandleFunc("/push/project/container",
		func(w http.ResponseWriter, r *http.Request) {
			var _, _, err = r.FormFile("pod")

			if err != nil {
				t.Error(err)
			}
		})

	var success, err = All("project", []string{"mycontainer"}, &deploy.Flags{})

	if err != nil {
		t.Errorf("Unexpected error %v on deploy", err)
	}

	var wantFeedback = tdata.FromFile("../deploy_feedback")

	if !strings.Contains(wantFeedback, strings.Join(success, "\n")) {
		t.Errorf("Wanted feedback to contain %v, got %v instead", wantFeedback, success)
	}

	globalconfigmock.Teardown()
	config.Teardown()
	os.Chdir(workingDir)
	servertest.Teardown()
}

func TestAllWithHooks(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Not testing with hooks on Windows")
	}

	servertest.Setup()
	var workingDir, _ = os.Getwd()

	if err := os.Chdir(filepath.Join(workingDir, "mocks/myproject")); err != nil {
		t.Error(err)
	}

	config.Setup()
	globalconfigmock.Setup()

	servertest.Mux.HandleFunc("/projects",
		func(w http.ResponseWriter, r *http.Request) {})

	servertest.Mux.HandleFunc("/containers",
		func(w http.ResponseWriter, r *http.Request) {})

	servertest.Mux.HandleFunc("/push/project/container",
		func(w http.ResponseWriter, r *http.Request) {
			var _, _, err = r.FormFile("pod")

			if err != nil {
				t.Error(err)
			}
		})

	var _, err = All("project", []string{"mycontainer"}, &deploy.Flags{
		Hooks: true,
	})

	if err != nil {
		t.Errorf("Unexpected error %v on deploy", err)
	}

	globalconfigmock.Teardown()
	config.Teardown()
	os.Chdir(workingDir)
	servertest.Teardown()
}

func TestAllWithBeforeHookFailure(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Not testing with hooks on Windows")
	}

	servertest.Setup()
	var workingDir, _ = os.Getwd()

	if err := os.Chdir(filepath.Join(workingDir, "mocks/myproject")); err != nil {
		t.Error(err)
	}

	config.Setup()
	globalconfigmock.Setup()

	servertest.Mux.HandleFunc("/projects",
		func(w http.ResponseWriter, r *http.Request) {})

	servertest.Mux.HandleFunc("/containers",
		func(w http.ResponseWriter, r *http.Request) {})

	servertest.Mux.HandleFunc("/push/project/container-before-hook-failure",
		func(w http.ResponseWriter, r *http.Request) {
			var _, _, err = r.FormFile("pod")

			if err != nil {
				t.Error(err)
			}
		})

	var success, err = All("project", []string{"container_before_hook_failure"}, &deploy.Flags{
		Hooks: true,
	})

	if err == nil || err.Error() != `List of errors (format is container path: error)
container_before_hook_failure: exit status 1` {
		t.Errorf("Expected error didn't happen, got %v instead", err)
	}

	var wantFeedback = tdata.FromFile("../deploy_before_hook_failure_feedback")

	if !strings.Contains(wantFeedback, strings.Join(success, "\n")) {
		t.Errorf("Wanted feedback to contain %v, got %v instead", wantFeedback, success)
	}

	globalconfigmock.Teardown()
	config.Teardown()
	os.Chdir(workingDir)
	servertest.Teardown()
}

func TestAllWithAfterHookFailure(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Not testing with hooks on Windows")
	}

	servertest.Setup()
	var workingDir, _ = os.Getwd()

	if err := os.Chdir(filepath.Join(workingDir, "mocks/myproject")); err != nil {
		t.Error(err)
	}

	config.Setup()
	globalconfigmock.Setup()

	servertest.Mux.HandleFunc("/projects",
		func(w http.ResponseWriter, r *http.Request) {})

	servertest.Mux.HandleFunc("/containers",
		func(w http.ResponseWriter, r *http.Request) {})

	servertest.Mux.HandleFunc("/push/project/container_after_hook_failure",
		func(w http.ResponseWriter, r *http.Request) {
			var _, _, err = r.FormFile("pod")

			if err != nil {
				t.Error(err)
			}
		})

	var success, err = All("project",
		[]string{"container_after_hook_failure"},
		&deploy.Flags{
			Hooks: true,
		})

	if err == nil || err.Error() != `List of errors (format is container path: error)
container_after_hook_failure: exit status 1` {
		t.Errorf("Expected error didn't happen, got %v instead", err)
	}

	var wantFeedback = tdata.FromFile("../deploy_after_hook_failure_feedback")

	if !strings.Contains(wantFeedback, strings.Join(success, "\n")) {
		t.Errorf("Wanted feedback to contain %v, got %v instead",
			wantFeedback, success)
	}

	globalconfigmock.Teardown()
	config.Teardown()
	os.Chdir(workingDir)
	servertest.Teardown()
}

func TestAllOnlyNewError(t *testing.T) {
	servertest.Setup()
	var workingDir, _ = os.Getwd()

	servertest.Mux.HandleFunc("/projects",
		func(w http.ResponseWriter, r *http.Request) {})

	var err = os.Chdir(filepath.Join(workingDir, "mocks/myproject"))

	if err != nil {
		t.Error(err)
	}

	config.Setup()
	globalconfigmock.Setup()

	_, err = All("project", []string{"nil"}, &deploy.Flags{})

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
		t.Errorf("Error is not of expected type.")
	}

	globalconfigmock.Teardown()
	config.Teardown()
	os.Chdir(workingDir)
	servertest.Teardown()
}

func TestAllMultipleWithOnlyNewError(t *testing.T) {
	servertest.Setup()
	var workingDir, _ = os.Getwd()

	var err = os.Chdir(filepath.Join(workingDir, "mocks/myproject"))

	if err != nil {
		t.Error(err)
	}

	config.Setup()
	globalconfigmock.Setup()

	servertest.Mux.HandleFunc("/projects",
		func(w http.ResponseWriter, r *http.Request) {})

	servertest.Mux.HandleFunc("/containers",
		func(w http.ResponseWriter, r *http.Request) {})

	servertest.Mux.HandleFunc("/push/project/container",
		func(w http.ResponseWriter, r *http.Request) {
			var _, _, err = r.FormFile("pod")

			if err != nil {
				t.Error(err)
			}
		})

	_, err = All("project", []string{"mycontainer", "nil", "nil2"}, &deploy.Flags{})

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
		t.Errorf("Error is not of expected type.")
	}

	globalconfigmock.Teardown()
	config.Teardown()
	servertest.Teardown()
	os.Chdir(workingDir)
}

func TestAllValidateOrCreateFailure(t *testing.T) {
	servertest.Setup()
	var workingDir, _ = os.Getwd()

	if err := os.Chdir(filepath.Join(workingDir, "mocks/myproject")); err != nil {
		t.Error(err)
	}

	config.Setup()
	globalconfigmock.Setup()

	servertest.Mux.HandleFunc("/projects",
		func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(403)
		})

	var _, err = All("project", []string{"mycontainer"}, &deploy.Flags{})

	if err == nil || err.(*apihelper.APIFault).Code != 403 {
		t.Errorf("Expected 403 Forbidden error, got %v instead", err)
	}

	globalconfigmock.Teardown()
	config.Teardown()
	os.Chdir(workingDir)
	servertest.Teardown()
}

func TestAllInstallContainerError(t *testing.T) {
	servertest.Setup()
	var workingDir, _ = os.Getwd()

	if err := os.Chdir(filepath.Join(workingDir, "mocks/myproject")); err != nil {
		t.Error(err)
	}

	config.Setup()
	globalconfigmock.Setup()

	servertest.Mux.HandleFunc("/projects",
		func(w http.ResponseWriter, r *http.Request) {})

	servertest.Mux.HandleFunc("/containers",
		func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(403)
		})

	var _, err = All("project", []string{"mycontainer"}, &deploy.Flags{})
	var el = err.(*Errors).List
	var af = el[0].Error.(*apihelper.APIFault)

	if err == nil || af.Code != 403 {
		t.Errorf("Expected 403 Forbidden error, got %v instead", err)
	}

	globalconfigmock.Teardown()
	config.Teardown()
	os.Chdir(workingDir)
	servertest.Teardown()
}
