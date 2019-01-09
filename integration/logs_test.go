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

	cmd.Run()

	if update {
		tdata.ToFile("mocks/logs/too-many-arguments", cmd.Stderr.String())
	}

	var e = &Expect{
		Stderr:   tdata.FromFile("mocks/logs/too-many-arguments"),
		ExitCode: 1,
	}

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

	cmd.Run()

	if update {
		tdata.ToFile("mocks/home/logs-incompatible-use", cmd.Stderr.String())
	}

	var e = &Expect{
		Stderr:   tdata.FromFile("mocks/home/logs-incompatible-use"),
		ExitCode: 1,
	}

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

	var wantErr = "Incompatible use: --project and --service are not allowed with host URL flag"

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
			w.Header().Set("Content-Type", "application/json; charset=UTF-8")
			_, _ = fmt.Fprint(w, tdata.FromFile("mocks/logs/logs_response.json"))
		})

	var cmd = &Command{
		Args: []string{
			"log",
			"-u",
			"nodejs5143-foo.wedeploy.me",
			"--watch=false",
			"--no-color"},
		Env: []string{"WEDEPLOY_CUSTOM_HOME=" + GetLoginHome()},
		Dir: "mocks/home/",
	}

	cmd.Run()

	if update {
		tdata.ToFile("mocks/logs/logs_response_print", cmd.Stdout.String())
	}

	var e = &Expect{
		Stdout:   tdata.FromFile("mocks/logs/logs_response_print"),
		ExitCode: 0,
	}

	e.Assert(t, cmd)
}

func TestLogsFromCurrentWorkingOnProjectDirectoryContext(t *testing.T) {
	defer Teardown()
	Setup()

	servertest.IntegrationMux.HandleFunc("/projects/foo/logs",
		func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Query().Get("serviceId") != "" {
				t.Errorf("Wrong value for serviceId")
			}

			w.Header().Set("Content-Type", "application/json; charset=UTF-8")
			_, _ = fmt.Fprint(w, tdata.FromFile("mocks/logs/logs_response.json"))
		})

	var cmd = &Command{
		Args: []string{
			"log",
			"--remote",
			"local",
			"-p",
			"foo",
			"--watch=false",
			"--no-color"},
		Env: []string{"WEDEPLOY_CUSTOM_HOME=" + GetLoginHome()},
		Dir: "mocks/home/bucket/foo",
	}

	cmd.Run()

	if update {
		tdata.ToFile("mocks/logs/logs_response_print", cmd.Stdout.String())
	}

	var e = &Expect{
		Stdout:   tdata.FromFile("mocks/logs/logs_response_print"),
		ExitCode: 0,
	}

	e.Assert(t, cmd)
}

func TestLogsWithLocalhostAddress(t *testing.T) {
	defer Teardown()
	Setup()

	servertest.IntegrationMux.HandleFunc("/projects/foo/services/nodejs5143/logs",
		func(w http.ResponseWriter, r *http.Request) {
			if strings.Index(r.Host, "localhost:") != 0 {
				t.Errorf("Expected host to be localhost, got %v instead", r.Host)
			}

			w.Header().Set("Content-Type", "application/json; charset=UTF-8")
			_, _ = fmt.Fprint(w, tdata.FromFile("mocks/logs/logs_response.json"))
		})

	var cmd = &Command{
		Args: []string{
			"log",
			"-u",
			"nodejs5143-foo.wedeploy.me",
			"--watch=false",
			"--no-color"},
		Env: []string{"WEDEPLOY_CUSTOM_HOME=" + GetLoginHome()},
		Dir: "mocks/home/",
	}

	cmd.Run()

	if update {
		tdata.ToFile("mocks/logs/logs_response_print", cmd.Stdout.String())
	}

	var e = &Expect{
		Stdout:   tdata.FromFile("mocks/logs/logs_response_print"),
		ExitCode: 0,
	}

	e.Assert(t, cmd)
}

func TestWatch(t *testing.T) {
	defer Teardown()
	Setup()

	servertest.IntegrationMux.HandleFunc("/projects/foo/services/nodejs5143/logs",
		func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json; charset=UTF-8")
			_, _ = fmt.Fprintf(w, "[]")
		})

	var cmd = &Command{
		Args: []string{
			"log",
			"-u",
			"nodejs5143-foo.wedeploy.me"},
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
