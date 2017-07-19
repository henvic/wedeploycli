package integration

import (
	"strings"
	"testing"

	"github.com/wedeploy/cli/services"
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

func TestGenerateErrorServiceTypeOnServiceOnly(t *testing.T) {
	var cmd = &Command{
		Args: []string{"generate", "--project", "foo", "--service-type", "auth"},
	}

	cmd.Run()

	if cmd.Stdout.Len() != 0 {
		t.Errorf("Expected stdout to be empty, got %v instead", cmd.Stdout)
	}

	var wantErr = "Flag --service is required by --service-type"

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

func TestGenerateProjectAlreadyExistsInsideBase(t *testing.T) {
	var cmd = &Command{
		Args: []string{"generate",
			"--project",
			"foo",
			"--directory",
			"mocks/generate/foo",
			"--service",
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
}

func TestGenerateProjectAndServiceWithoutServiceBoilerplate(t *testing.T) {
	t.Skipf("Registry is changed and the generate command is hidden on releases currently")
	removeAll("mocks/generate/example")
	defer removeAll("mocks/generate/example")

	var cmd = &Command{
		Args: []string{"generate",
			"--project",
			"example",
			"--service",
			"email",
			"--service-type",
			"auth",
			"--service-boilerplate=false",
			"--directory",
			"mocks/generate",
		},
	}

	cmd.Run()

	var wantProject = "Project generated at"
	if !strings.Contains(cmd.Stdout.String(), wantProject) {
		t.Errorf("Wanted stdout to have %v, got %v instead", wantProject, cmd.Stdout)
	}

	var wantService = "Service generated at"
	if !strings.Contains(cmd.Stdout.String(), wantService) {
		t.Errorf("Wanted stdout to have %v, got %v instead", wantService, cmd.Stdout)
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
}

func TestGenerateServiceInsideAlreadyExistingProjectWithoutServiceBoilerplate(t *testing.T) {
	t.Skipf("Registry is changed and the generate command is hidden on releases currently")

	removeAll("mocks/generate/foo/data")
	defer removeAll("mocks/generate/foo/data")

	var cmd = &Command{
		Args: []string{"generate",
			"--service",
			"data",
			"--service-type",
			"data",
			"--service-boilerplate=false",
			"--directory",
			"mocks/generate/foo",
		},
	}

	cmd.Run()

	var dontWantProject = "Project generated at"
	if strings.Contains(cmd.Stdout.String(), dontWantProject) {
		t.Errorf("Wanted stdout to not have %v, got %v instead", dontWantProject, cmd.Stdout)
	}

	var wantService = "Service generated at"
	if !strings.Contains(cmd.Stdout.String(), wantService) {
		t.Errorf("Wanted stdout to have %v, got %v instead", wantService, cmd.Stdout)
	}

	if cmd.Stderr.Len() != 0 {
		t.Errorf("Expected stderr to be empty, got %v instead", cmd.Stderr)
	}

	if cmd.ExitCode != 0 {
		t.Errorf("Expected exit code to be 0, got %v instead", cmd.ExitCode)
	}

	var cp, err = services.Read("mocks/generate/foo/data")

	if err != nil {
		t.Errorf("Expected reading service file, got error %v instead", err)
	}

	if cp.ID != "data" {
		t.Errorf(`Expected service to be generated with ID "data" got %v instead`, cp.ID)
	}
}

func TestGenerateServiceWithoutProjectError(t *testing.T) {
	removeAll("mocks/generate/example")
	defer removeAll("mocks/generate/example")

	var cmd = &Command{
		Args: []string{"generate", "--service", "foo", "--directory", "mocks/generate"},
	}

	cmd.Run()

	var wantErr = "Incompatible use: --service requires --project unless on a project directory"

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
