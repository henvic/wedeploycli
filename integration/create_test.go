package integration

import (
	"reflect"
	"strings"
	"testing"

	"github.com/wedeploy/cli/containers"
	"github.com/wedeploy/cli/projects"
)

func TestCreateIncompatibleUse(t *testing.T) {
	var cmd = &Command{
		Args: []string{"create", "--project", "foo", "mocks"},
	}

	cmd.Run()

	if cmd.Stdout.Len() != 0 {
		t.Errorf("Expected stdout to be empty, got %v instead", cmd.Stdout)
	}

	var wantErr = "Error: Incompatible use: --project and --container are not allowed with host format"

	if !strings.Contains(cmd.Stderr.String(), wantErr) {
		t.Errorf("Wanted stderr to have %v, got %v instead", wantErr, cmd.Stderr)
	}

	if cmd.ExitCode == 0 {
		t.Errorf("Expected exit code to be not 0, got %v instead", cmd.ExitCode)
	}
}

func TestCreateDirectoryNotExists(t *testing.T) {
	var cmd = &Command{
		Args: []string{"create", "--project", "foo", "--directory", "not-found"},
	}

	cmd.Run()

	if cmd.Stdout.Len() != 0 {
		t.Errorf("Expected stdout to be empty, got %v instead", cmd.Stdout)
	}

	var wantErr = "Error: Directory not exists"

	if !strings.Contains(cmd.Stderr.String(), wantErr) {
		t.Errorf("Wanted stderr to have %v, got %v instead", wantErr, cmd.Stderr)
	}

	if cmd.ExitCode == 0 {
		t.Errorf("Expected exit code to be not 0, got %v instead", cmd.ExitCode)
	}
}

func TestCreateErrorContainerTypeOnContainerOnly(t *testing.T) {
	var cmd = &Command{
		Args: []string{"create", "--project", "foo", "--container-type", "auth"},
	}

	cmd.Run()

	if cmd.Stdout.Len() != 0 {
		t.Errorf("Expected stdout to be empty, got %v instead", cmd.Stdout)
	}

	var wantErr = "Error: --container-type: flags requires --container directive"

	if !strings.Contains(cmd.Stderr.String(), wantErr) {
		t.Errorf("Wanted stderr to have %v, got %v instead", wantErr, cmd.Stderr)
	}

	if cmd.ExitCode == 0 {
		t.Errorf("Expected exit code to be not 0, got %v instead", cmd.ExitCode)
	}
}

func TestCreateProjectAlreadyExists(t *testing.T) {
	var cmd = &Command{
		Args: []string{"create", "--project", "foo", "--directory", "mocks/create", "--no-color"},
	}

	cmd.Run()

	if cmd.Stdout.Len() != 0 {
		t.Errorf("Expected stdout to be empty, got %v instead", cmd.Stdout)
	}

	var wantErr = "Error: Project foo already exists in"

	if !strings.Contains(cmd.Stderr.String(), wantErr) {
		t.Errorf("Wanted stderr to have %v, got %v instead", wantErr, cmd.Stderr)
	}

	if cmd.ExitCode == 0 {
		t.Errorf("Expected exit code to be not 0, got %v instead", cmd.ExitCode)
	}
}

func TestCreateProjectAlreadyExistsAttribute(t *testing.T) {
	var cmd = &Command{
		Args: []string{"create",
			"--project",
			"foo",
			"--directory",
			"mocks/create",
			"--project-custom-domain",
			"example.com",
			"--container",
			"bar",
			"--no-color"},
	}

	cmd.Run()

	if cmd.Stdout.Len() != 0 {
		t.Errorf("Expected stdout to be empty, got %v instead", cmd.Stdout)
	}

	var wantErr = `Jumping creation of project foo (already exists)
Error: --project-custom-domain: flags used when project already exists`

	if !strings.Contains(cmd.Stderr.String(), wantErr) {
		t.Errorf("Wanted stderr to have %v, got %v instead", wantErr, cmd.Stderr)
	}

	if cmd.ExitCode == 0 {
		t.Errorf("Expected exit code to be not 0, got %v instead", cmd.ExitCode)
	}
}

