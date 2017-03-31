package inspector

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"

	"github.com/hashicorp/errwrap"
	"github.com/wedeploy/cli/containers"
	"github.com/wedeploy/cli/findresource"
	"github.com/wedeploy/cli/projects"
	"github.com/wedeploy/cli/templates"
	"github.com/wedeploy/cli/usercontext"
	"github.com/wedeploy/cli/verbose"
)

// GetSpec of the passed type
func GetSpec(t interface{}) []string {
	var fields []string
	val := reflect.ValueOf(t)
	for i := 0; i < val.NumField(); i++ {
		field := val.Type().Field(i)
		fields = append(fields, fmt.Sprintf("%v %v", field.Name, field.Type))
	}

	return fields
}

// ContextOverview for the context visualization
type ContextOverview struct {
	Scope             usercontext.Scope
	ProjectRoot       string
	ContainerRoot     string
	ProjectID         string
	ServiceID         string
	ProjectContainers []containers.ContainerInfo
}

func (overview *ContextOverview) loadProjectPackage(directory string) error {
	var projectPath, project, perr = getProjectPackage(directory)

	switch {
	case os.IsNotExist(perr):
		return nil
	case perr != nil:
		return errwrap.Wrapf("Can not load project context on "+projectPath+": {{err}}", perr)
	}

	overview.Scope = usercontext.ProjectScope
	overview.ProjectRoot = projectPath
	overview.ProjectID = project.ID

	return overview.loadProjectContainersList()
}

func (overview *ContextOverview) loadProjectContainersList() error {
	var list, err = containers.GetListFromDirectory(overview.ProjectRoot)

	if err != nil {
		return errwrap.Wrapf("Error while trying to read list of containers on project: {{err}}", err)
	}

	for i, l := range list {
		list[i].Location = filepath.Join(overview.ProjectRoot, l.Location)
	}

	overview.ProjectContainers = list
	return nil
}

func (overview *ContextOverview) loadContainer(directory string) error {
	var containerPath, cp, cerr = getContainerPackage(directory)

	switch {
	case os.IsNotExist(cerr):
	case cerr != nil:
		return errwrap.Wrapf("Can not load container context on "+containerPath+": {{err}}", cerr)
	default:
		if overview.Scope == usercontext.ProjectScope {
			overview.Scope = usercontext.ContainerScope
		}

		overview.ContainerRoot = containerPath
		overview.ServiceID = cp.ID
	}

	return nil
}

// Load the context overview for a given directory
func (overview *ContextOverview) Load(directory string) error {
	overview.Scope = usercontext.GlobalScope

	if err := overview.loadProjectPackage(directory); err != nil {
		return err
	}

	if err := overview.loadContainer(directory); err != nil {
		return err
	}

	return nil
}

// InspectContext on a given directory, filtering by format
func InspectContext(format, directory string) (string, error) {
	// Can not rely on values on usercontext.Context given that
	// they are global and we accept the directory parameter
	var overview = ContextOverview{}

	if err := overview.Load(directory); err != nil {
		return "", err
	}

	return templates.ExecuteOrList(format, overview)
}

// InspectProject on a given directory, filtering by format
func InspectProject(format, directory string) (string, error) {
	var projectPath, project, perr = getProjectPackage(directory)

	switch {
	case os.IsNotExist(perr):
		return "", errwrap.Wrapf("Inspection failure: can not find project", perr)
	case perr != nil:
		return "", perr
	}

	verbose.Debug("Reading project at " + projectPath)
	return templates.ExecuteOrList(format, project)
}

func getProjectPackage(directory string) (path string, project *projects.ProjectPackage, err error) {
	var projectPath, cerr = getProjectRootDirectory(directory)

	if cerr != nil {
		return "", nil, cerr
	}

	project, err = projects.Read(projectPath)

	if err != nil {
		return projectPath, nil, errwrap.Wrapf("Inspection failure on project: {{err}}", err)
	}

	return projectPath, project, nil
}

func getContainerPackage(directory string) (path string, cp *containers.ContainerPackage, err error) {
	var containerPath, cerr = getContainerRootDirectory(directory)

	if cerr != nil {
		return "", nil, cerr
	}

	cp, err = containers.Read(containerPath)

	if err != nil {
		return containerPath, nil, errwrap.Wrapf("Inspection failure on container: {{err}}", err)
	}

	return containerPath, cp, nil
}

// InspectContainer on a given directory, filtering by format
func InspectContainer(format, directory string) (string, error) {
	var containerPath, container, cerr = getContainerPackage(directory)

	switch {
	case os.IsNotExist(cerr):
		return "", errwrap.Wrapf("Inspection failure: can not find container", cerr)
	case cerr != nil:
		return "", cerr
	}

	verbose.Debug("Reading container at " + containerPath)
	return templates.ExecuteOrList(format, container)
}

func getProjectRootDirectory(dir string) (string, error) {
	return getRootDirectory(dir, "project.json")
}

func getContainerRootDirectory(dir string) (string, error) {
	return getRootDirectory(dir, "container.json")
}

func getRootDirectory(dir, file string) (string, error) {
	return findresource.GetRootDirectory(dir, findresource.GetSysRoot(), file)
}
