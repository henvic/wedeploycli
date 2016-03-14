package hooks

import (
	"bytes"
	"crypto/md5"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/launchpad-project/api.go/jsonlib"
	"github.com/launchpad-project/cli/config"
	"github.com/launchpad-project/cli/context"
)

var (
	bufErrStream bytes.Buffer
	bufOutStream bytes.Buffer
)

func TestMain(m *testing.M) {
	var defaultErrStream = errStream
	var defaultOutStream = outStream
	errStream = &bufErrStream
	outStream = &bufOutStream
	ec := m.Run()
	errStream = defaultErrStream
	outStream = defaultOutStream
	os.Exit(ec)
}

func TestBuild(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Not testing hooks.Build() on Windows")
	}

	bufErrStream.Reset()
	bufOutStream.Reset()

	var workingDir, _ = os.Getwd()

	if err := os.Chdir(filepath.Join(workingDir, "mocks/project/simple-container")); err != nil {
		t.Error(err)
	}

	config.Setup()

	if err := Build(config.Context); err != nil {
		t.Errorf("Expected %v, got %v instead", nil, err)
	}

	if bufErrStream.Len() != 0 {
		t.Error("Error stream length different than 0")
	}

	var wantOutput = "before build\nduring build\nafter build\n"

	if bufOutStream.String() != wantOutput {
		t.Errorf("Output stream buffer doesn't match with expected value.")
	}

	os.Chdir(workingDir)
	config.Setup()
}

func TestBuildUnbuildable(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Not testing hooks.Build() on Windows")
	}

	var workingDir, _ = os.Getwd()

	if err := os.Chdir(filepath.Join(workingDir, "mocks/unbuildable")); err != nil {
		t.Error(err)
	}

	config.Setup()

	if err := Build(config.Context); err != ErrMissingHook {
		t.Errorf("Expected %v error, got %v instead", ErrMissingHook, err)
	}

	os.Chdir(workingDir)
	config.Setup()
}

func TestBuildProjectFailure(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Not testing hooks.Get() on Windows")
	}

	var workingDir, _ = os.Getwd()

	if err := os.Chdir(filepath.Join(workingDir, "mocks/missing-hooks")); err != nil {
		t.Error(err)
	}

	config.Setup()

	if err := Build(&context.Context{Scope: "project"}); err != ErrMissingHook {
		t.Errorf("Wanted %v, got %v instead", ErrMissingHook, err)
	}

	os.Chdir(workingDir)
	config.Setup()
}

func TestDeploy(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Not testing hooks.Build() on Windows")
	}

	bufErrStream.Reset()
	bufOutStream.Reset()

	var workingDir, _ = os.Getwd()

	if err := os.Chdir(filepath.Join(workingDir, "mocks/project/simple-container")); err != nil {
		t.Error(err)
	}

	config.Setup()

	if err := Deploy(config.Context); err != nil {
		t.Errorf("Expected %v, got %v instead", nil, err)
	}

	if bufErrStream.Len() != 0 {
		t.Error("Error stream length different than 0")
	}

	var wantOutput = "before deploy\nduring deploy\nafter deploy\n"

	if bufOutStream.String() != wantOutput {
		t.Errorf("Output stream buffer doesn't match with expected value.")
	}

	os.Chdir(workingDir)
	config.Setup()
}

func TestDeployProjectFailure(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Not testing hooks.Get() on Windows")
	}

	var workingDir, _ = os.Getwd()

	if err := os.Chdir(filepath.Join(workingDir, "mocks/missing-hooks")); err != nil {
		t.Error(err)
	}

	config.Setup()

	if err := Deploy(&context.Context{Scope: "project"}); err != ErrMissingHook {
		t.Errorf("Wanted %v, got %v instead", ErrMissingHook, err)
	}

	os.Chdir(workingDir)
	config.Setup()
}

func TestDeployUndeployable(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Not testing hooks.Deploy() on Windows")
	}

	var workingDir, _ = os.Getwd()

	if err := os.Chdir(filepath.Join(workingDir, "mocks/undeployable")); err != nil {
		t.Error(err)
	}

	config.Setup()

	if err := Deploy(config.Context); err != ErrMissingHook {
		t.Errorf("Expected %v error, got %v instead", ErrMissingHook, err)
	}

	os.Chdir(workingDir)
	config.Setup()
}

