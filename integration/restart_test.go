package integration

import (
	"net/http"
	"testing"

	"github.com/wedeploy/cli/servertest"
	"github.com/wedeploy/cli/tdata"
)

func TestRestartProjectQuiet(t *testing.T) {
	var handled bool
	defer Teardown()
	Setup()

	servertest.IntegrationMux.HandleFunc("/projects/foo",
		tdata.ServerJSONFileHandler("mocks/restart/foo/project_response.json"))

	servertest.IntegrationMux.HandleFunc("/restart/project",
		func(w http.ResponseWriter, r *http.Request) {
			handled = true

			var wantQS = "projectId=foo"

			if r.URL.RawQuery != wantQS {
				t.Errorf("Wanted %v, got %v instead", wantQS, r.URL.RawQuery)
			}
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

func TestRestartProjectQuietFromCurrentWorkingDirectoryContext(t *testing.T) {
	var handled bool
	defer Teardown()
	Setup()

	servertest.IntegrationMux.HandleFunc("/projects/foo",
		tdata.ServerJSONFileHandler("mocks/restart/foo/project_response.json"))

	servertest.IntegrationMux.HandleFunc("/restart/project",
		func(w http.ResponseWriter, r *http.Request) {
			handled = true

			var wantQS = "projectId=foo"

			if r.URL.RawQuery != wantQS {
				t.Errorf("Wanted %v, got %v instead", wantQS, r.URL.RawQuery)
			}
		})

	var cmd = &Command{
		Args: []string{"restart", "--remote", "local", "--quiet"},
		Env:  []string{"WEDEPLOY_CUSTOM_HOME=" + GetLoginHome()},
		Dir:  "mocks/restart/foo",
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

func TestRestartContainerQuiet(t *testing.T) {
	var handled bool
	defer Teardown()
	Setup()

	servertest.IntegrationMux.HandleFunc("/projects/foo/containers/bar",
		tdata.ServerJSONFileHandler("mocks/restart/foo/bar/container_response.json"))

	servertest.IntegrationMux.HandleFunc("/restart/container",
		func(w http.ResponseWriter, r *http.Request) {
			handled = true

			var wantQS = "projectId=foo&containerId=bar"

			if r.URL.RawQuery != wantQS {
				t.Errorf("Wanted %v, got %v instead", wantQS, r.URL.RawQuery)
			}
		})

	var cmd = &Command{
		Args: []string{"restart", "-u", "bar.foo.wedeploy.me", "-q"},
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

func TestRestartContainerQuietFromCurrentWorkingDirectoryContext(t *testing.T) {
	var handled bool
	defer Teardown()
	Setup()

	servertest.IntegrationMux.HandleFunc("/projects/foo/containers/bar",
		tdata.ServerJSONFileHandler("mocks/restart/foo/bar/container_response.json"))

	servertest.IntegrationMux.HandleFunc("/restart/container",
		func(w http.ResponseWriter, r *http.Request) {
			handled = true

			var wantQS = "projectId=foo&containerId=bar"

			if r.URL.RawQuery != wantQS {
				t.Errorf("Wanted %v, got %v instead", wantQS, r.URL.RawQuery)
			}
		})

	var cmd = &Command{
		Args: []string{"restart", "--remote", "local", "-q"},
		Env:  []string{"WEDEPLOY_CUSTOM_HOME=" + GetLoginHome()},
		Dir:  "mocks/restart/foo/bar",
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
