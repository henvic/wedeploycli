package deploy

import (
	"bytes"
	"crypto/sha1"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/launchpad-project/cli/apihelper"
	"github.com/launchpad-project/cli/config"
	"github.com/launchpad-project/cli/globalconfigmock"
	"github.com/launchpad-project/cli/servertest"
	"github.com/launchpad-project/cli/tdata"
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

func TestNew(t *testing.T) {
	var workingDir, _ = os.Getwd()

	if err := os.Chdir(filepath.Join(workingDir, "mocks/myproject")); err != nil {
		t.Error(err)
	}

	config.Setup()

	var _, err = New("mycontainer")

	if err != nil {
		t.Errorf("Expected New error to be null, got %v instead", err)
	}

	config.Teardown()
	os.Chdir(workingDir)
}

func TestNewErrorContainerNotFound(t *testing.T) {
	var workingDir, _ = os.Getwd()

	if err := os.Chdir(filepath.Join(workingDir, "mocks/myproject")); err != nil {
		t.Error(err)
	}

	config.Setup()

	var _, err = New("foo")

	if !os.IsNotExist(err) {
		t.Errorf("Expected container to be not found, got %v instead", err)
	}

	config.Teardown()
	os.Chdir(workingDir)
}

func TestPack(t *testing.T) {
	var workingDir, _ = os.Getwd()

	if err := os.Chdir(filepath.Join(workingDir, "mocks/myproject")); err != nil {
		t.Error(err)
	}

	config.Setup()

	var err = Pack(os.DevNull, "mycontainer")

	if err != nil {
		t.Errorf("Unexpected packing error: %v", err)
	}

	config.Teardown()
	os.Chdir(workingDir)
}

func TestDeploy(t *testing.T) {
	servertest.Setup()
	var workingDir, _ = os.Getwd()
	var tmp, _ = ioutil.TempFile(os.TempDir(), "launchpad-cli")

	if err := os.Chdir(filepath.Join(workingDir, "mocks/myproject")); err != nil {
		t.Error(err)
	}

	config.Setup()
	globalconfigmock.Setup()

	var packageSHA1 = "5b4238302c12e91f0faf44bcc912eb230e8f3094"

	servertest.Mux.HandleFunc("/push/project/container",
		func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "POST" {
				t.Errorf("Unexpected method %v", r.Method)
			}

			if !strings.Contains(r.Header.Get("Content-Type"), "multipart/form-data;") {
				t.Errorf("Expected multipart/form-data")
			}

			if r.Header.Get("WeDeploy-Package-SHA1") != packageSHA1 {
				t.Errorf("Expected SHA1 on the header doesn't match expected value")
			}

			var mf, _, err = r.FormFile("pod")

			if err != nil {
				t.Error(err)
			}

			var hash = sha1.New()

			var _, eh = io.Copy(hash, mf)

			if eh != nil {
				t.Error(eh)
			}

			var gotSHA1 = fmt.Sprintf("%x", hash.Sum(nil))

			if gotSHA1 != packageSHA1 {
				t.Errorf("Wanted SHA1 %v, got %v instead.", packageSHA1, gotSHA1)
			}
		})

	var deploy, err = New("mycontainer")

	if err != nil {
		t.Errorf("Expected New error to be null, got %v instead", err)
	}

	err = deploy.Deploy("../package")

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	os.Remove(tmp.Name())
	globalconfigmock.Teardown()
	config.Teardown()
	os.Chdir(workingDir)
	servertest.Teardown()
}

func TestDeployFileNotFound(t *testing.T) {
	var workingDir, _ = os.Getwd()
	var tmp, _ = ioutil.TempFile(os.TempDir(), "launchpad-cli")

	if err := os.Chdir(filepath.Join(workingDir, "mocks/myproject")); err != nil {
		t.Error(err)
	}

	config.Setup()
	globalconfigmock.Setup()

	var deploy, err = New("mycontainer")

	if err != nil {
		t.Errorf("Expected New error to be null, got %v instead", err)
	}

	err = deploy.Deploy(filepath.Join("not-found", fmt.Sprintf("%d", rand.Int())))

	if !os.IsNotExist(err) {
		t.Errorf("Unexpected error value: %v, expected file not found.", err)
	}

	os.Remove(tmp.Name())
	globalconfigmock.Teardown()
	config.Teardown()
	os.Chdir(workingDir)
}

