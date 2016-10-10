package integration

import (
	"strings"
	"testing"

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
}

func TestCreateProjectWithCustomDomain(t *testing.T) {
	removeAll("mocks/create/example")
	defer removeAll("mocks/create/example")

	var cmd = &Command{
		Args: []string{"create",
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

	var wantCustomDomain = "example.com"

	if p.CustomDomain != wantCustomDomain {
		t.Errorf("Expected custom domain for project to bem %v, got %v instead", wantCustomDomain, p.CustomDomain)
	}
}

func TestCreateProjectWithCustomDomainAndContainerWithoutBoilerplate(t *testing.T) {
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

	var wantCustomDomain = "example.com"

	if p.CustomDomain != wantCustomDomain {
		t.Errorf("Expected custom domain for project to bem %v, got %v instead", wantCustomDomain, p.CustomDomain)
	}
}
