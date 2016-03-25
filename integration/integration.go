package integration

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"testing"

	"github.com/launchpad-project/cli/configstore"
	"github.com/launchpad-project/cli/servertest"
)

// Expect structure
type Expect struct {
	Stderr   string
	Stdout   string
	ExitCode int
}

// Command structure
type Command struct {
	Args     []string
	Env      []string
	Dir      string
	Stdin    io.Reader
	Stderr   *bytes.Buffer
	Stdout   *bytes.Buffer
	ExitCode int
}

var (
	// ErrExitCodeNotAvailable is used for exit code retrieval failure
	ErrExitCodeNotAvailable = errors.New("Exit code not available")

	binaryDir string
	binary    string

	errStream io.Writer = os.Stderr
)

// GetExitCode tries to retrieve the exit code from an exit error
func GetExitCode(err error) int {
	if err == nil {
		return 0
	}

	if exit, ok := err.(*exec.ExitError); ok {
		if process, ok := exit.Sys().(syscall.WaitStatus); ok {
			return process.ExitStatus()
		}
	}

	fmt.Fprintln(errStream, err.Error())
	panic(ErrExitCodeNotAvailable)
}

func GetRegularHome() string {
	return getHomePath("home")
}

func GetLoginHome() string {
	return getHomePath("login")
}

func GetLogoutHome() string {
	return getHomePath("logout")
}

func Setup() {
	servertest.SetupIntegration()
	setupLoginHome()
}

func Teardown() {
	servertest.TeardownIntegration()
}

func (e *Expect) AssertExact(t *testing.T, cmd *Command) {
	if cmd.ExitCode != e.ExitCode {
		t.Errorf("Wanted exit code %v, got %v instead", e.ExitCode, cmd.ExitCode)
	}

	errString := cmd.Stderr.String()
	outString := cmd.Stdout.String()

	if errString != e.Stderr {
		t.Errorf("Wanted Stderr %v, got %v instead", e.Stderr, errString)
	}

	if outString != e.Stdout {
		t.Errorf("Wanted Stdout %v, got %v instead", e.Stdout, outString)
	}
}

// Run runs the command
func (cmd *Command) Run() {
	child := exec.Command(binary, cmd.Args...)

	if cmd.Stdin != nil {
		child.Stdin = cmd.Stdin
	}

	var serr = new(bytes.Buffer)
	var sout = new(bytes.Buffer)

	var customHome, err = filepath.Abs("./mocks/home")

	if err != nil {
		panic(err)
	}

	cmd.Env = append(cmd.Env, "LAUNCHPAD_CUSTOM_HOME="+customHome)
	cmd.Env = append(cmd.Env, os.Environ()...)

	if cmd.Dir != "" {
		cmd.Dir, err = filepath.Abs(cmd.Dir)

		if err != nil {
			panic(err)
		}
	}

	child.Env = cmd.Env
	child.Dir = cmd.Dir
	child.Stderr = serr
	child.Stdout = sout
	cmd.Stderr = serr
	cmd.Stdout = sout
	cmd.ExitCode = GetExitCode(child.Run())
}

func compile() {
	var workingDir, err = os.Getwd()

	if err != nil {
		panic(err)
	}

	binaryDir, err = filepath.Abs(filepath.Join(binaryDir, ".."))

	if err != nil {
		panic(err)
	}

	os.Chdir(binaryDir)
	cmd := exec.Command("go", "build")
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout

	if err := cmd.Run(); err != nil {
		panic(err)
	}

	binary, err = filepath.Abs(filepath.Join(binaryDir, "cli"))

	if err != nil {
		panic(err)
	}

	os.Chdir(workingDir)
}

func getHomePath(home string) string {
	var path, err = filepath.Abs("./mocks/" + home)

	if err != nil {
		panic(err)
	}

	return path
}

func init() {
	compile()
}

func setupLoginHome() {
	var csg = &configstore.Store{
		Name: "global",
		Path: filepath.Join(GetLoginHome(), "/.launchpad.json"),
	}

	csg.Set("endpoint", servertest.IntegrationServer.URL)
	csg.Set("username", "foo")
	csg.Set("password", "bar")
	csg.Save()
}