func TestCreateProjectAlreadyExistsInsideBase(t *testing.T) {
	var cmd = &Command{
		Args: []string{"create",
			"--project",
			"foo",
			"--directory",
			"mocks/create/foo",
			"--project-custom-domain",
			"example.com",
			"--container",
			"bar",
			"--no-color"},
	}

	cmd.Run()

	if cmd.Stdout.Len() != 0 {
		t.Errorf("Expected stdout to be empty, got %v instead", cmd.Stdout)
	}

	var wantErr = `Error: Can not use project flag (value: "foo") from inside a project`

	if !strings.Contains(cmd.Stderr.String(), wantErr) {
		t.Errorf("Wanted stderr to have %v, got %v instead", wantErr, cmd.Stderr)
	}

	if cmd.ExitCode == 0 {
		t.Errorf("Expected exit code to be not 0, got %v instead", cmd.ExitCode)
	}
}

func TestCreateProject(t *testing.T) {
	removeAll("mocks/create/example")
	defer removeAll("mocks/create/example")

	var cmd = &Command{
		Args: []string{"create", "--project", "example", "--directory", "mocks/create"},
	}

	cmd.Run()

	var want = "Project created at"
	if !strings.Contains(cmd.Stdout.String(), want) {
		t.Errorf("Wanted stdout to have %v, got %v instead", want, cmd.Stdout)
	}

	if cmd.Stderr.Len() != 0 {
		t.Errorf("Expected stderr to be empty, got %v instead", cmd.Stderr)
	}

	if cmd.ExitCode != 0 {
		t.Errorf("Expected exit code to be 0, got %v instead", cmd.ExitCode)
	}

	var p, err = projects.Read("mocks/create/example")

	if err != nil {
		t.Errorf("Expected reading project file, got error %v instead", err)
	}

	if p.ID != "example" {
		t.Errorf(`Expected project to be created with ID "example" got %v instead`, p.ID)
	}

	if len(p.CustomDomains) != 0 {
		t.Errorf("Expected no custom domain for project, got %v instead", p.CustomDomains)
	}
}

func TestCreateProjectWithCustomDomain(t *testing.T) {
	removeAll("mocks/create/example")
	defer removeAll("mocks/create/example")

	var cmd = &Command{
		Args: []string{"create",
			"--project",
			"example",
			"--project-custom-domain",
			"example.com",
			"--directory",
			"mocks/create",
		},
	}

	cmd.Run()

	var want = "Project created at"
	if !strings.Contains(cmd.Stdout.String(), want) {
		t.Errorf("Wanted stdout to have %v, got %v instead", want, cmd.Stdout)
	}

	if cmd.Stderr.Len() != 0 {
		t.Errorf("Expected stderr to be empty, got %v instead", cmd.Stderr)
	}

	if cmd.ExitCode != 0 {
		t.Errorf("Expected exit code to be 0, got %v instead", cmd.ExitCode)
	}

	var p, err = projects.Read("mocks/create/example")

	if err != nil {
		t.Errorf("Expected reading project file, got error %v instead", err)
	}

	if p.ID != "example" {
		t.Errorf(`Expected project to be created with ID "example" got %v instead`, p.ID)
	}

	var wantCustomDomains = []string{"example.com"}

	if !reflect.DeepEqual(p.CustomDomains, wantCustomDomains) {
		t.Errorf("Expected custom domain for project to be %v, got %v instead", wantCustomDomains, p.CustomDomains[0])
	}
}

