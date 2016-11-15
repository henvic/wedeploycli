package inspector

import (
	"fmt"
	"io"
	"os"
	"reflect"

	"github.com/hashicorp/errwrap"
	"github.com/wedeploy/cli/containers"
	"github.com/wedeploy/cli/findresource"
	"github.com/wedeploy/cli/projects"
	"github.com/wedeploy/cli/templates"
	"github.com/wedeploy/cli/verbose"
)

var outStream io.Writer = os.Stdout

func printTypeFieldNames(t interface{}) {
	val := reflect.ValueOf(t)
	for i := 0; i < val.NumField(); i++ {
		field := val.Type().Field(i)
		fmt.Fprintf(outStream, "%v %v\n", field.Name, field.Type.String())
	}
}

// ContextOverview for the context visualization
type ContextOverview struct {
	Scope         string
	ProjectRoot   string
	ContainerRoot string
	ProjectID     string
	ContainerID   string
}

// PrintContextSpec for the ContextOverview type
func PrintContextSpec() {
	printTypeFieldNames(ContextOverview{})
}

func (overview *ContextOverview) loadProject(directory string) error {
	var projectPath, project, perr = getProject(directory)

	switch {
	case os.IsNotExist(perr):
	case perr != nil:
		return errwrap.Wrapf("Can't load project context on "+projectPath+": {{err}}", perr)
	default:
		overview.Scope = "project"
		overview.ProjectRoot = projectPath
		overview.ProjectID = project.ID
	}

	return nil
}

func (overview *ContextOverview) loadContainer(directory string) error {
	var containerPath, container, cerr = getContainer(directory)

	switch {
	case os.IsNotExist(cerr):
	case cerr != nil:
		return errwrap.Wrapf("Can't load container context on "+containerPath+": {{err}}", cerr)
	default:
		overview.Scope = "container"
		overview.ContainerRoot = containerPath
		overview.ContainerID = container.ID
	}

	return nil
}

// Load the context overview for a given directory
func (overview *ContextOverview) Load(directory string) error {
	overview.Scope = "Global"

	if err := overview.loadProject(directory); err != nil {
		return err
	}

	if err := overview.loadContainer(directory); err != nil {
		return err
	}

	return nil
}

// InspectContext on a given directory, filtering by format
func InspectContext(format, directory string) error {
	// Can't rely on values on usercontext.Context given that
	// they are global and we accept the directory parameter
	var overview = ContextOverview{}

	if err := overview.Load(directory); err != nil {
		return err
	}

	var content, eerr = templates.ExecuteOrList(format, overview)

	if eerr != nil {
		return eerr
	}

	fmt.Fprintf(outStream, "%v\n", content)
	return nil
}

// PrintProjectSpec for the Project type
func PrintProjectSpec() {
	printTypeFieldNames(projects.Project{})
}

// InspectProject on a given directory, filtering by format
func InspectProject(format, directory string) error {
	var projectPath, project, perr = getProject(directory)

	switch {
	case os.IsNotExist(perr):
		return errwrap.Wrapf("Inspection failure: can't find project", perr)
	case perr != nil:
		return perr
	}

	verbose.Debug("Reading project at " + projectPath)

	var content, eerr = templates.ExecuteOrList(format, project)

	if eerr != nil {
		return eerr
	}

	fmt.Fprintf(outStream, "%v\n", content)
	return nil
}

func getProject(directory string) (path string, project *projects.Project, err error) {
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

func getContainer(directory string) (path string, container *containers.Container, err error) {
	var containerPath, cerr = getContainerRootDirectory(directory)

	if cerr != nil {
		return "", nil, cerr
	}

	container, err = containers.Read(containerPath)

	if err != nil {
		return "", nil, errwrap.Wrapf("Inspection failure on container: {{err}}", err)
	}

	return containerPath, container, nil
}

// PrintContainerSpec for the Container type
func PrintContainerSpec() {
	printTypeFieldNames(containers.Container{})
}

// InspectContainer on a given directory, filtering by format
func InspectContainer(format, directory string) error {
	var containerPath, container, cerr = getContainer(directory)

	switch {
	case os.IsNotExist(cerr):
		return errwrap.Wrapf("Inspection failure: can't find container", cerr)
	case cerr != nil:
		return cerr
	}

	verbose.Debug("Reading container at " + containerPath)

	var content, eerr = templates.ExecuteOrList(format, container)

	if eerr != nil {
		return eerr
	}

	fmt.Fprintf(outStream, "%v\n", content)
	return nil
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
