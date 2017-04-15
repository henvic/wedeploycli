package integration

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/wedeploy/cli/servertest"
	"github.com/wedeploy/cli/tdata"
)

func TestLink(t *testing.T) {
	if !checkDockerIsUp() {
		t.Skipf("Docker is not up")
	}

	defer Teardown()
	Setup()

	servertest.IntegrationMux.HandleFunc("/projects",
		func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "POST" {
				t.Errorf("Expected method to be POST, got %v instead", r.Method)
			}

			var wantRequestURI = "/projects"
			if r.RequestURI != wantRequestURI {
				t.Errorf("Expected RequestURI %v, got %v instead", wantRequestURI, r.RequestURI)
			}

			var body, err = ioutil.ReadAll(r.Body)

			if err != nil {
				t.Errorf("Wanted err to be nil, got %v instead", err)
			}

			var want = tdata.FromFile("mocks/home/bucket/project/project.json")
			var got = string(body)

			if want != got {
				t.Errorf("Wanted sent file to have contents %v, got %v instead", want, got)
			}

			w.Header().Set("Content-type", "application/json; charset=UTF-8")
			fmt.Fprintf(w, "%v", tdata.FromFile("mocks/home/bucket/project/projects_post_response.json"))
		})

	servertest.IntegrationMux.HandleFunc("/deploy",
		func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "POST" {
				t.Errorf("Expected method to be POST, got %v instead", r.Method)
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
		Args: []string{"run", "--skip-local-infra", "--no-color"},
		Env: []string{
			"WEDEPLOY_CUSTOM_HOME=" + GetLoginHome()},
		Dir: "mocks/home/bucket/project/container",
	}

	var e = &Expect{
		ExitCode: 0,
		Stdout:   tdata.FromFile("mocks/link/link_new_project"),
	}

	cmd.Run()
	e.Assert(t, cmd)
}

func TestLinkEmptyJSON(t *testing.T) {
	if !checkDockerIsUp() {
		t.Skipf("Docker is not up")
	}

	defer Teardown()
	Setup()

	var cmd = &Command{
		Args: []string{"run", "--skip-local-infra", "--project", "foo", "--no-color"},
		Env:  []string{"WEDEPLOY_CUSTOM_HOME=" + GetLoginHome()},
		Dir:  "mocks/link/empty-json",
	}

	var e = &Expect{
		ExitCode: 1,
		Stderr:   "Can not read container with no project: unexpected end of JSON input",
	}

	cmd.Run()
	e.Assert(t, cmd)
}

func TestLinkToProject(t *testing.T) {
	if !checkDockerIsUp() {
		t.Skipf("Docker is not up")
	}

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
			fmt.Fprintf(w, "%v", tdata.FromFile("mocks/home/bucket/container-outside-project/projects_post_response_alt_id.json"))
		})

	servertest.IntegrationMux.HandleFunc("/deploy",
		func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "POST" {
				t.Errorf("Expected method to be POST, got %v instead", r.Method)
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
		Args: []string{"run", "--skip-local-infra", "--project", "bar", "--no-color"},
		Env: []string{
			"WEDEPLOY_CUSTOM_HOME=" + GetLoginHome()},
		Dir: "mocks/home/bucket/container-outside-project",
	}

	var e = &Expect{
		ExitCode: 0,
		Stdout:   tdata.FromFile("mocks/link/link"),
	}

	cmd.Run()
	e.Assert(t, cmd)
}

