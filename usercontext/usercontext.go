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
	Scope                Scope
	ProjectRoot          string
	ServiceRoot          string
	Remote               string
	Infrastructure       string
	InfrastructureDomain string
	ServiceDomain        string
	Username             string
	Password             string
	Token                string
}

// Scope is the type for the current mode of the CLI tool (based on current working directory)
type Scope string

const (
	// GlobalScope is the scope when no service on project or project is active
	GlobalScope Scope = "global"

	// ProjectScope is the scope for when a project is active, but no service is active
	ProjectScope Scope = "project"

	// ServiceScope is the scope for when a service on a project is active
	ServiceScope Scope = "service"
)

// ErrServiceInProjectRoot happens when a project.json and wedeploy.json is found at the same directory level
var ErrServiceInProjectRoot = errors.New("Service and project definition files at the same directory level")

func (cx *Context) loadProject() error {
	var project, errProject = GetProjectRootDirectory(findresource.GetSysRoot())

	if errProject != nil && os.IsNotExist(errProject) {
		return nil
	}

	if errProject != nil {
		return errwrap.Wrapf("Error trying to read project: {{err}}", errProject)
	}

	cx.Scope = ProjectScope
	cx.ProjectRoot = project
	return nil
}

func (cx *Context) loadService() error {
	var service, errService = GetServiceRootDirectory(cx.ProjectRoot)

	if errService != nil && os.IsNotExist(errService) {
		return nil
	}

	if errService != nil {
		return errwrap.Wrapf("Error trying to read service: {{err}}", errService)
	}

	if filepath.Dir(service) == filepath.Dir(cx.ProjectRoot) {
		return ErrServiceInProjectRoot
	}

	cx.Scope = ServiceScope
	cx.ServiceRoot = service
	return nil
}

// Load a Context object with the current scope
func (cx *Context) Load() error {
	cx.Scope = GlobalScope

	if err := cx.loadProject(); err != nil {
		return err
	}

	if err := cx.loadService(); err != nil {
		return err
	}

	return nil
}

// GetProjectRootDirectory returns project dir for the current scope
func GetProjectRootDirectory(delimiter string) (string, error) {
	return getRootDirectory(delimiter, "project.json")
}

// GetServiceRootDirectory returns service dir for the current scope
func GetServiceRootDirectory(delimiter string) (string, error) {
	return getRootDirectory(delimiter, "wedeploy.json")
}

func getRootDirectory(delimiter, file string) (dir string, err error) {
	dir, err = os.Getwd()

	if err != nil {
		return "", err
	}

	return findresource.GetRootDirectory(dir, delimiter, file)
}
