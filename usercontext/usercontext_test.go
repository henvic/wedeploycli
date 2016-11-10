package usercontext

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

	chdir(workingDir)
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

	chdir(workingDir)
	setSysRoot("/")
}

func TestGetProjectRootDirectory(t *testing.T) {
	chdir(filepath.Join(workingDir, "mocks/project/container"))

	var dir, err = GetProjectRootDirectory(".")

	if err != nil {
		t.Errorf("Wanted err to be nil, got %v instead", err)
	}

	var wantDir = filepath.Join(workingDir, "mocks/project")

	if dir != wantDir {
		t.Errorf("Wanted dir to be %v, got %v instead", wantDir, dir)
	}

	chdir(workingDir)
}

func TestGetContainerRootDirectory(t *testing.T) {
	chdir(filepath.Join(workingDir, "mocks/project/container"))

	var dir, err = GetContainerRootDirectory(".")

	if err != nil {
		t.Errorf("Wanted err to be nil, got %v instead", err)
	}

	var wantDir = filepath.Join(workingDir, "mocks/project/container")

	if dir != wantDir {
		t.Errorf("Wanted dir to be %v, got %v instead", wantDir, dir)
	}

	chdir(workingDir)
}

func TestGlobalContext(t *testing.T) {
	setSysRoot("./mocks")
	chdir(filepath.Join(workingDir, "mocks/list/basic"))

	var usercontext, configurations = Get()
	var wantContext = "global"

	if usercontext.Scope != wantContext {
		t.Errorf("Expected context to be %s, got %s instead", wantContext, usercontext.Scope)
	}

	if configurations != nil {
		t.Errorf("Unexpected configuration error: %v", configurations)
	}

	chdir(workingDir)
	setSysRoot("/")
}

func TestProjectAndContainerInvalidContext(t *testing.T) {
	setSysRoot("./mocks")
	chdir(filepath.Join(workingDir, "mocks/schizophrenic"))

	var _, configurations = Get()

	if configurations != ErrContainerInProjectRoot {
		t.Errorf("Expected to have %v error, got %v instead", ErrContainerInProjectRoot, configurations)
	}

	chdir(workingDir)
	setSysRoot("/")
}

func TestProjectContext(t *testing.T) {
	setSysRoot("./mocks")
	var projectDir = filepath.Join(workingDir, "mocks/project")
	chdir(projectDir)

	var usercontext, err = Get()

	if usercontext.Scope != "project" {
		t.Errorf("Expected context to be project, got %s instead", usercontext.Scope)
	}

	if usercontext.ProjectRoot != projectDir {
		t.Errorf("Wanted projectDir %s, got %s instead", projectDir, usercontext.ProjectRoot)
	}

	if usercontext.ContainerRoot != "" {
		t.Errorf("Expected container root to be empty, got %s instead", usercontext.ContainerRoot)
	}

	if err != nil {
		t.Errorf("Unexpected context error: %v", err)
	}

	chdir(workingDir)
	setSysRoot("/")
}

func TestContainerContext(t *testing.T) {
	setSysRoot("./mocks")
	var projectDir = filepath.Join(workingDir, "mocks/project")
	var containerDir = filepath.Join(projectDir, "container")
	chdir(containerDir)

	var usercontext, err = Get()

	if usercontext.Scope != "container" {
		t.Errorf("Expected context to be container, got %s instead", usercontext.Scope)
	}

	if usercontext.ProjectRoot != projectDir {
		t.Errorf("Wanted projectDir %s, got %s instead", projectDir, usercontext.ProjectRoot)
	}

	if usercontext.ContainerRoot != containerDir {
		t.Errorf("Wanted containerDir %s, got %s instead", containerDir, usercontext.ContainerRoot)
	}

	if err != nil {
		t.Errorf("Unexpected context error: %v", err)
	}

	chdir(workingDir)
	setSysRoot("/")
}

func TestOrphanContainerContext(t *testing.T) {
	setSysRoot("./mocks")
	chdir(filepath.Join(workingDir, "mocks/orphan_container"))

	var usercontext, err = Get()

	if usercontext.Scope != "global" {
		t.Errorf("Expected context to be global, got %s instead", usercontext)
	}

	if err != nil {
		t.Errorf("Expected error to be nil, got %s instead", err)
	}

	if usercontext.ContainerRoot != "" {
		t.Errorf("Expected Container root to be empty, got %s instead", usercontext.ContainerRoot)
	}

	if usercontext.ProjectRoot != "" {
		t.Errorf("Expected Project root to be empty, got %s instead", usercontext.ProjectRoot)
	}

	chdir(workingDir)
	setSysRoot("/")
}

func TestInvalidContext(t *testing.T) {
	setSysRoot("./mocks")
	chdir(filepath.Join(workingDir, "mocks/schizophrenic"))

	var usercontext, err = Get()

	if usercontext.Scope != "project" {
		t.Errorf("Expected context type to be project, got %s instead", usercontext.Scope)
	}

	if usercontext.ProjectRoot != filepath.Join(workingDir, "mocks/schizophrenic") {
		t.Errorf("Unexpected project root")
	}

	if usercontext.ContainerRoot != "" {
		t.Errorf("Unexpected container root")
	}

	if err != ErrContainerInProjectRoot {
		t.Errorf("Expected error to be "+ErrContainerInProjectRoot.Error()+", got %s instead", err)
	}

	chdir(workingDir)
	setSysRoot("/")
}

func chdir(dir string) {
	if ech := os.Chdir(dir); ech != nil {
		panic(ech)
	}
}

func setSysRoot(dir string) {
	var err error

	sysRoot, err = filepath.Abs(dir)

	if err != nil {
		panic(err)
	}
}
