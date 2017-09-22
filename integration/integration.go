package integration

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io"
	"net"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"testing"

	"github.com/wedeploy/cli/config"
	"github.com/wedeploy/cli/defaults"
	"github.com/wedeploy/cli/envs"
	"github.com/wedeploy/cli/remotes"
	"github.com/wedeploy/cli/servertest"
	"github.com/wedeploy/cli/stringlib"
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

	binary    string
	errStream io.Writer = os.Stderr

	update bool
)

func init() {
	flag.BoolVar(&update, "update", false, "update golden files")
}

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

// GetBrokenHome gets mocked broken user's home path
func GetBrokenHome() string {
	return getHomePath("broken")
}

// GetRegularHome gets mocked regular user's home path
func GetRegularHome() string {
	return getHomePath("home")
}

// GetLoginHome gets mocked logged in user's home path
func GetLoginHome() string {
	return getHomePath("login")
}

// GetLogoutHome gets mocked logged out user's home path
func GetLogoutHome() string {
	return getHomePath("logout")
}

// Setup an integration test environment
func Setup() {
	servertest.SetupIntegration()
	setupLoginHome()
}

// Teardown an integration test environment
func Teardown() {
	servertest.TeardownIntegration()
}

// Assert tests if command executed exactly as described by Expect
func (e *Expect) Assert(t *testing.T, cmd *Command) {
	if cmd.ExitCode != e.ExitCode {
		t.Errorf("Wanted exit code %v, got %v instead", e.ExitCode, cmd.ExitCode)
	}

	errString := cmd.Stderr.String()
	outString := cmd.Stdout.String()

	if !stringlib.Similar(errString, e.Stderr) {
		t.Errorf("Wanted Stderr %v, got %v instead", e.Stderr, errString)
	}

	if !stringlib.Similar(outString, e.Stdout) {
		t.Errorf("Wanted Stdout %v, got %v instead", e.Stdout, outString)
	}
}

// Prepare prepares the executable command
func (cmd *Command) Prepare() *exec.Cmd {
	child := exec.Command(binary, cmd.Args...)
	cmd.setEnv()
	cmd.absDir()
	cmd.setChildChannels(child)

	return child
}

// Run runs the command
func (cmd *Command) Run() {
	c := cmd.Prepare()
	cmd.ExitCode = GetExitCode(c.Run())
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
	child.Stderr = cmd.Stderr
	child.Stdout = cmd.Stdout
}

func (cmd *Command) maybeSetHomeEnv() {
	for _, k := range cmd.Env {
		if strings.HasPrefix(k, fmt.Sprintf("%s=", envs.CustomHome)) {
			return
		}
	}

	var ch, err = filepath.Abs("./mocks/home")

	if err != nil {
		panic(err)
	}

	chEnv := fmt.Sprintf("%s=%s", envs.CustomHome, ch)
	cmd.Env = append(cmd.Env, chEnv)

}

func (cmd *Command) setEnv() {
	cmd.maybeSetHomeEnv()
	cmd.Env = append(cmd.Env, os.Environ()...)
}

func abs(path string) string {
	var abs, err = filepath.Abs(path)

	if err != nil {
		panic(err)
	}

	return abs
}

func chdir(dir string) {
	if ech := os.Chdir(dir); ech != nil {
		panic(ech)
	}
}

func getHomePath(home string) string {
	var path, err = filepath.Abs("./mocks/" + home)

	if err != nil {
		panic(err)
	}

	return path
}

func removeLoginHomeMock() {
	var weHomePath = GetLoginHome()
	var err = os.Remove(filepath.Join(weHomePath, ".we"))

	if err != nil && !os.IsNotExist(err) {
		panic(err)
	}

	var errMetrics = os.Remove(filepath.Join(weHomePath, ".we_metrics"))

	if errMetrics != nil && !os.IsNotExist(errMetrics) {
		panic(errMetrics)
	}
}

func setupLoginHome() {
	var file = filepath.Join(GetLoginHome(), ".we")
	removeLoginHomeMock()

	var mock = &config.Config{
		Path:    file,
		NoColor: true,
	}

	if err := mock.Load(); err != nil {
		panic(err)
	}

	var port = getIntegrationServerPort()

	var cr = mock.Remotes[defaults.CloudRemote]
	cr.Username = "admin"
	mock.Remotes[defaults.CloudRemote] = cr

	mock.Remotes["local"] = remotes.Entry{
		Infrastructure: fmt.Sprintf("http://localhost:%d", port),
		Service:        "wedeploy.me",
		Username:       "admin",
		Token:          "token",
	}

	if err := mock.Save(); err != nil {
		panic(err)
	}
}

func getIntegrationServerPort() int {
	var u, err = url.Parse(servertest.IntegrationServer.URL)

	if err != nil {
		panic(err)
	}

	_, port, err := net.SplitHostPort(u.Host)

	if err != nil {
		panic(err)
	}

	num, err := strconv.Atoi(port)

	if err != nil {
		panic(err)
	}

	return num
}

func removeAll(path string) {
	if err := os.RemoveAll(path); err != nil {
		panic(err)
	}
}
