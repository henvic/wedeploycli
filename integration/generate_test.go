package integration

import (
	"reflect"
	"strings"
	"testing"

	"github.com/wedeploy/cli/containers"
	"github.com/wedeploy/cli/projects"
)

func TestGenerateDirectoryNotExists(t *testing.T) {
	var cmd = &Command{
		Args: []string{"generate", "--project", "foo", "--directory", "not-found"},
	}

	cmd.Run()

	if cmd.Stdout.Len() != 0 {
		t.Errorf("Expected stdout to be empty, got %v instead", cmd.Stdout)
	}

	var wantErr = "Directory not exists"

	if !strings.Contains(cmd.Stderr.String(), wantErr) {
		t.Errorf("Wanted stderr to have %v, got %v instead", wantErr, cmd.Stderr)
	}

	if cmd.ExitCode == 0 {
		t.Errorf("Expected exit code to be not 0, got %v instead", cmd.ExitCode)
	}
}

func TestGenerateErrorContainerTypeOnContainerOnly(t *testing.T) {
	var cmd = &Command{
		Args: []string{"generate", "--project", "foo", "--container-type", "auth"},
	}

	cmd.Run()

	if cmd.Stdout.Len() != 0 {
		t.Errorf("Expected stdout to be empty, got %v instead", cmd.Stdout)
	}

	var wantErr = "Flag --container is required by --container-type"

	if !strings.Contains(cmd.Stderr.String(), wantErr) {
		t.Errorf("Wanted stderr to have %v, got %v instead", wantErr, cmd.Stderr)
	}

	if cmd.ExitCode == 0 {
		t.Errorf("Expected exit code to be not 0, got %v instead", cmd.ExitCode)
	}
}

func TestGenerateProjectAlreadyExists(t *testing.T) {
	var cmd = &Command{
		Args: []string{"generate", "--project", "foo", "--directory", "mocks/generate", "--no-color"},
	}

	cmd.Run()

	if cmd.Stdout.Len() != 0 {
		t.Errorf("Expected stdout to be empty, got %v instead", cmd.Stdout)
	}

	var wantErr = "Project foo already exists in"

	if !strings.Contains(cmd.Stderr.String(), wantErr) {
		t.Errorf("Wanted stderr to have %v, got %v instead", wantErr, cmd.Stderr)
	}

	if cmd.ExitCode == 0 {
		t.Errorf("Expected exit code to be not 0, got %v instead", cmd.ExitCode)
	}
}

