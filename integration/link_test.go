package integration

import (
	"net/http"
	"testing"

	"github.com/wedeploy/cli/servertest"
	"github.com/wedeploy/cli/tdata"
)

func TestLink(t *testing.T) {
	defer Teardown()
	Setup()

	servertest.IntegrationMux.HandleFunc("/projects",
		func(w http.ResponseWriter, r *http.Request) {})

	servertest.IntegrationMux.HandleFunc("/deploy",
		func(w http.ResponseWriter, r *http.Request) {})

	servertest.IntegrationMux.HandleFunc("/projects/app",
		tdata.ServerJSONFileHandler("mocks/link/list.json"))

	var cmd = &Command{
		Args: []string{"link", "--no-color"},
		Env: []string{
			"WEDEPLOY_CUSTOM_HOME=" + GetLoginHome()},
		Dir: "mocks/home/bucket/project/container",
	}

	var e = &Expect{
		ExitCode: 0,
		Stdout:   tdata.FromFile("mocks/link/link"),
	}

	cmd.Run()
	e.Assert(t, cmd)
}
