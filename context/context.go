package context

import (
	"errors"
	"os"
	"path/filepath"
)

// Context structure
type Context struct {
	Scope         string
	ProjectRoot   string
	ContainerRoot string
}

var (
	// ErrContainerInProjectRoot happens when a project.json and container.json is found at the same directory level
	ErrContainerInProjectRoot = errors.New("Container and project definition files at the same directory level")

	sysRoot = "/"
)

// Get returns a Context object with the current scope
func Get() (*Context, error) {
	cx := &Context{}

	var project, errProject = getRootDirectory(sysRoot, "project.json")

	cx.ProjectRoot = project

	if errProject != nil {
		cx.Scope = "global"
		return cx, nil
	}

	var container, errContainer = getRootDirectory(project, "container.json")

	if errContainer != nil {
		cx.Scope = "project"
		return cx, checkContainerNotInProjectRoot(cx.ProjectRoot)
	}

	cx.Scope = "container"
	cx.ContainerRoot = container

	return cx, nil
}

func checkContainerNotInProjectRoot(projectRoot string) error {
	stat, err := os.Stat(projectRoot + "/container.json")

	if err == nil && !stat.IsDir() {
		return ErrContainerInProjectRoot
	}

	return nil
}

func getRootDirectory(delimiter, file string) (dir string, err error) {
	dir, err = os.Getwd()

	if err != nil {
		return "", err
	}

	stat, err := os.Stat(delimiter)

	if err != nil || !stat.IsDir() {
		return "", os.ErrNotExist
	}

	// sysRoot = / = upper-bound / The Power of Ten rule 2
	for dir != sysRoot && dir != delimiter {
		stat, err := os.Stat(filepath.Join(dir, file))

		if stat == nil {
			dir = filepath.Join(dir, "..")
			dir, _ = filepath.Abs(dir)
			continue
		}

		return dir, err
	}

	return "", os.ErrNotExist
}

func setSysRoot(dir string) {
	var err error

	sysRoot, err = filepath.Abs(dir)

	if err != nil {
		panic(err)
	}
}
