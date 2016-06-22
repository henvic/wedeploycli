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

	servertest.IntegrationMux.HandleFunc("/restart/project",
		func(w http.ResponseWriter, r *http.Request) {})

	var cmd = &Command{
		Args: []string{"link"},
		Env: []string{
			"WEDEPLOY_CUSTOM_HOME=" + GetLoginHome(),
			"WEDEPLOY_OVERRIDE_LOCAL_ENDPOINT=true"},
		Dir: "mocks/home/bucket/project/container",
	}

	var e = &Expect{
		ExitCode: 0,
		Stdout:   tdata.FromFile("mocks/link"),
	}

	cmd.Run()
	e.Assert(t, cmd)
}
