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
	"strings"
	"testing"

	"github.com/launchpad-project/cli/config"
	"github.com/launchpad-project/cli/globalconfigmock"
	"github.com/launchpad-project/cli/servertest"
)

func TestErrors(t *testing.T) {
	var e error = &Errors{
		List: map[string]error{
			"foo": os.ErrExist,
		},
	}

	var want = "Deploy error: map[foo:file already exists]"

	if fmt.Sprintf("%v", e) != want {
		t.Errorf("Wanted error %v, got %v instead.", e, want)
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

func TestZip(t *testing.T) {
	var workingDir, _ = os.Getwd()

	if err := os.Chdir(filepath.Join(workingDir, "mocks/myproject")); err != nil {
		t.Error(err)
	}

	config.Setup()

	var err = Zip(os.DevNull, "mycontainer")

	if err != nil {
		t.Errorf("Unexpected zipping error: %v", err)
	}

	config.Teardown()
	os.Chdir(workingDir)
}

func TestDeploy(t *testing.T) {
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

	var packageSHA1 = "5b4238302c12e91f0faf44bcc912eb230e8f3094"

	servertest.Mux.HandleFunc("/api/push/project/container",
		func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "POST" {
				t.Errorf("Unexpected method %v", r.Method)
			}

			if !strings.Contains(r.Header.Get("Content-Type"), "multipart/form-data;") {
				t.Errorf("Expected multipart/form-data")
			}

			if r.Header.Get("Launchpad-Package-SHA1") != packageSHA1 {
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

	var wantFeedback = "Ready! container.project.liferay.io\n"

	if bufOutStream.String() != wantFeedback {
		t.Errorf("Wanted feedback %v, got %v instead", wantFeedback, bufOutStream.String())
	}

	os.Remove(tmp.Name())
	globalconfigmock.Teardown()
	config.Teardown()
	os.Chdir(workingDir)
	servertest.Teardown()
	outStream = defaultOutStream
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

func TestDeployOnly(t *testing.T) {
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

	servertest.Mux.HandleFunc("/api/push/project/container",
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

			var gotSHA1Header = r.Header.Get("Launchpad-Package-SHA1")
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

	var wantFeedback = "Ready! container.project.liferay.io\n"

	if bufOutStream.String() != wantFeedback {
		t.Errorf("Wanted feedback %v, got %v instead", wantFeedback, bufOutStream.String())
	}

	globalconfigmock.Teardown()
	config.Teardown()
	os.Chdir(workingDir)
	servertest.Teardown()
	outStream = defaultOutStream
}
