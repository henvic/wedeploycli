package cmdrunner

import (
	"fmt"
	"math/rand"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestRunTrue(t *testing.T) {
	var c = &Command{
		Name: "true",
	}

	if c.Started() {
		t.Errorf("Expected command to be not executed yet")
	}

	c.Run()

	if !c.Started() {
		t.Errorf("Expected command to be executed already")
	}

	var wantExitCode = 0

	if c.ExitCode != wantExitCode {
		t.Errorf("Expected exit code to be %v, got %v instead", wantExitCode, c.ExitCode)
	}
}

func TestRunFalse(t *testing.T) {
	var c = &Command{
		Name: "false",
	}

	if c.Started() {
		t.Errorf("Expected command to be not executed yet")
	}

	c.Run()

	if !c.Started() {
		t.Errorf("Expected command to be executed already")
	}

	var wantExitCode = 1

	if c.ExitCode != wantExitCode {
		t.Errorf("Expected exit code to be %v, got %v instead", wantExitCode, c.ExitCode)
	}
}

func TestCommandNotFound(t *testing.T) {
	var c = &Command{
		Name: fmt.Sprintf("invalid-command-%d", rand.Int()),
	}

	c.Run()

	if c.Error == nil || !strings.HasSuffix(c.Error.Error(), "executable file not found in $PATH") {
		t.Errorf("Expected error to be due to file not found, got %v instead", c.Error)
	}
}

func TestInvalidDir(t *testing.T) {
	var c = &Command{
		Dir: "foo",
	}

	c.Run()

	if c.Error == nil || !strings.HasSuffix(c.Error.Error(), "no such file or directory") {
		t.Errorf("Expected error to be due to file not found, got %v instead", c.Error)
	}
}

func TestIsCommandOutputNopNotStarted(t *testing.T) {
	var cmd = &Command{
		Name: "true",
	}

	defer func() {
		r := recover()
		want := "Command was not started"

		if r == nil || r.(error).Error() != want {
			t.Errorf("Expected command error panic: %v, got error %v instead", want, r)
		}
	}()

	IsCommandOutputNop(cmd)
}

func TestIsCommandOutputNopNotFinished(t *testing.T) {
	var cmd = &Command{
		Name: "sleep",
		Args: []string{"5"},
	}

	var mutex sync.Mutex

	go func() {
		mutex.Lock()
		cmd.Start()
		mutex.Unlock()
	}()

	time.Sleep(10 * time.Millisecond)

	defer func() {
		r := recover()
		want := "Command is still executing"

		if r == nil || r.(error).Error() != want {
			t.Errorf("Expected command error panic: %v, got error %v instead", want, r)
		}
	}()

	mutex.Lock()
	IsCommandOutputNop(cmd)
	mutex.Unlock()
}

func TestIsCommandTerminate(t *testing.T) {
	var cmd = &Command{
		Name: "sleep",
		Args: []string{"5"},
	}

	var mutex sync.Mutex

	go func() {
		mutex.Lock()
		cmd.Start()
		mutex.Unlock()
	}()

	time.Sleep(10 * time.Millisecond)
	mutex.Lock()

	if err := cmd.Terminate(); err != nil {
		t.Errorf("Error during termination of command: %v", err)
	}

	mutex.Unlock()
}

func TestIsCommandOutputNop(t *testing.T) {
	var cmd = &Command{
		Name: "true",
	}

	cmd.Run()

	if !IsCommandOutputNop(cmd) {
		t.Errorf("Expected command true to be a nop")
	}
}

func TestIsCommandOutputNopFalse(t *testing.T) {
	var cmd = &Command{
		Name: "false",
	}

	cmd.Run()

	if IsCommandOutputNop(cmd) {
		t.Errorf("Expected command false to not be a nop")
	}
}

func TestIsCommandOutputNopEchoFalse(t *testing.T) {
	var cmd = &Command{
		Name: "echo",
		Args: []string{"hi"},
	}

	cmd.Run()

	if IsCommandOutputNop(cmd) {
		t.Errorf(`Expected command "echo hi" to not be a nop`)
	}
}
