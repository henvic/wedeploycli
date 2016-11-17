package usercontext

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/wedeploy/cli/findresource"
)

var workingDir, _ = os.Getwd()

func TestGetProjectRootDirectory(t *testing.T) {
	chdir(filepath.Join(workingDir, "mocks/project/container"))
	defer chdir(workingDir)

	var dir, err = GetProjectRootDirectory(".")

	if err != nil {
		t.Errorf("Wanted err to be nil, got %v instead", err)
	}

	var wantDir = filepath.Join(workingDir, "mocks/project")

	if dir != wantDir {
		t.Errorf("Wanted dir to be %v, got %v instead", wantDir, dir)
	}
}

func TestGetContainerRootDirectory(t *testing.T) {
	chdir(filepath.Join(workingDir, "mocks/project/container"))
	defer chdir(workingDir)

	var dir, err = GetContainerRootDirectory(".")

	if err != nil {
		t.Errorf("Wanted err to be nil, got %v instead", err)
	}

	var wantDir = filepath.Join(workingDir, "mocks/project/container")

	if dir != wantDir {
		t.Errorf("Wanted dir to be %v, got %v instead", wantDir, dir)
	}
}

func TestGlobalContext(t *testing.T) {
	setSysRootOrPanic(abs("./mocks"))
	chdir(filepath.Join(workingDir, "mocks/"))
	defer chdir(workingDir)
	defer setSysRootOrPanic(abs("/"))

	var (
		cx  = Context{}
		err = cx.Load()
	)

	var wantContext = GlobalScope

	if cx.Scope != wantContext {
		t.Errorf("Expected context to be %s, got %s instead", wantContext, cx.Scope)
	}

	if err != nil {
		t.Errorf("Unexpected configuration error: %v", err)
	}
}

func TestProjectAndContainerInvalidContext(t *testing.T) {
	setSysRootOrPanic(abs("./mocks"))
	chdir(filepath.Join(workingDir, "mocks/schizophrenic"))
	defer chdir(workingDir)
	defer setSysRootOrPanic(abs("/"))

	var (
		cx  = Context{}
		err = cx.Load()
	)

	if err != ErrContainerInProjectRoot {
		t.Errorf("Expected to have %v error, got %v instead", ErrContainerInProjectRoot, err)
	}
}

func TestProjectContext(t *testing.T) {
	var projectDir = filepath.Join(workingDir, "mocks/project")
	setSysRootOrPanic(abs("./mocks"))
	chdir(projectDir)
	defer setSysRootOrPanic(abs("./mocks"))
	defer chdir(workingDir)

	var (
		cx  = Context{}
		err = cx.Load()
	)

	if cx.Scope != ProjectScope {
		t.Errorf("Expected context to be project, got %s instead", cx.Scope)
	}

	if cx.ProjectRoot != projectDir {
		t.Errorf("Wanted projectDir %s, got %s instead", projectDir, cx.ProjectRoot)
	}

	if cx.ContainerRoot != "" {
		t.Errorf("Expected container root to be empty, got %s instead", cx.ContainerRoot)
	}

	if err != nil {
		t.Errorf("Unexpected context error: %v", err)
	}
}

func TestContainerContext(t *testing.T) {
	var projectDir = filepath.Join(workingDir, "mocks/project")
	var containerDir = filepath.Join(projectDir, "container")

	setSysRootOrPanic(abs("./mocks"))
	chdir(containerDir)
	defer setSysRootOrPanic(abs("./mocks"))
	defer chdir(workingDir)

	var (
		cx  = Context{}
		err = cx.Load()
	)

	if cx.Scope != ContainerScope {
		t.Errorf("Expected context to be container, got %s instead", cx.Scope)
	}

	if cx.ProjectRoot != projectDir {
		t.Errorf("Wanted projectDir %s, got %s instead", projectDir, cx.ProjectRoot)
	}

	if cx.ContainerRoot != containerDir {
		t.Errorf("Wanted containerDir %s, got %s instead", containerDir, cx.ContainerRoot)
	}

	if err != nil {
		t.Errorf("Unexpected context error: %v", err)
	}
}

func TestOrphanContainerContext(t *testing.T) {
	setSysRootOrPanic(abs("./mocks"))
	chdir(filepath.Join(workingDir, "mocks/orphan_container"))
	defer setSysRootOrPanic(abs("/"))
	defer chdir(workingDir)

	var (
		cx  = Context{}
		err = cx.Load()
	)

	if cx.Scope != "global" {
		t.Errorf("Expected context to be global, got %s instead", cx)
	}

	if err != nil {
		t.Errorf("Expected error to be nil, got %s instead", err)
	}

	if cx.ContainerRoot != "" {
		t.Errorf("Expected Container root to be empty, got %s instead", cx.ContainerRoot)
	}

	if cx.ProjectRoot != "" {
		t.Errorf("Expected Project root to be empty, got %s instead", cx.ProjectRoot)
	}
}

func TestInvalidContext(t *testing.T) {
	setSysRootOrPanic(abs("./mocks"))
	chdir(filepath.Join(workingDir, "mocks/schizophrenic"))
	defer setSysRootOrPanic(abs("/"))
	defer chdir(workingDir)

	var (
		cx  = Context{}
		err = cx.Load()
	)

	if cx.Scope != ProjectScope {
		t.Errorf("Expected context type to be project, got %s instead", cx.Scope)
	}

	if cx.ProjectRoot != filepath.Join(workingDir, "mocks/schizophrenic") {
		t.Errorf("Unexpected project root")
	}

	if cx.ContainerRoot != "" {
		t.Errorf("Unexpected container root")
	}

	if err != ErrContainerInProjectRoot {
		t.Errorf("Expected error to be "+ErrContainerInProjectRoot.Error()+", got %s instead", err)
	}
}

func setSysRootOrPanic(dir string) {
	if err := findresource.SetSysRoot(dir); err != nil {
		panic(err)
	}
}

func chdir(dir string) {
	if ech := os.Chdir(dir); ech != nil {
		panic(ech)
	}
}

func abs(path string) string {
	var abs, err = filepath.Abs(path)

	if err != nil {
		panic(err)
	}

	return abs
}
