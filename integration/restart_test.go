package integration

import (
	"fmt"
	"net/http"
	"sync"
	"testing"

	"github.com/wedeploy/cli/servertest"
	"github.com/wedeploy/cli/tdata"
)

func TestRestartInternalServerError(t *testing.T) {
	defer Teardown()
	Setup()

	servertest.IntegrationMux.HandleFunc(
		"/projects/foo",
		func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, `{
    "status": 500,
    "message": "Internal Server Error",
    "errors": [
        {
            "reason": "internalError",
			"context": {
				"message": "The request failed due to an internal error"
			}
        }
    ]
}`)
		})

	var cmd = &Command{
		Args: []string{"restart", "--remote", "local", "--project", "foo", "--quiet"},
		Env:  []string{"WEDEPLOY_CUSTOM_HOME=" + GetLoginHome()},
	}

	cmd.Run()

	if update {
		tdata.ToFile("mocks/restart/internal-server-error", cmd.Stderr.String())
	}

	var e = &Expect{
		Stderr:   tdata.FromFile("mocks/restart/internal-server-error"),
		ExitCode: 1,
	}

	e.Assert(t, cmd)
}

func TestRestartProjectQuiet(t *testing.T) {
	var handled bool
	defer Teardown()
	Setup()

	servertest.IntegrationMux.HandleFunc("/projects/foo",
		tdata.ServerJSONFileHandler("mocks/restart/foo/project_response.json"))

	servertest.IntegrationMux.HandleFunc("/projects/foo/services",
		tdata.ServerJSONFileHandler("mocks/restart/foo/services_response.json"))

	servertest.IntegrationMux.HandleFunc("/projects/foo/restart",
		func(w http.ResponseWriter, r *http.Request) {
			handled = true
		})

	var cmd = &Command{
		Args: []string{"restart", "--remote", "local", "--project", "foo", "--quiet"},
		Env:  []string{"WEDEPLOY_CUSTOM_HOME=" + GetLoginHome()},
	}

	var e = &Expect{
		ExitCode: 0,
	}

	cmd.Run()
	e.Assert(t, cmd)

	if !handled {
		t.Errorf("Restart request not handled.")
	}
}
func TestRestartServiceQuiet(t *testing.T) {
	var handled bool
	var handledMutex sync.Mutex
	defer Teardown()
	Setup()

	servertest.IntegrationMux.HandleFunc("/projects/foo",
		tdata.ServerJSONFileHandler("mocks/restart/foo/bar/project_response.json"))

	servertest.IntegrationMux.HandleFunc("/projects/foo/services/bar",
		tdata.ServerJSONFileHandler("mocks/restart/foo/bar/service_response.json"))

	servertest.IntegrationMux.HandleFunc("/projects/foo/services/bar/restart",
		func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "POST" {
				t.Errorf("Expected method to be POST")
			}

			handledMutex.Lock()
			handled = true
			handledMutex.Unlock()
		})

	var cmd = &Command{
		Args: []string{"restart", "--project", "foo", "--service", "bar", "--remote", "local", "-q"},
		Env:  []string{"WEDEPLOY_CUSTOM_HOME=" + GetLoginHome()},
		Dir:  "mocks/restart/foo/bar",
	}

	cmd.Run()

	var e = &Expect{
		ExitCode: 0,
	}

	e.Assert(t, cmd)

	handledMutex.Lock()
	if !handled {
		t.Errorf("Restart request not handled.")
	}
	handledMutex.Unlock()
}
