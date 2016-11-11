package usercontext

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/wedeploy/cli/findresource"
)

// Context structure
type Context struct {
	Scope         string
	ProjectRoot   string
	ContainerRoot string
	Remote        string
	RemoteAddress string
	Endpoint      string
	Username      string
	Password      string
	Token         string
}

// ErrContainerInProjectRoot happens when a project.json and container.json is found at the same directory level
var ErrContainerInProjectRoot = errors.New("Container and project definition files at the same directory level")

// Get returns a Context object with the current scope
func Get() (*Context, error) {
	cx := &Context{}

	var project, errProject = GetProjectRootDirectory(findresource.GetSysRoot())

	cx.ProjectRoot = project

	if errProject != nil {
		cx.Scope = "global"
		return cx, nil
	}

	var container, errContainer = GetContainerRootDirectory(project)

	if errContainer != nil {
		cx.Scope = "project"
		return cx, checkContainerNotInProjectRoot(cx.ProjectRoot)
	}

	cx.Scope = "container"
	cx.ContainerRoot = container

	return cx, nil
}

// GetProjectRootDirectory returns project dir for the current scope
func GetProjectRootDirectory(delimiter string) (string, error) {
	return findresource.GetRootDirectory(delimiter, "project.json")
}

// GetContainerRootDirectory returns container dir for the current scope
func GetContainerRootDirectory(delimiter string) (string, error) {
	return findresource.GetRootDirectory(delimiter, "container.json")
}

func checkContainerNotInProjectRoot(projectRoot string) error {
	stat, err := os.Stat(filepath.Join(projectRoot, "container.json"))

	if err == nil && !stat.IsDir() {
		return ErrContainerInProjectRoot
	}

	return nil
}
