package wdircontext

import (
	"os"
	"testing"

	"github.com/wedeploy/cli/config"
	"github.com/wedeploy/cli/projects"
)

func TestGetProjectID(t *testing.T) {
	var workingDir, _ = os.Getwd()
	chdir("./mocks/project/")
	if err := config.Setup(); err != nil {
		panic(err)
	}

	project, err := GetProjectID()
	want := "extraction"

	if project != want {
		t.Errorf("Wanted empty project %v, got %v instead", want, project)
	}

	if err != nil {
		t.Errorf("Wanted error %v, got %v instead", nil, err)
	}

	chdir(workingDir)
	config.Teardown()
}

func TestGetProjectIDInvalidContext(t *testing.T) {
	project, err := GetProjectID()

	if project != "" {
		t.Errorf("Wanted empty project ID, got %v instead", project)
	}

	if err != ErrContextNotFound {
		t.Errorf("Wanted error %v, got %v instead", ErrContextNotFound, err)
	}
}

func TestGetProjectIDWithInvalidProject(t *testing.T) {
	var workingDir, _ = os.Getwd()
	chdir("./mocks/invalid-project/")
	if err := config.Setup(); err != nil {
		panic(err)
	}

	project, err := GetProjectID()

	if project != "" {
		t.Errorf("Wanted empty project ID, got %v instead", project)
	}

	if err != projects.ErrInvalidProjectID {
		t.Errorf("Wanted error %v, got %v instead", projects.ErrInvalidProjectID, err)
	}

	chdir(workingDir)
	config.Teardown()
}

func TestGetProjectOrContainerIDInvalidContext(t *testing.T) {
	project, container, err := GetProjectOrContainerID()

	if project != "" {
		t.Errorf("Wanted empty project ID, got %v instead", project)
	}

	if container != "" {
		t.Errorf("Wanted empty container ID, got %v instead", container)
	}

	if err != ErrContextNotFound {
		t.Errorf("Wanted error %v, got %v instead", ErrContextNotFound, err)
	}
}

func TestGetProjectOrContainerID(t *testing.T) {
	var workingDir, _ = os.Getwd()
	chdir("./mocks/project/")
	if err := config.Setup(); err != nil {
		panic(err)
	}

	project, container, err := GetProjectOrContainerID()
	wantProject := "extraction"

	if project != wantProject {
		t.Errorf("Wanted empty project %v, got %v instead", wantProject, project)
	}

	if container != "" {
		t.Errorf("Wanted empty container ID, got %v instead", container)
	}

	if err != nil {
		t.Errorf("Wanted error %v, got %v instead", nil, err)
	}

	chdir(workingDir)
	config.Teardown()
}

func TestGetProjectOrContainerIDWithContainer(t *testing.T) {
	var workingDir, _ = os.Getwd()
	chdir("./mocks/project/container")
	if err := config.Setup(); err != nil {
		panic(err)
	}

	project, container, err := GetProjectOrContainerID()
	wantProject := "extraction"
	wantContainer := "mycontainer"

	if project != wantProject {
		t.Errorf("Wanted empty project %v, got %v instead", wantProject, project)
	}

	if container != wantContainer {
		t.Errorf("Wanted container %v, got %v instead", wantContainer, container)
	}

	if err != nil {
		t.Errorf("Wanted error %v, got %v instead", nil, err)
	}

	chdir(workingDir)
	config.Teardown()
}

func TestGetProjectOrContainerIDInvalidContainer(t *testing.T) {
	var workingDir, _ = os.Getwd()
	chdir("./mocks/project/invalid-container")
	if err := config.Setup(); err != nil {
		panic(err)
	}

	project, container, err := GetProjectOrContainerID()
	wantProject := "extraction"

	if project != wantProject {
		t.Errorf("Wanted empty project %v, got %v instead", wantProject, project)
	}

	if container != "" {
		t.Errorf("Wanted empty container ID, got %v instead", container)
	}

	if err == nil {
		t.Errorf("Wanted error, got %v instead", err)
	}

	chdir(workingDir)
	config.Teardown()
}

func TestGetContainerID(t *testing.T) {
	var workingDir, _ = os.Getwd()
	chdir("./mocks/project/container")
	if err := config.Setup(); err != nil {
		panic(err)
	}

	container, err := GetContainerID()
	want := "mycontainer"

	if container != want {
		t.Errorf("Wanted empty container %v, got %v instead", want, container)
	}

	if err != nil {
		t.Errorf("Wanted error %v, got %v instead", nil, err)
	}

	chdir(workingDir)
	config.Teardown()
}

func TestGetContainerIDInvalidContext(t *testing.T) {
	container, err := GetContainerID()

	if container != "" {
		t.Errorf("Wanted empty container ID, got %v instead", container)
	}

	if err != ErrContextNotFound {
		t.Errorf("Wanted error %v, got %v instead", ErrContextNotFound, err)
	}
}

func chdir(dir string) {
	if ech := os.Chdir(dir); ech != nil {
		panic(ech)
	}
}