func TestLinkToProjectServerFailure(t *testing.T) {
	if !checkDockerIsUp() {
		t.Skipf("Docker is not up")
	}

	defer Teardown()
	Setup()

	servertest.IntegrationMux.HandleFunc("/projects",
		func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "POST" {
				t.Errorf("Expected method to be POST, got %v instead", r.Method)
			}

			var wantRequestURI = "/projects"
			if r.RequestURI != wantRequestURI {
				t.Errorf("Expected RequestURI %v, got %v instead", wantRequestURI, r.RequestURI)
			}

			w.Header().Set("Content-type", "application/json; charset=UTF-8")
			fmt.Fprintf(w, "%v", tdata.FromFile("mocks/home/bucket/project/container/projects_post_response_alt_id.json"))
		})

	servertest.IntegrationMux.HandleFunc("/projects/app/services",
		func(w http.ResponseWriter, r *http.Request) {
			// this is a hack to make the link test more robust
			// a nicer approach would be to clear the strings and match, though
			time.Sleep(5 * time.Millisecond)
			w.WriteHeader(500)
		})

	var cmd = &Command{
		Args: []string{"run", "--skip-local-infra", "--no-color"},
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
		`Linking errors:`,
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
	if !checkDockerIsUp() {
		t.Skipf("Docker is not up")
	}

	defer Teardown()
	Setup()

	servertest.IntegrationMux.HandleFunc("/projects",
		func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "POST" {
				t.Errorf("Expected method to be POST, got %v instead", r.Method)
			}

			var wantRequestURI = "/projects"
			if r.RequestURI != wantRequestURI {
				t.Errorf("Expected RequestURI %v, got %v instead", wantRequestURI, r.RequestURI)
			}

			w.Header().Set("Content-type", "application/json; charset=UTF-8")
			fmt.Fprintf(w, "%v", tdata.FromFile("mocks/home/bucket/project/container/projects_post_response_alt_id.json"))
		})

	servertest.IntegrationMux.HandleFunc("/projects/app/services",
		func(w http.ResponseWriter, r *http.Request) {
			// this is a hack to make the link test more robust
			// a nicer approach would be to clear the strings and match, though
			time.Sleep(5 * time.Millisecond)
			w.WriteHeader(500)
		})

	var cmd = &Command{
		Args: []string{"run", "--skip-local-infra", "--no-color", "--quiet"},
		Env: []string{
			"WEDEPLOY_CUSTOM_HOME=" + GetLoginHome()},
		Dir: "mocks/home/bucket/project/container",
	}

	cmd.Run()

	if cmd.ExitCode != 1 {
		t.Errorf("Unexpected exit code: %v", cmd.ExitCode)
	}

	var wantErrsContains = []string{
		`Linking errors:`,
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
		Args: []string{"run", "--skip-local-infra", "--remote=foo"},
		Env: []string{
			"WEDEPLOY_CUSTOM_HOME=" + GetLoginHome()},
		Dir: "mocks/home/bucket/project/container",
	}

	cmd.Run()

	if cmd.ExitCode != 1 {
		t.Errorf("Unexpected exit code %v", cmd.ExitCode)
	}

	var got = cmd.Stderr.String()
	var want = "unknown flag: --remote"

	if !strings.Contains(got, want) {
		t.Errorf("Error message doesn't contain expected value %v, got %v instead", want, got)
	}
}

func TestLinkRemoteShortcutError(t *testing.T) {
	defer Teardown()
	Setup()

	var cmd = &Command{
		Args: []string{"run", "--skip-local-infra", "-r=foo"},
		Env: []string{
			"WEDEPLOY_CUSTOM_HOME=" + GetLoginHome()},
		Dir: "mocks/home/bucket/project/container",
	}

	cmd.Run()

	if cmd.ExitCode != 1 {
		t.Errorf("Unexpected exit code %v", cmd.ExitCode)
	}

	var got = cmd.Stderr.String()
	var want = "unknown shorthand flag: 'r' in -r=foo"

	if !strings.Contains(got, want) {
		t.Errorf("Error message doesn't contain expected value %v, got %v instead", want, got)
	}
}

func TestLinkHostWithContainerError(t *testing.T) {
	defer Teardown()
	Setup()

	var cmd = &Command{
		Args: []string{"run", "--skip-local-infra", "--project", "foo", "--container", "bar"},
		Env: []string{
			"WEDEPLOY_CUSTOM_HOME=" + GetLoginHome()},
		Dir: "mocks/home/bucket/project/container",
	}

	cmd.Run()

	if cmd.ExitCode != 1 {
		t.Errorf("Unexpected exit code %v", cmd.ExitCode)
	}

	var got = cmd.Stderr.String()
	var want = "unknown flag: --container"

	if !strings.Contains(got, want) {
		t.Errorf("Error message doesn't contain expected value %v, got %v instead", want, got)
	}
}

func checkDockerIsUp() bool {
	var params = []string{
		"version", "--format", "{{.Client.Version}}",
	}

	var versionErrBuf bytes.Buffer
	var version = exec.Command("docker", params...)
	version.Stderr = &versionErrBuf

	return version.Run() == nil
}
