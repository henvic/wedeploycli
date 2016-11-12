package inspector

import (
	"errors"
	"fmt"
	"os"
	"reflect"

	"github.com/hashicorp/errwrap"
	"github.com/wedeploy/cli/containers"
	"github.com/wedeploy/cli/findresource"
	"github.com/wedeploy/cli/projects"
	"github.com/wedeploy/cli/templates"
)

func printTypeFieldNames(t interface{}) {
	val := reflect.ValueOf(t)
	for i := 0; i < val.NumField(); i++ {
		field := val.Type().Field(i)
		fmt.Println(field.Name + " " + field.Type.String())
	}
}

// PrintProjectSpec for the Project type
func PrintProjectSpec() {
	printTypeFieldNames(projects.Project{})
}

// InspectProject on a given directory, filtering by format
func InspectProject(format, directory string) error {
	var projectPath, cerr = getProjectPath(directory)

	if cerr != nil {
		return cerr
	}

	var project, err = projects.Read(projectPath)

	if err != nil {
		return errwrap.Wrapf("Inspection failure: {{err}}", err)
	}

	var content, eerr = templates.ExecuteOrList(format, project)

	if eerr != nil {
		return eerr
	}

	fmt.Println(content)
	return nil
}

func getProjectPath(directory string) (string, error) {
	var project, err = getProjectRootDirectory(directory)

	switch {
	case err == nil:
		return project, nil
	case os.IsNotExist(err):
		return "", errors.New("Inspection failure: can't find project.")
	default:
		return "", err
	}
}

func getContainerPath(directory string) (string, error) {
	var container, err = getContainerRootDirectory(directory)

	switch {
	case err == nil:
		return container, nil
	case os.IsNotExist(err):
		return "", errors.New("Inspection failure: can't find container.")
	default:
		return "", err
	}
}

// PrintContainerSpec for the Container type
func PrintContainerSpec() {
	printTypeFieldNames(containers.Container{})
}

// InspectContainer on a given directory, filtering by format
func InspectContainer(format, directory string) error {
	var containerPath, cerr = getContainerPath(directory)

	if cerr != nil {
		return cerr
	}

	var container, err = containers.Read(containerPath)

	if err != nil {
		return errwrap.Wrapf("Inspection failure: {{err}}", err)
	}

	var content, eerr = templates.ExecuteOrList(format, container)

	if eerr != nil {
		return eerr
	}

	fmt.Println(content)
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
