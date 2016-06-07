package integration

import (
	"net/http"
	"testing"

	"github.com/wedeploy/cli/servertest"
)

func TestRestartProject(t *testing.T) {
	var handled bool
	defer Teardown()
	Setup()

	servertest.IntegrationMux.HandleFunc("/restart/project",
		func(w http.ResponseWriter, r *http.Request) {
			handled = true

			var wantQS = "projectId=foo"

			if r.URL.RawQuery != wantQS {
				t.Errorf("Wanted %v, got %v instead", wantQS, r.URL.RawQuery)
			}
		})

	var cmd = &Command{
		Args: []string{"restart", "foo"},
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

func TestRestartContainer(t *testing.T) {
	var handled bool
	defer Teardown()
	Setup()

	servertest.IntegrationMux.HandleFunc("/restart/container",
		func(w http.ResponseWriter, r *http.Request) {
			handled = true

			var wantQS = "projectId=foo&containerId=bar"

			if r.URL.RawQuery != wantQS {
				t.Errorf("Wanted %v, got %v instead", wantQS, r.URL.RawQuery)
			}
		})

	var cmd = &Command{
		Args: []string{"restart", "foo", "bar"},
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
