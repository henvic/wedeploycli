package inspector

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"

	"github.com/hashicorp/errwrap"
	"github.com/wedeploy/cli/findresource"
	"github.com/wedeploy/cli/projects"
	"github.com/wedeploy/cli/services"
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
	Scope           usercontext.Scope
	ProjectRoot     string
	ServiceRoot     string
	ProjectID       string
	ServiceID       string
	ProjectServices []services.ServiceInfo
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

	return overview.loadProjectServicesList()
}

func (overview *ContextOverview) loadProjectServicesList() error {
	var list, err = services.GetListFromDirectory(overview.ProjectRoot)

	if err != nil {
		return errwrap.Wrapf("Error while trying to read list of services on project: {{err}}", err)
	}

	for i, l := range list {
		list[i].Location = filepath.Join(overview.ProjectRoot, l.Location)
	}

	overview.ProjectServices = list
	return nil
}

func (overview *ContextOverview) loadService(directory string) error {
	var servicePath, cp, cerr = getServicePackage(directory)

	switch {
	case os.IsNotExist(cerr):
	case cerr != nil:
		return errwrap.Wrapf("Can not load service context on "+servicePath+": {{err}}", cerr)
	default:
		if overview.Scope == usercontext.ProjectScope {
			overview.Scope = usercontext.ServiceScope
		}

		overview.ServiceRoot = servicePath
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

	if err := overview.loadService(directory); err != nil {
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

func getServicePackage(directory string) (path string, cp *services.ServicePackage, err error) {
	var servicePath, cerr = getServiceRootDirectory(directory)

	if cerr != nil {
		return "", nil, cerr
	}

	cp, err = services.Read(servicePath)

	if err != nil {
		return servicePath, nil, errwrap.Wrapf("Inspection failure on service: {{err}}", err)
	}

	return servicePath, cp, nil
}

// InspectService on a given directory, filtering by format
func InspectService(format, directory string) (string, error) {
	var servicePath, service, cerr = getServicePackage(directory)

	switch {
	case os.IsNotExist(cerr):
		return "", errwrap.Wrapf("Inspection failure: can not find service", cerr)
	case cerr != nil:
		return "", cerr
	}

	verbose.Debug("Reading service at " + servicePath)
	return templates.ExecuteOrList(format, service)
}

func getProjectRootDirectory(dir string) (string, error) {
	return getRootDirectory(dir, "project.json")
}

func getServiceRootDirectory(dir string) (string, error) {
	return getRootDirectory(dir, "wedeploy.json")
}

func getRootDirectory(dir, file string) (string, error) {
	return findresource.GetRootDirectory(dir, findresource.GetSysRoot(), file)
}
