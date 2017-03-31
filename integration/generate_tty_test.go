// +build !windows

package integration

import (
	"os"
	"regexp"
	"testing"
	"time"

	"github.com/henvic/pseudoterm"
	"github.com/wedeploy/cli/containers"
	"github.com/wedeploy/cli/projects"
)

func TestGeneratePromptProjectThenContainer(t *testing.T) {
	Setup()
	defer Teardown()
	removeAll("mocks/generate/example")
	defer removeAll("mocks/generate/example")

	if ok := t.Run("testGeneratePromptProject", testGeneratePromptProject); ok {
		t.Run("testGeneratePromptContainer", testGeneratePromptContainer)
	}
}

func TestGeneratePromptProjectAndContainerAtOnce(t *testing.T) {
	Setup()
	defer Teardown()
	removeAll("mocks/generate/example")
	defer removeAll("mocks/generate/example")

	var cmd = (&Command{
		Args: []string{"generate"},
		Env:  []string{"WEDEPLOY_CUSTOM_HOME=" + GetLoginHome()},
		Dir:  "mocks/generate",
	}).Prepare()

	var term = &pseudoterm.Terminal{
		Command: cmd,
	}

	if testing.Verbose() {
		term.EchoStream = os.Stdout
	}

	var story = &pseudoterm.QueueStory{
		Timeout: 30 * time.Second,
	}

	story.Add(
		pseudoterm.Step{
			Read:      "Generate: ",
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
		t.Errorf("we generate did not execute successfully")
	}

	project, err := projects.Read("mocks/generate/example")

	if err != nil {
		t.Errorf("Wanted err to be nil, got %v instead", err)
	}

	var wantProject = "example"

	if project.ID != wantProject {
		t.Errorf("Expected project ID to be %v, got %v instead", wantProject, project.ID)
	}

	cp, err := containers.Read("mocks/generate/example/auth")

	var wantContainer = "auth"

	if cp.ID != wantContainer {
		t.Errorf("Expected container ID to be %v, got %v instead", wantContainer, cp.ID)
	}

	if err != nil {
		t.Errorf("Wanted err to be nil, got %v instead", err)
	}
}

func testGeneratePromptProject(t *testing.T) {
	var cmd = (&Command{
		Args: []string{"generate"},
		Env:  []string{"WEDEPLOY_CUSTOM_HOME=" + GetLoginHome()},
		Dir:  "mocks/generate",
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
			Read:      "Generate: ",
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
		t.Errorf("we generate did not execute successfully")
	}

	project, err := projects.Read("mocks/generate/example")

	if err != nil {
		t.Errorf("Wanted err to be nil, got %v instead", err)
	}

	var wantProject = "example"

	if project.ID != wantProject {
		t.Errorf("Expected project ID to be %v, got %v instead", wantProject, project.ID)
	}
}

func testGeneratePromptContainer(t *testing.T) {
	var cmd = (&Command{
		Args: []string{"generate"},
		Env:  []string{"WEDEPLOY_CUSTOM_HOME=" + GetLoginHome()},
		Dir:  "mocks/generate/example",
	}).Prepare()

	var term = &pseudoterm.Terminal{
		Command: cmd,
	}

	if testing.Verbose() {
		term.EchoStream = os.Stdout
	}

	var story = &pseudoterm.QueueStory{
		Timeout: 30 * time.Second,
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
		t.Errorf("we generate did not execute successfully")
	}

	cp, err := containers.Read("mocks/generate/example/auth")

	var wantContainer = "auth"

	if cp.ID != wantContainer {
		t.Errorf("Expected container ID to be %v, got %v instead", wantContainer, cp.ID)
	}

	if err != nil {
		t.Errorf("Wanted err to be nil, got %v instead", err)
	}

	if _, err := os.Stat("mocks/generate/example/auth/README.md"); err != nil {
		t.Errorf("Expected boilerplate file to exist, got %v instead", err)
	}
}
