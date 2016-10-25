package integration

import (
	"os"
	"reflect"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/henvic/pseudoterm"
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

	var wantErr = `Error: Can't use project flag (value: "foo") from inside a project`

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

func TestCreatePromptProjectAndContainerAtOnce(t *testing.T) {
	removeAll("mocks/create/example")
	defer removeAll("mocks/create/example")

	var cmd = (&Command{
		Args: []string{"create"},
		Dir:  "mocks/create",
	}).Prepare()

	var term = &pseudoterm.Terminal{
		Command: cmd,
	}

	if testing.Verbose() {
		term.EchoStream = os.Stdout
	}

	var story = &pseudoterm.QueueStory{
		Timeout: 20 * time.Second,
	}

	story.Add(
		pseudoterm.Step{
			Read:      "Create: ",
			SkipWrite: true,
		},
		pseudoterm.Step{
			Read:      "2) a project and a container inside it",
			SkipWrite: true,
		},
		pseudoterm.Step{
			Read:  "Select from 1..2:",
			Write: "2",
		},
		pseudoterm.Step{
			Read:  "Project:",
			Write: "example",
		},
		pseudoterm.Step{
			Read:  "Custom domain for project:",
			Write: "",
		},
		pseudoterm.Step{
			Read:      "Container type:",
			SkipWrite: true,
		},
		pseudoterm.Step{
			ReadRegex: regexp.MustCompile("Select from 1..([0-9]+):"),
			Write:     "auth",
		},
		pseudoterm.Step{
			Read:  "Container ID [default: wedeploy-auth]:",
			Write: "auth",
		},
		pseudoterm.Step{
			Read:      "Go to the container directory to keep hacking! :)",
			SkipWrite: true,
		},
	)

	if err := term.Run(story); err != nil {
		t.Fatalf("Unexpected error %v", err)
	}

	if !term.Wait().Success() {
		t.Errorf("we create did not execute successfully")
	}

	project, err := projects.Read("mocks/create/example")

	if err != nil {
		t.Errorf("Wanted err to be nil, got %v instead", err)
	}

	var wantProject = "example"

	if project.ID != wantProject {
		t.Errorf("Expected project ID to be %v, got %v instead", wantProject, project.ID)
	}

	container, err := containers.Read("mocks/create/example/auth")

	var wantContainer = "auth"

	if container.ID != wantContainer {
		t.Errorf("Expected container ID to be %v, got %v instead", wantContainer, container.ID)
	}

	if err != nil {
		t.Errorf("Wanted err to be nil, got %v instead", err)
	}
}

func testCreatePromptProject(t *testing.T) {
	var cmd = (&Command{
		Args: []string{"create"},
		Dir:  "mocks/create",
	}).Prepare()

	var term = &pseudoterm.Terminal{
		Command: cmd,
	}

	if testing.Verbose() {
		term.EchoStream = os.Stdout
	}

	var story = &pseudoterm.QueueStory{
		Timeout: 3 * time.Second,
	}

	story.Add(
		pseudoterm.Step{
			Read:      "Create: ",
			SkipWrite: true,
		},
		pseudoterm.Step{
			Read:      "2) a project and a container inside it",
			SkipWrite: true,
		},
		pseudoterm.Step{
			Read:  "Select from 1..2:",
			Write: "1",
		},
		pseudoterm.Step{
			Read:  "Project:",
			Write: "example",
		},
		pseudoterm.Step{
			Read:  "Custom domain for project:",
			Write: "",
		},
	)

	if err := term.Run(story); err != nil {
		t.Fatalf("Unexpected error %v", err)
	}

	if !term.Wait().Success() {
		t.Errorf("we create did not execute successfully")
	}

	project, err := projects.Read("mocks/create/example")

	if err != nil {
		t.Errorf("Wanted err to be nil, got %v instead", err)
	}

	var wantProject = "example"

	if project.ID != wantProject {
		t.Errorf("Expected project ID to be %v, got %v instead", wantProject, project.ID)
	}
}

func testCreatePromptContainer(t *testing.T) {
	var cmd = (&Command{
		Args: []string{"create"},
		Dir:  "mocks/create/example",
	}).Prepare()

	var term = &pseudoterm.Terminal{
		Command: cmd,
	}

	if testing.Verbose() {
		term.EchoStream = os.Stdout
	}

	var story = &pseudoterm.QueueStory{
		Timeout: 20 * time.Second,
	}

	story.Add(
		pseudoterm.Step{
			Read:      "Container type:",
			SkipWrite: true,
		},
		pseudoterm.Step{
			Read:  "Select from 1..9:",
			Write: "1",
		},
		pseudoterm.Step{
			Read:  "Container ID [default: wedeploy-auth]:",
			Write: "auth",
		},
	)

	if err := term.Run(story); err != nil {
		t.Fatalf("Unexpected error %v", err)
	}

	if !term.Wait().Success() {
		t.Errorf("we create did not execute successfully")
	}

	container, err := containers.Read("mocks/create/example/auth")

	var wantContainer = "auth"

	if container.ID != wantContainer {
		t.Errorf("Expected container ID to be %v, got %v instead", wantContainer, container.ID)
	}

	if err != nil {
		t.Errorf("Wanted err to be nil, got %v instead", err)
	}

	if _, err := os.Stat("mocks/create/example/auth/README.md"); err != nil {
		t.Errorf("Expected boilerplate file to exist, got %v instead", err)
	}
}

func TestCreatePromptProjectThenContainer(t *testing.T) {
	removeAll("mocks/create/example")
	defer removeAll("mocks/create/example")

	if ok := t.Run("testCreatePromptProject", testCreatePromptProject); ok {
		t.Run("testCreatePromptContainer", testCreatePromptContainer)
	}
}
