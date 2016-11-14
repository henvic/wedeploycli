package usercontext

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/hashicorp/errwrap"
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
	cx.Scope = "global"

	if errProject != nil && os.IsNotExist(errProject) {
		return cx, nil
	}

	if errProject != nil {
		return cx, errwrap.Wrapf("Error trying to read project: {{err}}", errProject)
	}

	cx.Scope = "project"

	var container, errContainer = GetContainerRootDirectory(project)

	if errContainer != nil && os.IsNotExist(errContainer) {
		return cx, nil
	}

	if errContainer != nil {
		return cx, errwrap.Wrapf("Error trying to read container: {{err}}", errContainer)
	}

	if filepath.Dir(container) == filepath.Dir(project) {
		return cx, ErrContainerInProjectRoot
	}

	cx.Scope = "container"
	cx.ContainerRoot = container

	return cx, nil
}

// GetProjectRootDirectory returns project dir for the current scope
func GetProjectRootDirectory(delimiter string) (string, error) {
	return getRootDirectory(delimiter, "project.json")
}

// GetContainerRootDirectory returns container dir for the current scope
func GetContainerRootDirectory(delimiter string) (string, error) {
	return getRootDirectory(delimiter, "container.json")
}

func getRootDirectory(delimiter, file string) (dir string, err error) {
	dir, err = os.Getwd()

	if err != nil {
		return "", err
	}

	return findresource.GetRootDirectory(dir, delimiter, file)
}
