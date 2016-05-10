package integration

import (
	"net/http"
	"os"
	"testing"

	"github.com/launchpad-project/cli/servertest"
	"github.com/launchpad-project/cli/tdata"
)

func TestDeploy(t *testing.T) {
	defer Teardown()
	Setup()

	servertest.IntegrationMux.HandleFunc("/projects",
		func(w http.ResponseWriter, r *http.Request) {})

	servertest.IntegrationMux.HandleFunc("/containers",
		func(w http.ResponseWriter, r *http.Request) {})

	servertest.IntegrationMux.HandleFunc("/push/app/container",
		func(w http.ResponseWriter, r *http.Request) {
			var _, _, err = r.FormFile("pod")

			if err != nil {
				t.Error(err)
			}
		})

	var cmd = &Command{
		Args: []string{"deploy", "-q"},
		Env:  []string{"LAUNCHPAD_CUSTOM_HOME=" + GetLoginHome()},
		Dir:  "mocks/home/bucket/project/container",
	}

	var e = &Expect{
		ExitCode: 0,
		Stdout:   tdata.FromFile("mocks/deploy"),
	}

	cmd.Run()
	e.Assert(t, cmd)
}

func TestDeployOutputErrorMultiple(t *testing.T) {
	defer Teardown()
	Setup()

	var cmd = &Command{
		Args: []string{"deploy", "-o", "foo.pod"},
		Env:  []string{"LAUNCHPAD_CUSTOM_HOME=" + GetLoginHome()},
		Dir:  "mocks/home/bucket/images",
	}

	var e = &Expect{
		Stdout:   "",
		Stderr:   "Can only output a single container to file, not a whole project.",
		ExitCode: 1,
	}

	cmd.Run()
	e.Assert(t, cmd)
}

func TestDeployOutput(t *testing.T) {
	defer Teardown()
	Setup()

	var cmd = &Command{
		Args: []string{"deploy", "-o", os.DevNull, "--quiet"},
		Env:  []string{"LAUNCHPAD_CUSTOM_HOME=" + GetLoginHome()},
		Dir:  "mocks/home/bucket/project/container",
	}

	var e = &Expect{
		ExitCode: 0,
	}

	cmd.Run()
	e.Assert(t, cmd)
}
