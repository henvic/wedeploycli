package integration

import (
	"fmt"
	"net/http"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/wedeploy/cli/servertest"
	"github.com/wedeploy/cli/tdata"
)

func TestLogsIncompatibleUse(t *testing.T) {
	defer Teardown()
	Setup()

	var cmd = &Command{
		Args: []string{"log"},
		Env:  []string{"WEDEPLOY_CUSTOM_HOME=" + GetLoginHome()},
		Dir:  "mocks/home/",
	}

	var e = &Expect{
		Stderr:   "Error: Incompatible use: --container requires --project or local project.json context",
		ExitCode: 1,
	}

	cmd.Run()
	e.Assert(t, cmd)
}

func TestLogs(t *testing.T) {
	defer Teardown()
	Setup()

	servertest.IntegrationMux.HandleFunc("/logs/foo",
		func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Query().Get("containerId") != "nodejs5143" {
				t.Errorf("Wrong value for containerId")
			}

			w.Header().Set("Content-type", "application/json; charset=UTF-8")
			fmt.Fprintf(w, tdata.FromFile("mocks/logs/logs_response.json"))
		})

	var cmd = &Command{
		Args: []string{
			"log",
			"nodejs5143.foo",
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

	servertest.IntegrationMux.HandleFunc("/logs/foo",
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

	servertest.IntegrationMux.HandleFunc("/logs/foo",
		func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Query().Get("containerId") != "nodejs5143" {
				t.Errorf("Wrong value for containerId")
			}

			w.Header().Set("Content-type", "application/json; charset=UTF-8")
			fmt.Fprintf(w, tdata.FromFile("mocks/logs/logs_response.json"))
		})

	var cmd = &Command{
		Args: []string{
			"log",
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

	servertest.IntegrationMux.HandleFunc("/logs/foo",
		func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Query().Get("containerId") != "nodejs5143" {
				t.Errorf("Wrong value for containerId")
			}

			w.Header().Set("Content-type", "application/json; charset=UTF-8")
			fmt.Fprintf(w, tdata.FromFile("mocks/logs/logs_response.json"))
		})

	var cmd = &Command{
		Args: []string{
			"log",
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

	servertest.IntegrationMux.HandleFunc("/logs/foo",
		func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Query().Get("containerId") != "nodejs5143" {
				t.Errorf("Wrong value for containerId")
			}

			w.Header().Set("Content-type", "application/json; charset=UTF-8")
			fmt.Fprintf(w, tdata.FromFile("mocks/logs/logs_response.json"))
		})

	var cmd = &Command{
		Args: []string{
			"log",
			"nodejs5143.foo.wedeploy.me",
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

	servertest.IntegrationMux.HandleFunc("/logs/foo",
		func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Query().Get("containerId") != "nodejs5143" {
				t.Errorf("Wrong value for containerId")
			}

			w.Header().Set("Content-type", "application/json; charset=UTF-8")
			fmt.Fprintf(w, "[]")
		})

	var cmd = &Command{
		Args: []string{
			"log",
			"nodejs5143.foo",
			"--watch"},
		Env: []string{"WEDEPLOY_CUSTOM_HOME=" + GetLoginHome()},
		Dir: "mocks/home/",
	}

	var e = &Expect{
		ExitCode: 0,
	}

	var wg sync.WaitGroup

	var c = cmd.Prepare()

	if err := c.Start(); err != nil {
		t.Errorf("Expected no error, got %v instead", err)
	}

	wg.Add(1)

	go func() {
		time.Sleep(100 * time.Millisecond)

		if err := c.Process.Signal(syscall.SIGINT); err != nil {
			panic(err)
		}

		time.Sleep(100 * time.Millisecond)

		if !c.ProcessState.Exited() {
			t.Error("Expected process to be gone.")
		}

		e.Assert(t, cmd)

		wg.Done()
	}()

	if err := c.Wait(); err != nil {
		t.Errorf("Expected no error, got %v instead", err)
	}

	wg.Wait()
}
