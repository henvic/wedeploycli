package integration

import (
	"net/http"
	"testing"

	"github.com/launchpad-project/cli/servertest"
)

func TestRestartProject(t *testing.T) {
	var handled bool
	defer Teardown()
	Setup()

	servertest.IntegrationMux.HandleFunc("/api/restart/project",
		func(w http.ResponseWriter, r *http.Request) {
			handled = true

			var wantQS = "projectId=foo"

			if r.URL.RawQuery != wantQS {
				t.Errorf("Wanted %v, got %v instead", wantQS, r.URL.RawQuery)
			}
		})

	var cmd = &Command{
		Args: []string{"restart", "foo"},
		Env:  []string{"LAUNCHPAD_CUSTOM_HOME=" + GetLoginHome()},
	}

	var e = &Expect{
		ExitCode: 0,
	}

	cmd.Run()
	e.AssertExact(t, cmd)

	if !handled {
		t.Errorf("Restart request not handled.")
	}
}

func TestRestartContainer(t *testing.T) {
	var handled bool
	defer Teardown()
	Setup()

	servertest.IntegrationMux.HandleFunc("/api/restart/container",
		func(w http.ResponseWriter, r *http.Request) {
			handled = true

			var wantQS = "projectId=foo&containerId=bar"

			if r.URL.RawQuery != wantQS {
				t.Errorf("Wanted %v, got %v instead", wantQS, r.URL.RawQuery)
			}
		})

	var cmd = &Command{
		Args: []string{"restart", "foo", "bar"},
		Env:  []string{"LAUNCHPAD_CUSTOM_HOME=" + GetLoginHome()},
	}

	var e = &Expect{
		ExitCode: 0,
	}

	cmd.Run()
	e.AssertExact(t, cmd)

	if !handled {
		t.Errorf("Restart request not handled.")
	}
}
