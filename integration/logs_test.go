package integration

import (
	"fmt"
	"net/http"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/wedeploy/cli/servertest"
	"github.com/wedeploy/cli/tdata"
)

func TestLogsTooManyArguments(t *testing.T) {
	defer Teardown()
	Setup()

	var cmd = &Command{
		Args: []string{"log", "do", "re", "mi", "fa", "so", "la", "ti", "do"},
		Env:  []string{"WEDEPLOY_CUSTOM_HOME=" + GetLoginHome()},
		Dir:  "mocks/home/",
	}

	var e = &Expect{
		Stderr:   "Error: Invalid number of arguments: too many",
		ExitCode: 1,
	}

	cmd.Run()
	e.Assert(t, cmd)
}

func TestLogsIncompatibleUse(t *testing.T) {
	defer Teardown()
	Setup()

	var cmd = &Command{
		Args: []string{"log"},
		Env:  []string{"WEDEPLOY_CUSTOM_HOME=" + GetLoginHome()},
		Dir:  "mocks/home/",
	}

	var e = &Expect{
		Stderr:   "Error: Use flag --project or call this from inside a project directory",
		ExitCode: 1,
	}

	cmd.Run()
	e.Assert(t, cmd)
}

func TestLogIncompatibleUseProjectAndHostURLFlag(t *testing.T) {
	var cmd = &Command{
		Args: []string{"log", "--project", "foo", "-u", "mocks"},
	}

	cmd.Run()

	if cmd.Stdout.Len() != 0 {
		t.Errorf("Expected stdout to be empty, got %v instead", cmd.Stdout)
	}

	var wantErr = "Incompatible use: --project and --container are not allowed with host URL flag"

	if !strings.Contains(cmd.Stderr.String(), wantErr) {
		t.Errorf("Wanted stderr to have %v, got %v instead", wantErr, cmd.Stderr)
	}

	if cmd.ExitCode == 0 {
		t.Errorf("Expected exit code to be not 0, got %v instead", cmd.ExitCode)
	}
}

func TestLogs(t *testing.T) {
	defer Teardown()
	Setup()

	servertest.IntegrationMux.HandleFunc("/projects/foo/services/nodejs5143/logs",
		func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-type", "application/json; charset=UTF-8")
			fmt.Fprintf(w, tdata.FromFile("mocks/logs/logs_response.json"))
		})

	var cmd = &Command{
		Args: []string{
			"log",
			"-u",
			"nodejs5143-foo.wedeploy.me",
			"--no-color"},
		Env: []string{"WEDEPLOY_CUSTOM_HOME=" + GetLoginHome()},
		Dir: "mocks/home/",
	}

	var e = &Expect{
		Stdout:   tdata.FromFile("mocks/logs/logs_response_print"),
		ExitCode: 0,
	}

	cmd.Run()
	e.Assert(t, cmd)
}

func TestLogsFromCurrentWorkingOnProjectDirectoryContext(t *testing.T) {
	defer Teardown()
	Setup()

	servertest.IntegrationMux.HandleFunc("/projects/foo/logs",
		func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Query().Get("containerId") != "" {
				t.Errorf("Wrong value for containerId")
			}

			w.Header().Set("Content-type", "application/json; charset=UTF-8")
			fmt.Fprintf(w, tdata.FromFile("mocks/logs/logs_response.json"))
		})

	var cmd = &Command{
		Args: []string{
			"log",
			"--remote",
			"local",
			"--no-color"},
		Env: []string{"WEDEPLOY_CUSTOM_HOME=" + GetLoginHome()},
		Dir: "mocks/home/bucket/foo",
	}

	var e = &Expect{
		Stdout:   tdata.FromFile("mocks/logs/logs_response_print"),
		ExitCode: 0,
	}

	cmd.Run()
	e.Assert(t, cmd)
}

