package findresource

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
	{"file", "mocks/", os.ErrNotExist},
	{"file", "", errReachedDirectoryTreeRoot},
}

func TestGetTargetFileDirectory(t *testing.T) {
	setSysRoot("./mocks")
	defer chdir(workingDir)
	defer setSysRoot("/")

	for _, each := range rootDirectoryCases {
		if err := os.Chdir(filepath.Join(workingDir, each.dir)); err != nil {
			t.Error(err)
		}

		var directory, err = GetRootDirectory(abs("."), sysRoot, each.file)

		if err != nil {
			t.Error(err)
		}

		var want = filepath.Join(workingDir, each.want)
		want, _ = filepath.Abs(want)

		if directory != want {
			t.Errorf("Wanted to find config at %s, got %s instead", want, directory)
		}
	}
}

func TestGetTargetFileDirectoryFailure(t *testing.T) {
	setSysRoot("./mocks")
	defer chdir(workingDir)
	defer setSysRoot("/")

	for _, each := range rootDirectoryFailureCases {
		if err := os.Chdir(filepath.Join(workingDir, each.dir)); err != nil {
			t.Error(err)
		}

		var _, err = GetRootDirectory(abs("."), sysRoot, each.file)

		if each.want != err {
			t.Errorf("Expected error %s, got %s instead", each.want, err)
		}
	}
}

func TestSetAndGetSysRoot(t *testing.T) {
	var old = sysRoot
	var want = "/foo/bar"
	SetSysRoot("/foo/bar")
	defer SetSysRoot(old)

	if sysRoot != want {
		t.Errorf("Wanted sysRoot to be %v, got %v instead", want, sysRoot)
	}

	if GetSysRoot() != want {
		t.Errorf("Wanted sysRoot to be %v, got %v instead", want, sysRoot)
	}
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

func abs(path string) string {
	var abs, err = filepath.Abs(path)

	if err != nil {
		panic(err)
	}

	return abs
}
