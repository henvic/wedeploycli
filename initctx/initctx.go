package initctx

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/launchpad-project/cli/config"
	"github.com/launchpad-project/cli/containers"
	"github.com/launchpad-project/cli/projects"
	"github.com/launchpad-project/cli/prompt"
)

var (
	// ErrContainerPath indicates an invalid container location
	ErrContainerPath = errors.New("A container immediate parent dir must be the root of a project")

	// ErrProjectPath indicates an invalid project location
	ErrProjectPath = errors.New("A project can not have another project as its parent")

	// ErrInvalidID indicates an invalid resource ID (such as empty string)
	ErrInvalidID = errors.New("Value for resource ID is invalid")

	// ErrResourceExists indicates that two resouce can not share the same location
	ErrResourceExists = errors.New("A resouce already exists on the root of this location")
)

// New creates a resource
func New() error {
	switch config.Context.Scope {
	case "project":
		return NewContainer()
	case "global":
		return NewProject()
	default:
		return ErrResourceExists
	}
}

// NewContainer creates a container resource
func NewContainer() error {
	var rel string
	var bin []byte

	if config.Context.Scope == "container" {
		return ErrResourceExists
	}

	if config.Context.Scope != "project" {
		return ErrContainerPath
	}

	projectRoot := config.Context.ProjectRoot
	workingDir, err := os.Getwd()

	if err == nil {
		rel, err = filepath.Rel(projectRoot, workingDir)
	}

	if err != nil {
		return err
	}

	// only allow container creation at first subdir level
	if strings.ContainsRune(rel, os.PathSeparator) {
		return ErrContainerPath
	}

	var c = &containers.Container{}

	if rel == "." {
		return ErrResourceExists
	}

	fmt.Println("Creating container")
	c.ID = prompt.Prompt("ID")

	if c.ID == "" {
		return ErrInvalidID
	}

	c.Name = prompt.Prompt("Name")

	bin, err = json.MarshalIndent(c, "", "    ")

	if err != nil {
		return err
	}

	err = ioutil.WriteFile(
		filepath.Join(workingDir, "container.json"),
		bin,
		0644)

	return err
}

// NewProject creates a project resource
func NewProject() error {
	var bin []byte

	if config.Context.Scope != "global" {
		return ErrProjectPath
	}

	workingDir, err := os.Getwd()

	if err != nil {
		return err
	}

	var p = &projects.Project{}

	fmt.Println("Creating project")
	p.ID = prompt.Prompt("ID")

	if p.ID == "" {
		return ErrInvalidID
	}

	p.Name = prompt.Prompt("Name")

	bin, err = json.MarshalIndent(p, "", "    ")

	if err != nil {
		return err
	}

	err = ioutil.WriteFile(
		filepath.Join(workingDir, "project.json"),
		bin,
		0644)

	return err
}
