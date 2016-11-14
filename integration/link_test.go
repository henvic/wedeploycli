package integration

import (
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/wedeploy/cli/servertest"
	"github.com/wedeploy/cli/tdata"
)

func TestLink(t *testing.T) {
	defer Teardown()
	Setup()

	servertest.IntegrationMux.HandleFunc("/projects",
		func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "POST" {
				t.Errorf("Expected method to be POST, got %v instead", r.Method)
			}

			var wantRequestURI = "/projects?id=app"
			if r.RequestURI != wantRequestURI {
				t.Errorf("Expected RequestURI %v, got %v instead", wantRequestURI, r.RequestURI)
			}

			w.Header().Set("Content-type", "application/json; charset=UTF-8")
			fmt.Fprintf(w, "%v", tdata.FromFile("mocks/home/bucket/project/projects_post_response.json"))
		})

	servertest.IntegrationMux.HandleFunc("/deploy",
		func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "PUT" {
				t.Errorf("Expected method to be PUT, got %v instead", r.Method)
			}

			var wantRequestURI = "containerId=container&projectId=app"
			if !strings.Contains(r.RequestURI, wantRequestURI) {
				t.Errorf("Expected RequestURI %v, got %v instead", wantRequestURI, r.RequestURI)
			}

			w.Header().Set("Content-type", "application/json; charset=UTF-8")
			fmt.Fprintf(w, "%v", `{
    "id": "public",
    "type": "wedeploy/hosting"
}`)
		})

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

func TestLinkToProject(t *testing.T) {
	defer Teardown()
	Setup()

	servertest.IntegrationMux.HandleFunc("/projects",
		func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "POST" {
				t.Errorf("Expected method to be POST, got %v instead", r.Method)
			}

			var wantRequestURI = "/projects?id=bar"
			if r.RequestURI != wantRequestURI {
				t.Errorf("Expected RequestURI %v, got %v instead", wantRequestURI, r.RequestURI)
			}

			w.Header().Set("Content-type", "application/json; charset=UTF-8")
			fmt.Fprintf(w, "%v", tdata.FromFile("mocks/home/bucket/project/container/projects_post_response_alt_id.json"))
		})

	servertest.IntegrationMux.HandleFunc("/deploy",
		func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "PUT" {
				t.Errorf("Expected method to be PUT, got %v instead", r.Method)
			}

			var wantRequestURI = "containerId=container&projectId=bar"
			if !strings.Contains(r.RequestURI, wantRequestURI) {
				t.Errorf("Expected RequestURI %v, got %v instead", wantRequestURI, r.RequestURI)
			}

			w.Header().Set("Content-type", "application/json; charset=UTF-8")
			fmt.Fprintf(w, "%v", `{
    "id": "public",
    "type": "wedeploy/hosting"
}`)
		})

	servertest.IntegrationMux.HandleFunc("/projects/bar",
		func(w http.ResponseWriter, r *http.Request) {
			// this is a hack to make the link test more robust
			// a nicer approach would be to clear the strings and match, though
			time.Sleep(5 * time.Millisecond)
			w.Header().Set("Content-type", "application/json; charset=UTF-8")
			fmt.Fprintf(w, tdata.FromFile("mocks/link/list.json"))
		})

	var cmd = &Command{
		Args: []string{"link", "--project", "bar", "--no-color"},
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

func TestLinkToProjectServerFailure(t *testing.T) {
	defer Teardown()
	Setup()

	servertest.IntegrationMux.HandleFunc("/projects",
		func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "POST" {
				t.Errorf("Expected method to be POST, got %v instead", r.Method)
			}

			var wantRequestURI = "/projects?id=bar"
			if r.RequestURI != wantRequestURI {
				t.Errorf("Expected RequestURI %v, got %v instead", wantRequestURI, r.RequestURI)
			}

			w.Header().Set("Content-type", "application/json; charset=UTF-8")
			fmt.Fprintf(w, "%v", tdata.FromFile("mocks/home/bucket/project/container/projects_post_response_alt_id.json"))
		})

	servertest.IntegrationMux.HandleFunc("/deploy",
		func(w http.ResponseWriter, r *http.Request) {
			// this is a hack to make the link test more robust
			// a nicer approach would be to clear the strings and match, though
			time.Sleep(5 * time.Millisecond)
			w.WriteHeader(500)
		})

	var cmd = &Command{
		Args: []string{"link", "--project", "bar", "--no-color"},
		Env: []string{
			"WEDEPLOY_CUSTOM_HOME=" + GetLoginHome()},
		Dir: "mocks/home/bucket/project/container",
	}

	cmd.Run()

	if cmd.ExitCode != 1 {
		t.Errorf("Unexpected exit code: %v", cmd.ExitCode)
	}

	var wantErrsContains = []string{
		`Killing linking watcher after linking errors (use "we list" to see what is up).`,
		`Error: Linking errors`,
		`mocks/home/bucket/project/container: WeDeploy API error: 500 Internal Server Error`,
	}

	var got = cmd.Stderr.String()

	for _, we := range wantErrsContains {
		if !strings.Contains(got, we) {
			t.Errorf("Expected stderr to contain %v, but not found it", we)
		}
	}
}

