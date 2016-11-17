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
	Scope         Scope
	ProjectRoot   string
	ContainerRoot string
	Remote        string
	RemoteAddress string
	Endpoint      string
	Username      string
	Password      string
	Token         string
}

// Scope is the type for the current mode of the CLI tool (based on current working directory)
type Scope string

const (
	// GlobalScope is the scope when no container on project or project is active
	GlobalScope Scope = "global"

	// ProjectScope is the scope for when a project is active, but no container is active
	ProjectScope Scope = "project"

	// ContainerScope is the scope for when a container on a project is active
	ContainerScope Scope = "container"
)

// ErrContainerInProjectRoot happens when a project.json and container.json is found at the same directory level
var ErrContainerInProjectRoot = errors.New("Container and project definition files at the same directory level")

// Load a Context object with the current scope
func (cx *Context) Load() error {
	cx.Scope = GlobalScope

	var project, errProject = GetProjectRootDirectory(findresource.GetSysRoot())

	if errProject != nil && os.IsNotExist(errProject) {
		return nil
	}

	cx.ProjectRoot = project

	if errProject != nil {
		return errwrap.Wrapf("Error trying to read project: {{err}}", errProject)
	}

	cx.Scope = ProjectScope

	var container, errContainer = GetContainerRootDirectory(project)

	if errContainer != nil && os.IsNotExist(errContainer) {
		return nil
	}

	if errContainer != nil {
		return errwrap.Wrapf("Error trying to read container: {{err}}", errContainer)
	}

	if filepath.Dir(container) == filepath.Dir(project) {
		return ErrContainerInProjectRoot
	}

	cx.Scope = ContainerScope
	cx.ContainerRoot = container

	return nil
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