func TestGetContainer(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Not testing hooks.Get() on Windows")
	}

	var workingDir, _ = os.Getwd()

	if err := os.Chdir(filepath.Join(workingDir, "mocks/project/container")); err != nil {
		t.Error(err)
	}

	config.Setup()
	var hooks, err = Get("container")

	if err != nil {
		t.Error(err)
	}

	if hooks.BeforeBuild != "echo hello" ||
		hooks.Build != "ls -l" ||
		hooks.AfterBuild != "ps" ||
		hooks.BeforeDeploy != "who" ||
		hooks.Deploy != "git --help" ||
		hooks.AfterDeploy != "cal" {
		t.Errorf("Unexpected hook values")
	}

	jsonlib.AssertJSONMarshal(t, `{
        "before_build": "echo hello",
        "build": "ls -l",
        "after_build": "ps",
        "before_deploy": "who",
        "deploy": "git --help",
        "after_deploy": "cal"
    }`, hooks)

	os.Chdir(workingDir)
	config.Setup()
}

func TestGetProject(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Not testing hooks.Get() on Windows")
	}

	var workingDir, _ = os.Getwd()

	if err := os.Chdir(filepath.Join(workingDir, "mocks/project")); err != nil {
		t.Error(err)
	}

	config.Setup()
	var hooks, err = Get("project")

	if err != nil {
		t.Error(err)
	}

	if hooks.BeforeBuild != "echo hi" ||
		hooks.Build != "ls -la" ||
		hooks.AfterBuild != "time" ||
		hooks.BeforeDeploy != "pwd" ||
		hooks.Deploy != "ls" ||
		hooks.AfterDeploy != "date" {
		t.Errorf("Unexpected hook values")
	}

	jsonlib.AssertJSONMarshal(t, `{
        "before_build": "echo hi",
        "build": "ls -la",
        "after_build": "time",
        "before_deploy": "pwd",
        "deploy": "ls",
        "after_deploy": "date"
    }`, hooks)

	os.Chdir(workingDir)
	config.Setup()
}

func TestGetProjectFailure(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Not testing hooks.Get() on Windows")
	}

	var workingDir, _ = os.Getwd()

	if err := os.Chdir(filepath.Join(workingDir, "mocks/missing-hooks")); err != nil {
		t.Error(err)
	}

	config.Setup()

	if _, err := Get("project"); err != ErrMissingHook {
		t.Errorf("Wanted %v, got %v instead", ErrMissingHook, err)
	}

	os.Chdir(workingDir)
	config.Setup()
}

func TestRun(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Not testing hooks.Run() on Windows")
	}

	bufErrStream.Reset()
	bufOutStream.Reset()

	if err := Run("openssl md5 hooks.go"); err != nil {
		t.Errorf("Unexpected error %v when running md5 hooks.go", err)
	}

	h := md5.New()

	data, _ := ioutil.ReadFile("./hooks.go")
	io.WriteString(h, string(data))

	if !strings.Contains(bufOutStream.String(), fmt.Sprintf("%x", h.Sum(nil))) {
		t.Errorf("Expected Run() test to contain md5 output similar to crypto.md5")
	}

	if bufErrStream.Len() != 0 {
		t.Errorf("Unexpected err output")
	}
}

func TestRunAndExitOnFailureOnSuccess(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Not testing hooks.Run() on Windows")
	}

	bufErrStream.Reset()
	bufOutStream.Reset()

	RunAndExitOnFailure("openssl md5 hooks.go")

	h := md5.New()

	data, _ := ioutil.ReadFile("./hooks.go")
	io.WriteString(h, string(data))

	if !strings.Contains(bufOutStream.String(), fmt.Sprintf("%x", h.Sum(nil))) {
		t.Errorf("Expected Run() test to contain md5 output similar to crypto.md5")
	}

	if bufErrStream.Len() != 0 {
		t.Errorf("Unexpected err output")
	}
}

func TestRunAndExitOnFailureFailure(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Not testing hooks.Run() on Windows")
	}

	bufErrStream.Reset()
	bufOutStream.Reset()

	if os.Getenv("PING_CRASHER") == "1" {
		outStream = os.Stdout
		errStream = os.Stderr
		RunAndExitOnFailure("ping")
		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=TestRunAndExitOnFailureFailure")
	cmd.Stderr = errStream
	cmd.Stdout = outStream
	cmd.Env = append(os.Environ(), "PING_CRASHER=1")
	err := cmd.Run()

	if err.Error() != "exit status 1" {
		t.Errorf("Expected exit status 1 for ping process, got %v instead", err)
	}

	if bufErrStream.Len() == 0 {
		t.Error("Expected ping output to be piped to stderr")
	}
}
