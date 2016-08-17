package integration

import (
	"fmt"
	"net/http"
	"testing"
	"time"

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
		func(w http.ResponseWriter, r *http.Request) {
			// this is a hack to make the link test more robust
			// a nicer approach would be to clear the strings and match, though
			time.Sleep(5 * time.Millisecond)
			w.Header().Set("Content-type", "application/json; charset=UTF-8")
			fmt.Fprintf(w, tdata.FromFile("mocks/link/list.json"))
		})

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

func TestLinkRemoteError(t *testing.T) {
	defer Teardown()
	Setup()

	var cmd = &Command{
		Args: []string{"link", "--remote=foo"},
		Env: []string{
			"WEDEPLOY_CUSTOM_HOME=" + GetLoginHome()},
		Dir: "mocks/home/bucket/project/container",
	}

	var e = &Expect{
		ExitCode: 1,
		Stderr:   "Error: can not use command with a remote",
	}

	cmd.Run()
	e.Assert(t, cmd)
}
