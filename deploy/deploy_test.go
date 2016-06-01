package deploy

import (
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
)

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

			if r.Header.Get("Package-Size") == "0" {
				t.Errorf("Expected package size to have length > 0")
			}

			if r.Header.Get("Package-SHA1") != packageSHA1 {
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

			var gotSHA1Header = r.Header.Get("Package-SHA1")
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

func TestDeployWithHooks(t *testing.T) {
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

	var deploy, err = New("mycontainer")

	if err != nil {
		t.Errorf("Unexpected error %v", err)
	}

	err = deploy.HooksAndOnly(&Flags{
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

func TestDeployWithHooksFlagFalse(t *testing.T) {
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

	var deploy, err = New("mycontainer-hooks-failure")

	if err != nil {
		t.Errorf("Unexpected error %v", err)
	}

	err = deploy.HooksAndOnly(&Flags{
		Hooks: false,
	})

	if err != nil {
		t.Errorf("Unexpected error %v on deploy", err)
	}

	globalconfigmock.Teardown()
	config.Teardown()
	os.Chdir(workingDir)
	servertest.Teardown()
}

func TestDeployWithHooksNull(t *testing.T) {
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

	var deploy, err = New("my-hookless-container")

	if err != nil {
		t.Errorf("Unexpected error %v", err)
	}

	err = deploy.HooksAndOnly(&Flags{
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

	var deploy, err = New("container_before_hook_failure")

	if err != nil {
		t.Errorf("Unexpected error %v", err)
	}

	err = deploy.HooksAndOnly(&Flags{
		Hooks: true,
	})

	if err == nil || err.Error() != "exit status 1" {
		t.Errorf("Expected error didn't happen, got %v instead", err)
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

	var deploy, err = New("container_after_hook_failure")

	if err != nil {
		t.Errorf("Unexpected error %v", err)
	}

	err = deploy.HooksAndOnly(&Flags{
		Hooks: true,
	})

	if err == nil || err.Error() != "exit status 1" {
		t.Errorf("Expected error didn't happen, got %v instead", err)
	}

	globalconfigmock.Teardown()
	config.Teardown()
	os.Chdir(workingDir)
	servertest.Teardown()
}
