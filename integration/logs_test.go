package integration

import (
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/launchpad-project/cli/servertest"
	"github.com/launchpad-project/cli/tdata"
)

func TestLogs(t *testing.T) {
	defer Teardown()
	Setup()

	servertest.IntegrationMux.HandleFunc(
		"/logs/foo/nodejs5143/foo_nodejs5143_sqimupf5tfsf9iylzpg3e4zj",
		tdata.ServerJSONFileHandler("../logs/mocks/logs_response.json"))

	var cmd = &Command{
		Args: []string{"logs", "foo", "nodejs5143", "foo_nodejs5143_sqimupf5tfsf9iylzpg3e4zj"},
		Env:  []string{"LAUNCHPAD_CUSTOM_HOME=" + GetLoginHome()},
		Dir:  "mocks/home/",
	}

	var e = &Expect{
		Stdout:   tdata.FromFile("../logs/mocks/logs_response_print"),
		ExitCode: 0,
	}

	cmd.Run()
	e.Assert(t, cmd)
}

func TestWatch(t *testing.T) {
	defer Teardown()
	Setup()

	servertest.IntegrationMux.HandleFunc(
		"/logs/foo/nodejs5143/foo_nodejs5143_sqimupf5tfsf9iylzpg3e4zj",
		tdata.ServerJSONHandler("[]"))

	var cmd = &Command{
		Args: []string{"logs", "foo", "nodejs5143", "foo_nodejs5143_sqimupf5tfsf9iylzpg3e4zj", "-f"},
		Env:  []string{"LAUNCHPAD_CUSTOM_HOME=" + GetLoginHome()},
		Dir:  "mocks/home/",
	}

	var e = &Expect{
		ExitCode: 0,
	}

	var wg sync.WaitGroup

	var c = cmd.Prepare()

	c.Start()

	wg.Add(1)

	go func() {
		time.Sleep(100 * time.Millisecond)

		if err := syscall.Kill(c.Process.Pid, syscall.SIGINT); err != nil {
			panic(err)
		}

		time.Sleep(100 * time.Millisecond)

		if !c.ProcessState.Exited() {
			t.Error("Expected process to be gone.")
		}

		e.Assert(t, cmd)

		wg.Done()
	}()

	c.Wait()
	wg.Wait()
}