func TestGenerateProjectAlreadyExistsAttribute(t *testing.T) {
	var cmd = &Command{
		Args: []string{"generate",
			"--project",
			"foo",
			"--directory",
			"mocks/generate",
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
Project flags (--project-custom-domain) can only be used on new projects`

	if !strings.Contains(cmd.Stderr.String(), wantErr) {
		t.Errorf("Wanted stderr to have %v, got %v instead", wantErr, cmd.Stderr)
	}

	if cmd.ExitCode == 0 {
		t.Errorf("Expected exit code to be not 0, got %v instead", cmd.ExitCode)
	}
}

func TestGenerateProjectAlreadyExistsInsideBase(t *testing.T) {
	var cmd = &Command{
		Args: []string{"generate",
			"--project",
			"foo",
			"--directory",
			"mocks/generate/foo",
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

	var wantErr = `Can not use project flag (value: "foo") from inside a project`

	if !strings.Contains(cmd.Stderr.String(), wantErr) {
		t.Errorf("Wanted stderr to have %v, got %v instead", wantErr, cmd.Stderr)
	}

	if cmd.ExitCode == 0 {
		t.Errorf("Expected exit code to be not 0, got %v instead", cmd.ExitCode)
	}
}

func TestGenerateProject(t *testing.T) {
	removeAll("mocks/generate/example")
	defer removeAll("mocks/generate/example")

	var cmd = &Command{
		Args: []string{"generate", "--project", "example", "--directory", "mocks/generate"},
	}

	cmd.Run()

	var want = "Project generated at"
	if !strings.Contains(cmd.Stdout.String(), want) {
		t.Errorf("Wanted stdout to have %v, got %v instead", want, cmd.Stdout)
	}

	if cmd.Stderr.Len() != 0 {
		t.Errorf("Expected stderr to be empty, got %v instead", cmd.Stderr)
	}

	if cmd.ExitCode != 0 {
		t.Errorf("Expected exit code to be 0, got %v instead", cmd.ExitCode)
	}

	var p, err = projects.Read("mocks/generate/example")

	if err != nil {
		t.Errorf("Expected reading project file, got error %v instead", err)
	}

	if p.ID != "example" {
		t.Errorf(`Expected project to be generated with ID "example" got %v instead`, p.ID)
	}

	if len(p.CustomDomains) != 0 {
		t.Errorf("Expected no custom domain for project, got %v instead", p.CustomDomains)
	}
}

func TestGenerateProjectWithCustomDomain(t *testing.T) {
	removeAll("mocks/generate/example")
	defer removeAll("mocks/generate/example")

	var cmd = &Command{
		Args: []string{"generate",
			"--project",
			"example",
			"--project-custom-domain",
			"example.com",
			"--directory",
			"mocks/generate",
		},
	}

	cmd.Run()

	var want = "Project generated at"
	if !strings.Contains(cmd.Stdout.String(), want) {
		t.Errorf("Wanted stdout to have %v, got %v instead", want, cmd.Stdout)
	}

	if cmd.Stderr.Len() != 0 {
		t.Errorf("Expected stderr to be empty, got %v instead", cmd.Stderr)
	}

	if cmd.ExitCode != 0 {
		t.Errorf("Expected exit code to be 0, got %v instead", cmd.ExitCode)
	}

	var p, err = projects.Read("mocks/generate/example")

	if err != nil {
		t.Errorf("Expected reading project file, got error %v instead", err)
	}

	if p.ID != "example" {
		t.Errorf(`Expected project to be generated with ID "example" got %v instead`, p.ID)
	}

	var wantCustomDomains = []string{"example.com"}

	if !reflect.DeepEqual(p.CustomDomains, wantCustomDomains) {
		t.Errorf("Expected custom domain for project to be %v, got %v instead", wantCustomDomains, p.CustomDomains[0])
	}
}

func TestGenerateProjectWithCustomDomainAndContainerWithoutContainerBoilerplate(t *testing.T) {
	removeAll("mocks/generate/example")
	defer removeAll("mocks/generate/example")

	var cmd = &Command{
		Args: []string{"generate",
			"--project",
			"example",
			"--container",
			"email",
			"--project-custom-domain",
			"example.com",
			"--container-type",
			"auth",
			"--container-boilerplate=false",
			"--directory",
			"mocks/generate",
		},
	}

	cmd.Run()

	var wantProject = "Project generated at"
	if !strings.Contains(cmd.Stdout.String(), wantProject) {
		t.Errorf("Wanted stdout to have %v, got %v instead", wantProject, cmd.Stdout)
	}

	var wantContainer = "Container generated at"
	if !strings.Contains(cmd.Stdout.String(), wantContainer) {
		t.Errorf("Wanted stdout to have %v, got %v instead", wantContainer, cmd.Stdout)
	}

	if cmd.Stderr.Len() != 0 {
		t.Errorf("Expected stderr to be empty, got %v instead", cmd.Stderr)
	}

	if cmd.ExitCode != 0 {
		t.Errorf("Expected exit code to be 0, got %v instead", cmd.ExitCode)
	}

	var p, err = projects.Read("mocks/generate/example")

	if err != nil {
		t.Errorf("Expected reading project file, got error %v instead", err)
	}

	if p.ID != "example" {
		t.Errorf(`Expected project to be generated with ID "example" got %v instead`, p.ID)
	}

	var wantCustomDomains = []string{"example.com"}

	if !reflect.DeepEqual(p.CustomDomains, wantCustomDomains) {
		t.Errorf("Expected custom domain for project to be %v, got %v instead", wantCustomDomains, p.CustomDomains[0])
	}
}

func TestGenerateContainerInsideAlreadyExistingProjectWithoutContainerBoilerplate(t *testing.T) {
	removeAll("mocks/generate/foo/data")
	defer removeAll("mocks/generate/foo/data")

	var cmd = &Command{
		Args: []string{"generate",
			"--container",
			"data",
			"--container-type",
			"data",
			"--container-boilerplate=false",
			"--directory",
			"mocks/generate/foo",
		},
	}

	cmd.Run()

	var dontWantProject = "Project generated at"
	if strings.Contains(cmd.Stdout.String(), dontWantProject) {
		t.Errorf("Wanted stdout to not have %v, got %v instead", dontWantProject, cmd.Stdout)
	}

	var wantContainer = "Container generated at"
	if !strings.Contains(cmd.Stdout.String(), wantContainer) {
		t.Errorf("Wanted stdout to have %v, got %v instead", wantContainer, cmd.Stdout)
	}

	if cmd.Stderr.Len() != 0 {
		t.Errorf("Expected stderr to be empty, got %v instead", cmd.Stderr)
	}

	if cmd.ExitCode != 0 {
		t.Errorf("Expected exit code to be 0, got %v instead", cmd.ExitCode)
	}

	var cp, err = containers.Read("mocks/generate/foo/data")

	if err != nil {
		t.Errorf("Expected reading container file, got error %v instead", err)
	}

	if cp.ID != "data" {
		t.Errorf(`Expected container to be generated with ID "data" got %v instead`, cp.ID)
	}
}

func TestGenerateContainerWithoutProjectError(t *testing.T) {
	removeAll("mocks/generate/example")
	defer removeAll("mocks/generate/example")

	var cmd = &Command{
		Args: []string{"generate", "--container", "foo", "--directory", "mocks/generate"},
	}

	cmd.Run()

	var wantErr = "Incompatible use: --container requires --project unless on a project directory"

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