func TestLogsFromCurrentWorkingOnProjectDirectoryContextFilteringByContainer(t *testing.T) {
	defer Teardown()
	Setup()

	servertest.IntegrationMux.HandleFunc("/projects/foo/services/nodejs5143/logs",
		func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-type", "application/json; charset=UTF-8")
			fmt.Fprintf(w, tdata.FromFile("mocks/logs/logs_response.json"))
		})

	var cmd = &Command{
		Args: []string{
			"log",
			"--remote",
			"local",
			"--container=nodejs5143",
			"--no-color"},
		Env: []string{"WEDEPLOY_CUSTOM_HOME=" + GetLoginHome()},
		Dir: "mocks/home/bucket/foo",
	}

	var e = &Expect{
		Stdout:   tdata.FromFile("mocks/logs/logs_response_print"),
		ExitCode: 0,
	}

	cmd.Run()
	e.Assert(t, cmd)
}

func TestLogsFromCurrentWorkingOnContainerDirectoryContext(t *testing.T) {
	defer Teardown()
	Setup()

	servertest.IntegrationMux.HandleFunc("/projects/foo/services/nodejs5143/logs",
		func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-type", "application/json; charset=UTF-8")
			fmt.Fprintf(w, tdata.FromFile("mocks/logs/logs_response.json"))
		})

	var cmd = &Command{
		Args: []string{
			"log",
			"--remote",
			"local",
			"--no-color"},
		Env: []string{"WEDEPLOY_CUSTOM_HOME=" + GetLoginHome()},
		Dir: "mocks/home/bucket/foo/nodejs5143",
	}

	var e = &Expect{
		Stdout:   tdata.FromFile("mocks/logs/logs_response_print"),
		ExitCode: 0,
	}

	cmd.Run()
	e.Assert(t, cmd)
}

func TestLogsWithWeDeployDotMeAddress(t *testing.T) {
	defer Teardown()
	Setup()

	servertest.IntegrationMux.HandleFunc("/projects/foo/services/nodejs5143/logs",
		func(w http.ResponseWriter, r *http.Request) {
			if strings.Index(r.Host, "api.wedeploy.me:") != 0 {
				t.Errorf("Expected host to be localhost, got %v instead", r.Host)
			}

			w.Header().Set("Content-type", "application/json; charset=UTF-8")
			fmt.Fprintf(w, tdata.FromFile("mocks/logs/logs_response.json"))
		})

	var cmd = &Command{
		Args: []string{
			"log",
			"-u",
			"nodejs5143-foo.wedeploy.me",
			"--no-color"},
		Env: []string{"WEDEPLOY_CUSTOM_HOME=" + GetLoginHome()},
		Dir: "mocks/home/",
	}

	var e = &Expect{
		Stdout:   tdata.FromFile("mocks/logs/logs_response_print"),
		ExitCode: 0,
	}

	cmd.Run()
	e.Assert(t, cmd)
}

func TestWatch(t *testing.T) {
	defer Teardown()
	Setup()

	servertest.IntegrationMux.HandleFunc("/projects/foo/services/nodejs5143/logs",
		func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-type", "application/json; charset=UTF-8")
			fmt.Fprintf(w, "[]")
		})

	var cmd = &Command{
		Args: []string{
			"log",
			"-u",
			"nodejs5143-foo.wedeploy.me",
			"--watch"},
		Env: []string{"WEDEPLOY_CUSTOM_HOME=" + GetLoginHome()},
		Dir: "mocks/home/",
	}

	var e = &Expect{
		ExitCode: 0,
	}

	var c = cmd.Prepare()

	if err := c.Start(); err != nil {
		t.Errorf("Expected no error, got %v instead", err)
	}

	time.Sleep(100 * time.Millisecond)

	if err := c.Process.Signal(syscall.SIGINT); err != nil {
		panic(err)
	}

	time.Sleep(100 * time.Millisecond)

	if err := c.Wait(); err != nil {
		t.Errorf("Expected no error, got %v instead", err)
	}

	if !c.ProcessState.Exited() {
		t.Error("Expected process to be gone.")
	}

	e.Assert(t, cmd)
}
