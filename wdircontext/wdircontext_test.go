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
	if err := config.Setup("mocks/.we"); err != nil {
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
	if err := config.Setup("mocks/.we"); err != nil {
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

func TestGetProjectOrServiceIDInvalidContext(t *testing.T) {
	project, service, err := GetProjectOrServiceID()

	if project != "" {
		t.Errorf("Wanted empty project ID, got %v instead", project)
	}

	if service != "" {
		t.Errorf("Wanted empty service ID, got %v instead", service)
	}

	if err != ErrContextNotFound {
		t.Errorf("Wanted error %v, got %v instead", ErrContextNotFound, err)
	}
}

func TestGetProjectOrServiceID(t *testing.T) {
	var workingDir, _ = os.Getwd()
	chdir("./mocks/project/")
	if err := config.Setup("mocks/.we"); err != nil {
		panic(err)
	}

	project, service, err := GetProjectOrServiceID()
	wantProject := "extraction"

	if project != wantProject {
		t.Errorf("Wanted empty project %v, got %v instead", wantProject, project)
	}

	if service != "" {
		t.Errorf("Wanted empty service ID, got %v instead", service)
	}

	if err != nil {
		t.Errorf("Wanted error %v, got %v instead", nil, err)
	}

	chdir(workingDir)
	config.Teardown()
}

func TestGetProjectOrServiceIDWithService(t *testing.T) {
	var workingDir, _ = os.Getwd()
	chdir("./mocks/project/service")
	if err := config.Setup("mocks/.we"); err != nil {
		panic(err)
	}

	project, service, err := GetProjectOrServiceID()
	wantProject := "extraction"
	wantService := "myservice"

	if project != wantProject {
		t.Errorf("Wanted empty project %v, got %v instead", wantProject, project)
	}

	if service != wantService {
		t.Errorf("Wanted service %v, got %v instead", wantService, service)
	}

	if err != nil {
		t.Errorf("Wanted error %v, got %v instead", nil, err)
	}

	chdir(workingDir)
	config.Teardown()
}

func TestGetProjectOrServiceIDInvalidService(t *testing.T) {
	var workingDir, _ = os.Getwd()
	chdir("./mocks/project/invalid-service")
	if err := config.Setup("mocks/.we"); err != nil {
		panic(err)
	}

	project, service, err := GetProjectOrServiceID()
	wantProject := "extraction"

	if project != wantProject {
		t.Errorf("Wanted empty project %v, got %v instead", wantProject, project)
	}

	if service != "" {
		t.Errorf("Wanted empty service ID, got %v instead", service)
	}

	if err == nil {
		t.Errorf("Wanted error, got %v instead", err)
	}

	chdir(workingDir)
	config.Teardown()
}

func TestGetServiceID(t *testing.T) {
	var workingDir, _ = os.Getwd()
	chdir("./mocks/project/service")
	if err := config.Setup("mocks/.we"); err != nil {
		panic(err)
	}

	service, err := GetServiceID()
	want := "myservice"

	if service != want {
		t.Errorf("Wanted empty service %v, got %v instead", want, service)
	}

	if err != nil {
		t.Errorf("Wanted error %v, got %v instead", nil, err)
	}

	chdir(workingDir)
	config.Teardown()
}

func TestGetServiceIDInvalidContext(t *testing.T) {
	service, err := GetServiceID()

	if service != "" {
		t.Errorf("Wanted empty service ID, got %v instead", service)
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