func TestLinkToProjectServerFailureQuiet(t *testing.T) {
	defer Teardown()
	Setup()

	servertest.IntegrationMux.HandleFunc("/projects",
		func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "POST" {
				t.Errorf("Expected method to be POST, got %v instead", r.Method)
			}

			var wantRequestURI = "/projects?id=bar"
			if r.RequestURI != wantRequestURI {
				t.Errorf("Expected RequestURI %v, got %v instead", wantRequestURI, r.RequestURI)
			}

			w.Header().Set("Content-type", "application/json; charset=UTF-8")
			fmt.Fprintf(w, "%v", tdata.FromFile("mocks/home/bucket/project/container/projects_post_response_alt_id.json"))
		})

	servertest.IntegrationMux.HandleFunc("/deploy",
		func(w http.ResponseWriter, r *http.Request) {
			// this is a hack to make the link test more robust
			// a nicer approach would be to clear the strings and match, though
			time.Sleep(5 * time.Millisecond)
			w.WriteHeader(500)
		})

	var cmd = &Command{
		Args: []string{"link", "--project", "bar", "--no-color", "--quiet"},
		Env: []string{
			"WEDEPLOY_CUSTOM_HOME=" + GetLoginHome()},
		Dir: "mocks/home/bucket/project/container",
	}

	cmd.Run()

	if cmd.ExitCode != 1 {
		t.Errorf("Unexpected exit code: %v", cmd.ExitCode)
	}

	var wantErrsContains = []string{
		`Error: Linking errors`,
		`mocks/home/bucket/project/container: WeDeploy API error: 500 Internal Server Error`,
	}

	var got = cmd.Stderr.String()

	for _, we := range wantErrsContains {
		if !strings.Contains(got, we) {
			t.Errorf("Expected stderr to contain %v, but not found it; got %v instead", we, got)
		}
	}
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

	cmd.Run()

	if cmd.ExitCode != 1 {
		t.Errorf("Unexpected exit code %v", cmd.ExitCode)
	}

	var got = cmd.Stderr.String()
	var want = "Error: unknown flag: --remote"

	if !strings.Contains(got, want) {
		t.Errorf("Error message doesn't contain expected value %v, got %v instead", want, got)
	}
}

func TestLinkRemoteShortcutError(t *testing.T) {
	defer Teardown()
	Setup()

	var cmd = &Command{
		Args: []string{"link", "-r=foo"},
		Env: []string{
			"WEDEPLOY_CUSTOM_HOME=" + GetLoginHome()},
		Dir: "mocks/home/bucket/project/container",
	}

	cmd.Run()

	if cmd.ExitCode != 1 {
		t.Errorf("Unexpected exit code %v", cmd.ExitCode)
	}

	var got = cmd.Stderr.String()
	var want = "Error: unknown shorthand flag: 'r' in -r=foo"

	if !strings.Contains(got, want) {
		t.Errorf("Error message doesn't contain expected value %v, got %v instead", want, got)
	}
}

func TestLinkHostWithContainerError(t *testing.T) {
	defer Teardown()
	Setup()

	var cmd = &Command{
		Args: []string{"link", "foo.bar"},
		Env: []string{
			"WEDEPLOY_CUSTOM_HOME=" + GetLoginHome()},
		Dir: "mocks/home/bucket/project/container",
	}

	cmd.Run()

	if cmd.ExitCode != 1 {
		t.Errorf("Unexpected exit code %v", cmd.ExitCode)
	}

	var got = cmd.Stderr.String()
	var want = "Error: Container parameter is not allowed for this command"

	if !strings.Contains(got, want) {
		t.Errorf("Error message doesn't contain expected value %v, got %v instead", want, got)
	}
}
