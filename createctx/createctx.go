package createctx

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/hashicorp/errwrap"
	"github.com/mitchellh/go-wordwrap"
	"github.com/wedeploy/cli/color"
	"github.com/wedeploy/cli/config"
	"github.com/wedeploy/cli/containers"
	"github.com/wedeploy/cli/projects"
	"github.com/wedeploy/cli/prompt"
)

var (
	// ErrContainerPath indicates an invalid container location
	ErrContainerPath = errors.New("A container immediate parent dir must be the root of a project")

	// ErrProjectPath indicates an invalid project location
	ErrProjectPath = errors.New("A project can not have another project as its parent")

	// ErrInvalidID indicates an invalid resource ID (such as empty string)
	ErrInvalidID = errors.New("Value for ID is invalid")

	// ErrProjectAlreadyExists indicates that a project already exists
	ErrProjectAlreadyExists = errors.New("Invalid path for new configuration: project already exists")

	// ErrContainerAlreadyExists indicates that a container already exists
	ErrContainerAlreadyExists = errors.New("Invalid path for new configuration: container already exists")
)

// New creates a resource
func New(id string) error {
	switch config.Context.Scope {
	case "project":
		return NewContainer(id)
	case "global":
		return NewProject(id)
	default:
		return getErrContainerAlreadyExists(config.Context.ContainerRoot)
	}
}

// NewContainer creates a container resource
func NewContainer(id string) error {
	var rel string
	var bin []byte

	if config.Context.Scope == "container" {
		return getErrContainerAlreadyExists(config.Context.ContainerRoot)
	}

	if config.Context.Scope != "project" {
		return getErrProjectAlreadyExists(config.Context.ProjectRoot)
	}

	var c = &containers.Container{
		ID: id,
	}

	if c.ID != "" {
		if err := tryContainerID(c.ID); err != nil {
			return err
		}
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

	registry, err := containers.GetRegistry()

	if err != nil {
		return errwrap.Wrapf("Can't get the registry: {{err}}", err)
	}

	for pos, r := range registry {
		ne := fmt.Sprintf("%d) %v", pos+1, r.Name)

		p := 80 - len(ne) - len(r.Type) + 1

		if p < 1 {
			p = 1
		}

		fmt.Fprintf(os.Stdout, "%v%v%v\n", ne, pad(p), r.Type)
		fmt.Fprintf(os.Stdout, "%v\n\n", color.Format(color.FgHiBlack, wordwrap.WrapString(r.Description, 80)))
	}

	var option = prompt.Prompt(fmt.Sprintf("\nSelect container type from 1..%d", len(registry)))

	var index int

	index, err = strconv.Atoi(option)

	index--

	if err != nil || index < 0 || index > len(registry) {
		return errors.New("Invalid option.")
	}

	var reg = registry[index]

	if c.ID == "" {
		c.ID = prompt.Prompt("Id [default: " + reg.ID + "]")
	}

	if c.ID == "" {
		c.ID = reg.ID
	}

	if err := tryContainerID(c.ID); err != nil {
		return err
	}

	c.Name = prompt.Prompt("Name [default: " + reg.Name + "]")

	if c.Name == "" {
		c.Name = reg.Name
	}

	c.Type = reg.Type

	bin, err = json.MarshalIndent(c, "", "    ")

	if err != nil {
		return err
	}

	err = ioutil.WriteFile(filepath.Join(config.Context.ProjectRoot, c.ID, "container.json"), bin, 0644)

	if err == nil {
		abs, ea := filepath.Abs(filepath.Join(config.Context.ProjectRoot, c.ID))

		if ea != nil {
			panic(ea)
		}

		fmt.Println("Project created at " + abs)
	}

	return err
}

// NewProject creates a project resource
func NewProject(id string) error {
	var bin []byte

	if config.Context.Scope != "global" {
		return ErrProjectPath
	}

	var p = &projects.Project{}

	p.ID = id

	if p.ID == "" {
		fmt.Println("Creating project:")
		p.ID = prompt.Prompt("ID")
	}

	if err := tryProjectID(p.ID); err != nil {
		return err
	}

	p.Name = prompt.Prompt("Name")
	p.CustomDomain = prompt.Prompt("Custom domain")

	bin, err := json.MarshalIndent(p, "", "    ")

	if err != nil {
		return err
	}

	err = ioutil.WriteFile(filepath.Join(p.ID, "project.json"), bin, 0644)

	if err == nil {
		abs, ea := filepath.Abs(p.ID)

		if ea != nil {
			panic(ea)
		}

		fmt.Println("Project created at " + abs)
	}

	return err
}

func tryProjectID(ID string) error {
	if ID == "" {
		return ErrInvalidID
	}

	var err = os.MkdirAll(ID, 0775)

	if err != nil {
		return errwrap.Wrapf("Can't create project directory: {{err}}", err)
	}

	_, err = os.Stat(filepath.Join(ID, "project.json"))

	abs, eabs := filepath.Abs(filepath.Join(config.Context.ProjectRoot, ID))

	if eabs != nil {
		fmt.Fprintf(os.Stderr, "%v\n", eabs)
	}

	switch {
	case err == nil:
		return getErrProjectAlreadyExists(abs)
	case os.IsNotExist(err):
		return nil
	default:
		return err
	}
}

func tryContainerID(ID string) error {
	if ID == "" {
		return ErrInvalidID
	}

	var err = os.MkdirAll(filepath.Join(config.Context.ProjectRoot, ID), 0775)

	if err != nil {
		return errwrap.Wrapf("Can't create container directory: {{err}}", err)
	}

	_, err = os.Stat(filepath.Join(config.Context.ProjectRoot, ID, "container.json"))

	abs, eabs := filepath.Abs(filepath.Join(config.Context.ProjectRoot, ID))

	if eabs != nil {
		fmt.Fprintf(os.Stderr, "%v\n", eabs)
	}

	switch {
	case err == nil:
		return getErrContainerAlreadyExists(abs)
	case os.IsNotExist(err):
		return nil
	default:
		return err
	}
}

func pad(space int) string {
	return strings.Join(make([]string, space), " ")
}

func getErrProjectAlreadyExists(directory string) error {
	return errwrap.Wrapf("{{err}} in "+directory, ErrProjectAlreadyExists)
}

func getErrContainerAlreadyExists(directory string) error {
	return errwrap.Wrapf("{{err}} in "+directory, ErrContainerAlreadyExists)
}
