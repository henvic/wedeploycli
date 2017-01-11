package cmdrunner

import (
	"bytes"
	"errors"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
)

// Command structure
type Command struct {
	Name         string
	Args         []string
	Env          []string
	Dir          string
	TeePipe      bool
	Process      *os.Process
	ProcessState *os.ProcessState
	Stdin        io.Reader
	Stderr       *bytes.Buffer
	Stdout       *bytes.Buffer
	ExitCode     int
	Error        error
	started      bool
	cmd          *exec.Cmd
}

// GetExitCode tries to retrieve the exit code from an exit error
func GetExitCode(err error) (int, error) {
	if err == nil {
		return 0, nil
	}

	if exit, ok := err.(*exec.ExitError); ok {
		if process, ok := exit.Sys().(syscall.WaitStatus); ok {
			return process.ExitStatus(), err
		}
	}

	return -1, err
}

// IsCommandOutputNop checks if a command exits with no output or error
func IsCommandOutputNop(cmd *Command) bool {
	if !cmd.Started() {
		panic(errors.New("Command was not started"))
	}

	if cmd.ProcessState == nil || !cmd.ProcessState.Exited() {
		panic(errors.New("Command is still executing"))
	}

	if cmd.Error != nil ||
		cmd.ExitCode != 0 ||
		cmd.Stderr.Len() != 0 ||
		cmd.Stdout.Len() != 0 {
		return false
	}

	return true
}

// Prepare prepares the executable command
func (cmd *Command) Prepare() *exec.Cmd {
	child := exec.Command(cmd.Name, cmd.Args...)
	cmd.absDir()
	cmd.setChildChannels(child)
	return child
}

// Start command
func (cmd *Command) Start() {
	cmd.cmd = cmd.Prepare()
	err := cmd.cmd.Start()

	cmd.Process = cmd.cmd.Process
	cmd.started = true

	if err != nil {
		cmd.ExitCode, cmd.Error = GetExitCode(err)
	}
}

// Wait for command to end
func (cmd *Command) Wait() {
	err := cmd.cmd.Wait()
	cmd.ProcessState = cmd.cmd.ProcessState
	cmd.ExitCode, cmd.Error = GetExitCode(err)
}

// Terminate sends a SIGTERM signal
func (cmd *Command) Terminate() error {
	var ec = make(chan error, 1)

	go func() {
		ec <- cmd.Process.Signal(syscall.SIGTERM)
	}()

	var err = <-ec
	cmd.Wait()

	return err
}

// Run runs the command
func (cmd *Command) Run() {
	cmd.Start()

	if cmd.Error == nil {
		cmd.Wait()
	}
}

// Started tells if the command was started or not
func (cmd *Command) Started() bool {
	return cmd.started
}

func (cmd *Command) absDir() {
	if cmd.Dir == "" {
		return
	}

	var dir, err = filepath.Abs(cmd.Dir)
	cmd.Dir = dir

	if err != nil {
		panic(err)
	}
}

func (cmd *Command) setChildChannels(child *exec.Cmd) {
	cmd.Stderr = new(bytes.Buffer)
	cmd.Stdout = new(bytes.Buffer)
	child.Env = cmd.Env
	child.Dir = cmd.Dir
	child.Stdin = cmd.Stdin

	if cmd.TeePipe {
		child.Stderr = io.MultiWriter(cmd.Stderr, os.Stderr)
		child.Stdout = io.MultiWriter(cmd.Stdout, os.Stdout)
	} else {
		child.Stderr = cmd.Stderr
		child.Stdout = cmd.Stdout
	}
}

// Run command
func Run(command string) error {
	return run(command)
}