func TestDeployFailure(t *testing.T) {
	var defaultOutStream = outStream
	var bufOutStream bytes.Buffer
	outStream = &bufOutStream
	servertest.Setup()
	var workingDir, _ = os.Getwd()
	var tmp, _ = ioutil.TempFile(os.TempDir(), "launchpad-cli")

	if err := os.Chdir(filepath.Join(workingDir, "mocks/myproject")); err != nil {
		t.Error(err)
	}

	config.Setup()
	globalconfigmock.Setup()

	servertest.Mux.HandleFunc("/push/project/container",
		func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(500)
		})

	var deploy, err = New("mycontainer")

	if err != nil {
		t.Errorf("Expected New error to be null, got %v instead", err)
	}

	err = deploy.Deploy("../package")

	if err.(*apihelper.APIFault).Code != 500 {
		t.Errorf("Expected request error code doesn't match.")
	}

	os.Remove(tmp.Name())
	globalconfigmock.Teardown()
	config.Teardown()
	os.Chdir(workingDir)
	servertest.Teardown()
	outStream = defaultOutStream
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

	var success, err = All([]string{"mycontainer"}, &Flags{})

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

	var _, err = All([]string{"mycontainer"}, &Flags{
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

	var success, err = All([]string{"container_before_hook_failure"}, &Flags{
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

	var success, err = All([]string{"container_after_hook_failure"}, &Flags{
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

	_, err = All([]string{"nil"}, &Flags{})

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

		if !os.IsNotExist(nilerr.Error) {
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

	_, err = All([]string{"mycontainer", "nil", "nil2"}, &Flags{})

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

func TestOnly(t *testing.T) {
	var defaultOutStream = outStream
	var bufOutStream bytes.Buffer
	outStream = &bufOutStream
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

	var err = Only("mycontainer", &Flags{})

	if err != nil {
		t.Errorf("Unexpected error %v on deploy", err)
	}

	globalconfigmock.Teardown()
	config.Teardown()
	os.Chdir(workingDir)
	servertest.Teardown()
	outStream = defaultOutStream
}

func TestOnlyNewError(t *testing.T) {
	var workingDir, _ = os.Getwd()

	if err := os.Chdir(filepath.Join(workingDir, "mocks/myproject")); err != nil {
		t.Error(err)
	}

	config.Setup()
	globalconfigmock.Setup()

	var err = Only("nil", &Flags{})

	if !os.IsNotExist(err) {
		t.Errorf("Wanted error to be file not exists, got %v instead", err)
	}

	globalconfigmock.Teardown()
	config.Teardown()
	os.Chdir(workingDir)
}

func TestOnlyProjectValidationOrCreationFailure(t *testing.T) {
	var defaultOutStream = outStream
	var bufOutStream bytes.Buffer
	outStream = &bufOutStream
	servertest.Setup()
	var workingDir, _ = os.Getwd()

	if err := os.Chdir(filepath.Join(workingDir, "mocks/myproject")); err != nil {
		t.Error(err)
	}

	config.Setup()
	globalconfigmock.Setup()

	var err = Only("mycontainer", &Flags{})

	switch err.(type) {
	case *apihelper.APIFault:
	default:
		t.Errorf("Wanted error to be of type *apihelper.APIFault, got error %v instead", err)
	}

	globalconfigmock.Teardown()
	config.Teardown()
	os.Chdir(workingDir)
	servertest.Teardown()
	outStream = defaultOutStream
}

func TestOnlyContainerValidationOrCreationFailure(t *testing.T) {
	var defaultOutStream = outStream
	var bufOutStream bytes.Buffer
	outStream = &bufOutStream
	servertest.Setup()
	var workingDir, _ = os.Getwd()

	if err := os.Chdir(filepath.Join(workingDir, "mocks/myproject")); err != nil {
		t.Error(err)
	}

	config.Setup()
	globalconfigmock.Setup()

	servertest.Mux.HandleFunc("/projects",
		func(w http.ResponseWriter, r *http.Request) {})

	var err = Only("mycontainer", &Flags{})

	switch err.(type) {
	case *apihelper.APIFault:
	default:
		t.Errorf("Wanted error to be of type *apihelper.APIFault, got error %v instead", err)
	}

	globalconfigmock.Teardown()
	config.Teardown()
	os.Chdir(workingDir)
	servertest.Teardown()
	outStream = defaultOutStream
}

func TestDeployOnly(t *testing.T) {
	servertest.Setup()
	var workingDir, _ = os.Getwd()

	if err := os.Chdir(filepath.Join(workingDir, "mocks/myproject")); err != nil {
		t.Error(err)
	}

	config.Setup()
	globalconfigmock.Setup()

	servertest.Mux.HandleFunc("/push/project/container",
		func(w http.ResponseWriter, r *http.Request) {
			var mf, _, err = r.FormFile("pod")

			if err != nil {
				t.Error(err)
			}

			var hash = sha1.New()

			var _, eh = io.Copy(hash, mf)

			if eh != nil {
				t.Error(eh)
			}

			var gotSHA1Header = r.Header.Get("WeDeploy-Package-SHA1")
			var gotSHA1 = fmt.Sprintf("%x", hash.Sum(nil))

			if gotSHA1 != gotSHA1Header {
				t.Errorf("SHA1 from package doesn't match SHA1 from header.")
			}
		})

	var deploy, err = New("mycontainer")

	if err != nil {
		t.Errorf("Expected New error to be null, got %v instead", err)
	}

	err = deploy.Only()

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	globalconfigmock.Teardown()
	config.Teardown()
	os.Chdir(workingDir)
	servertest.Teardown()
}
