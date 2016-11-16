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

func TestMain(m *testing.M) {
	var (
		defaultSysRoot string
		ec             int
	)

	defer func() {
		setSysRootOrPanic(defaultSysRoot)
		os.Exit(ec)
	}()

	defaultSysRoot = GetSysRoot()
	setSysRootOrPanic("./mocks")
	ec = m.Run()
}

func TestGetTargetFileDirectory(t *testing.T) {
	setSysRootOrPanic("./mocks")
	defer chdir(workingDir)
	defer setSysRootOrPanic("/")

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

func TestGetTargetFileDirectoryWithRelativePathSearch(t *testing.T) {
	setSysRootOrPanic("./mocks")
	defer chdir(workingDir)
	defer setSysRootOrPanic("/")

	for _, each := range rootDirectoryCases {
		if err := os.Chdir(filepath.Join(workingDir, each.dir)); err != nil {
			t.Error(err)
		}

		var directory, err = GetRootDirectory(".", sysRoot, each.file)

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
	setSysRootOrPanic("./mocks")
	defer chdir(workingDir)
	defer setSysRootOrPanic("/")

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

func TestGetTargetFileDelimiterDirectoryNotFoundFailure(t *testing.T) {
	setSysRootOrPanic("./mocks")
	defer chdir(workingDir)
	defer setSysRootOrPanic("/")

	var _, err = GetRootDirectory(abs("./mocks/list/"), "./mocks/not", "basic")

	if !os.IsNotExist(err) {
		t.Errorf("Expected not exist error, got %v instead", err)
	}
}

func TestGetTargetFileDelimiterDirectoryNotDirectoryFailure(t *testing.T) {
	setSysRootOrPanic("./mocks")
	defer chdir(workingDir)
	defer setSysRootOrPanic("/")

	var _, err = GetRootDirectory(abs("./mocks/list/"), "./mocks/list/basic", "foo")

	if !os.IsNotExist(err) {
		t.Errorf("Expected not exist error, got %v instead", err)
	}
}

func TestGetTargetFileDirectoryNotFoundFailure(t *testing.T) {
	setSysRootOrPanic("./mocks")
	defer chdir(workingDir)
	defer setSysRootOrPanic("/")

	var _, err = GetRootDirectory(abs("./mocks/list/subdir/not/there"), sysRoot, "basic")

	if !os.IsNotExist(err) {
		t.Errorf("Expected not exist error, got %v instead", err)
	}
}

func TestGetTargetFileDirectoryNotDirectoryFailure(t *testing.T) {
	setSysRootOrPanic("./mocks")
	defer chdir(workingDir)
	defer setSysRootOrPanic("/")

	var _, err = GetRootDirectory(abs("./mocks/list/basic"), sysRoot, "foo")

	if !os.IsNotExist(err) {
		t.Errorf("Expected not exist error, got %v instead", err)
	}
}

func TestSetAndGetSysRoot(t *testing.T) {
	var old = sysRoot
	var want = "/foo/bar"
	setSysRootOrPanic("/foo/bar")
	defer setSysRootOrPanic(old)

	var got = GetSysRoot()

	if !filepath.IsAbs(got) {
		t.Errorf("SetSysRoot should be absolute, got %v instead", got)
	}

	if sysRoot != want {
		t.Errorf("Wanted sysRoot to be %v, got %v instead", want, sysRoot)
	}

	if got != want {
		t.Errorf("Wanted sysRoot to be %v, got %v instead", want, sysRoot)
	}
}

func chdir(dir string) {
	if ech := os.Chdir(dir); ech != nil {
		panic(ech)
	}
}

func setSysRootOrPanic(dir string) {
	if err := SetSysRoot(dir); err != nil {
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
