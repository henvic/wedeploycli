package integration

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/launchpad-project/cli/servertest"
)

func TestLoggedOut(t *testing.T) {
	var cmd = &Command{
		Args: []string{"projects", "-v"},
		Env:  []string{"LAUNCHPAD_CUSTOM_HOME=" + GetLogoutHome()},
	}

	var e = &Expect{
		Stderr:   "Please run \"launchpad login\" first.\n",
		ExitCode: 1,
	}

	cmd.Run()
	e.AssertExact(t, cmd)
}

func TestProjects(t *testing.T) {
	defer Teardown()
	Setup()

	servertest.IntegrationMux.HandleFunc("/api/projects", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, `[{
    "name": "Image Server",
    "id": "images"
}]`)
	})

	var cmd = &Command{
		Args: []string{"projects"},
		Env:  []string{"LAUNCHPAD_CUSTOM_HOME=" + GetLoginHome()},
	}

	var e = &Expect{
		Stdout:   "images (Image Server)\ntotal 1\n",
		ExitCode: 0,
	}

	cmd.Run()
	e.AssertExact(t, cmd)
}