func TestCreateProjectWithCustomDomainAndContainerWithoutContainerBoilerplate(t *testing.T) {
	removeAll("mocks/create/example")
	defer removeAll("mocks/create/example")

	var cmd = &Command{
		Args: []string{"create",
			"mail.example",
			"--project-custom-domain",
			"example.com",
			"--container-type",
			"auth",
			"--container-boilerplate=false",
			"--directory",
			"mocks/create",
		},
	}

	cmd.Run()

	var wantProject = "Project created at"
	if !strings.Contains(cmd.Stdout.String(), wantProject) {
		t.Errorf("Wanted stdout to have %v, got %v instead", wantProject, cmd.Stdout)
	}

	var wantContainer = "Container created at"
	if !strings.Contains(cmd.Stdout.String(), wantContainer) {
		t.Errorf("Wanted stdout to have %v, got %v instead", wantContainer, cmd.Stdout)
	}

	if cmd.Stderr.Len() != 0 {
		t.Errorf("Expected stderr to be empty, got %v instead", cmd.Stderr)
	}

	if cmd.ExitCode != 0 {
		t.Errorf("Expected exit code to be 0, got %v instead", cmd.ExitCode)
	}

	var p, err = projects.Read("mocks/create/example")

	if err != nil {
		t.Errorf("Expected reading project file, got error %v instead", err)
	}

	if p.ID != "example" {
		t.Errorf(`Expected project to be created with ID "example" got %v instead`, p.ID)
	}

	var wantCustomDomains = []string{"example.com"}

	if !reflect.DeepEqual(p.CustomDomains, wantCustomDomains) {
		t.Errorf("Expected custom domain for project to be %v, got %v instead", wantCustomDomains, p.CustomDomains[0])
	}
}

func TestCreateContainerInsideAlreadyExistingProjectWithoutContainerBoilerplate(t *testing.T) {
	removeAll("mocks/create/foo/data")
	defer removeAll("mocks/create/foo/data")

	var cmd = &Command{
		Args: []string{"create",
			"--container",
			"data",
			"--container-type",
			"data",
			"--container-boilerplate=false",
			"--directory",
			"mocks/create/foo",
		},
	}

	cmd.Run()

	var dontWantProject = "Project created at"
	if strings.Contains(cmd.Stdout.String(), dontWantProject) {
		t.Errorf("Wanted stdout to not have %v, got %v instead", dontWantProject, cmd.Stdout)
	}

	var wantContainer = "Container created at"
	if !strings.Contains(cmd.Stdout.String(), wantContainer) {
		t.Errorf("Wanted stdout to have %v, got %v instead", wantContainer, cmd.Stdout)
	}

	if cmd.Stderr.Len() != 0 {
		t.Errorf("Expected stderr to be empty, got %v instead", cmd.Stderr)
	}

	if cmd.ExitCode != 0 {
		t.Errorf("Expected exit code to be 0, got %v instead", cmd.ExitCode)
	}

	var c, err = containers.Read("mocks/create/foo/data")

	if err != nil {
		t.Errorf("Expected reading container file, got error %v instead", err)
	}

	if c.ID != "data" {
		t.Errorf(`Expected container to be created with ID "data" got %v instead`, c.ID)
	}
}

func TestCreateContainerWithoutProjectError(t *testing.T) {
	removeAll("mocks/create/example")
	defer removeAll("mocks/create/example")

	var cmd = &Command{
		Args: []string{"create", "--container", "foo", "--directory", "mocks/create"},
	}

	cmd.Run()

	var wantErr = "Error: Incompatible use: --container requires --project unless on a project directory"

	if !strings.Contains(cmd.Stderr.String(), wantErr) {
		t.Errorf("Wanted stdout to have %v, got %v instead", wantErr, cmd.Stdout)
	}

	if cmd.Stdout.Len() != 0 {
		t.Errorf("Expected stdout to be empty, got %v instead", cmd.Stdout)
	}

	if cmd.ExitCode == 0 {
		t.Errorf("Expected exit code to be not 0, got %v instead", cmd.ExitCode)
	}
}
