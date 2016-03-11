package context

import (
	"os"
	"path/filepath"
	"testing"
)

var workingDir, _ = os.Getwd()

type rootDirectoryProvider struct {
	file, dir, want string
}

type rootDirectoryFailureProvider struct {
	file, dir string
	want      error
}

var rootDirectoryCases = []rootDirectoryProvider{
	{"file", "mocks/list/basic", "mocks/list/basic"},
	{"file", "mocks/list/basic/level/", "mocks/list/basic"},
	{"file", "mocks/list/basic/level/2", "mocks/list/basic"},
	{"file", "mocks/list/basic/level/2/4", "mocks/list/basic"},
	{"file", "mocks/list/basic/level/3", "mocks/list/basic"},
}

var rootDirectoryFailureCases = []rootDirectoryFailureProvider{
	{"file", "mocks/list", os.ErrNotExist},
	{"file", "mocks/list/nothing", os.ErrNotExist},
	{"file", "mocks/list/nothing/here", os.ErrNotExist},
}

func TestGetTargetFileDirectory(t *testing.T) {
	setSysRoot("./mocks")

	for _, each := range rootDirectoryCases {
		if err := os.Chdir(filepath.Join(workingDir, each.dir)); err != nil {
			t.Error(err)
		}

		var directory, err = getRootDirectory(sysRoot, each.file)

		if err != nil {
			t.Error(err)
		}

		var want = filepath.Join(workingDir, each.want)
		want, _ = filepath.Abs(want)

		if directory != want {
			t.Errorf("Wanted to find config at %s, got %s instead", want, directory)
		}
	}

	os.Chdir(workingDir)
	setSysRoot("/")
}

func TestGetTargetFileDirectoryFailure(t *testing.T) {
	setSysRoot("./mocks")

	for _, each := range rootDirectoryFailureCases {
		if err := os.Chdir(filepath.Join(workingDir, each.dir)); err != nil {
			t.Error(err)
		}

		var _, err = getRootDirectory(sysRoot, each.file)

		if each.want != err {
			t.Errorf("Expected error %s, got %s instead", each.want, err)
		}
	}

	os.Chdir(workingDir)
	setSysRoot("/")
}

func TestGlobalContext(t *testing.T) {
	setSysRoot("./mocks")
	os.Chdir(filepath.Join(workingDir, "mocks/list/basic"))

	var context, configurations = Get()
	var wantContext = "global"

	if context.Scope != wantContext {
		t.Errorf("Expected context to be %s, got %s instead", wantContext, context.Scope)
	}

	if configurations != nil {
		t.Errorf("Unexpected configuration error: %v", configurations)
	}

	os.Chdir(workingDir)
	setSysRoot("/")
}

func TestProjectAndContainerInvalidContext(t *testing.T) {
	setSysRoot("./mocks")
	os.Chdir(filepath.Join(workingDir, "mocks/schizophrenic"))

	var _, configurations = Get()

	if configurations != ErrContainerInProjectRoot {
		t.Errorf("Expected to have %v error, got %v instead", ErrContainerInProjectRoot, configurations)
	}

	os.Chdir(workingDir)
	setSysRoot("/")
}

func TestProjectContext(t *testing.T) {
	setSysRoot("./mocks")
	var projectDir = filepath.Join(workingDir, "mocks/project")
	os.Chdir(projectDir)

	var context, err = Get()

	if context.Scope != "project" {
		t.Errorf("Expected context to be project, got %s instead", context.Scope)
	}

	if context.ProjectRoot != projectDir {
		t.Errorf("Wanted projectDir %s, got %s instead", projectDir, context.ProjectRoot)
	}

	if context.ContainerRoot != "" {
		t.Errorf("Expected container root to be empty, got %s instead", context.ContainerRoot)
	}

	if err != nil {
		t.Errorf("Unexpected context error: %v", err)
	}

	os.Chdir(workingDir)
	setSysRoot("/")
}

func TestContainerContext(t *testing.T) {
	setSysRoot("./mocks")
	var projectDir = filepath.Join(workingDir, "mocks/project")
	var containerDir = filepath.Join(projectDir, "container")
	os.Chdir(containerDir)

	var context, err = Get()

	if context.Scope != "container" {
		t.Errorf("Expected context to be container, got %s instead", context.Scope)
	}

	if context.ProjectRoot != projectDir {
		t.Errorf("Wanted projectDir %s, got %s instead", projectDir, context.ProjectRoot)
	}

	if context.ContainerRoot != containerDir {
		t.Errorf("Wanted containerDir %s, got %s instead", containerDir, context.ContainerRoot)
	}

	if err != nil {
		t.Errorf("Unexpected context error: %v", err)
	}

	os.Chdir(workingDir)
	setSysRoot("/")
}

func TestOrphanContainerContext(t *testing.T) {
	setSysRoot("./mocks")
	os.Chdir(filepath.Join(workingDir, "mocks/orphan_container"))

	var context, err = Get()

	if context.Scope != "global" {
		t.Errorf("Expected context to be global, got %s instead", context)
	}

	if err != nil {
		t.Errorf("Expected error to be nil, got %s instead", err)
	}

	if context.ContainerRoot != "" {
		t.Errorf("Expected Container root to be empty, got %s instead", context.ContainerRoot)
	}

	if context.ProjectRoot != "" {
		t.Errorf("Expected Project root to be empty, got %s instead", context.ProjectRoot)
	}

	os.Chdir(workingDir)
	setSysRoot("/")
}

func TestInvalidContext(t *testing.T) {
	setSysRoot("./mocks")
	os.Chdir(filepath.Join(workingDir, "mocks/schizophrenic"))

	var context, err = Get()

	if context.Scope != "project" {
		t.Errorf("Expected context type to be project, got %s instead", context.Scope)
	}

	if context.ProjectRoot != filepath.Join(workingDir, "mocks/schizophrenic") {
		t.Errorf("Unexpected project root")
	}

	if context.ContainerRoot != "" {
		t.Errorf("Unexpected container root")
	}

	if err != ErrContainerInProjectRoot {
		t.Errorf("Expected error to be "+ErrContainerInProjectRoot.Error()+", got %s instead", err)
	}

	os.Chdir(workingDir)
	setSysRoot("/")
}
