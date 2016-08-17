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
func New(id, directory string) error {
	if directory == "" {
		directory = id
	}

	// On Windows probably something like C:\ and D:\ can't be related
	// so don't check for err here
	var rel, _ = filepath.Rel(directory, config.Context.ProjectRoot)

	if rel == "." && config.Context.Scope == "container" {
		return getErrContainerAlreadyExists(config.Context.ContainerRoot)
	}

	if rel == "." && config.Context.Scope == "project" {
		return getErrProjectAlreadyExists(config.Context.ProjectRoot)
	}

	switch config.Context.Scope {
	case "project":
		return newContainer(id, directory)
	case "global":
		return newProject(id, directory)
	default:
		return getErrContainerAlreadyExists(config.Context.ContainerRoot)
	}
}

type containerCreator struct {
	Container *containers.Container
	Register  containers.Register
	Directory string
}

func newContainer(id, directory string) error {
	if config.Context.Scope == "container" {
		return getErrContainerAlreadyExists(config.Context.ContainerRoot)
	}

	if config.Context.Scope != "project" {
		return getErrProjectAlreadyExists(config.Context.ProjectRoot)
	}

	return (&containerCreator{
		Directory: directory,
		Container: &containers.Container{
			ID: id,
		},
	}).run()
}

func (cc *containerCreator) run() error {
	var rel string

	if cc.Directory != "" {
		if err := tryContainerDirectory(cc.Directory); err != nil {
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

	if err = cc.getContainersRegister(); err != nil {
		return err
	}

	if err = cc.chooseContainerOptions(); err != nil {
		return err
	}

	return cc.saveContainer()
}

func (cc *containerCreator) getContainersRegister() error {
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
	index, err := strconv.Atoi(option)

	index--

	if err != nil || index < 0 || index > len(registry) {
		return errors.New("Invalid option.")
	}

	cc.Register = registry[index]
	return err
}

func (cc *containerCreator) chooseContainerOptions() error {
	if cc.Container.ID == "" {
		cc.Container.ID = prompt.Prompt("Id [default: " + cc.Register.ID + "]")
	}

	if cc.Container.ID == "" {
		cc.Container.ID = cc.Register.ID
	}

	if cc.Directory == "" {
		cc.Directory = cc.Container.ID
	}

	var err = tryContainerDirectory(cc.Directory)

	if err != nil {
		return err
	}

	cc.Container.Name = prompt.Prompt("Name [default: " + cc.Register.Name + "]")

	if cc.Container.Name == "" {
		cc.Container.Name = cc.Register.Name
	}

	cc.Container.Type = cc.Register.Type

	return err
}

func (cc *containerCreator) saveContainer() error {
	var bin, err = json.MarshalIndent(cc.Container, "", "    ")

	if err != nil {
		return err
	}

	err = ioutil.WriteFile(filepath.Join(cc.Directory, "container.json"),
		bin, 0644)

	if err == nil {
		abs, ea := filepath.Abs(filepath.Join(cc.Directory))

		if ea != nil {
			panic(ea)
		}

		fmt.Println("Container created at " + abs)
	}

	return err
}

func newProject(id, directory string) error {
	if config.Context.Scope != "global" {
		return ErrProjectPath
	}

	var p = &projects.Project{}

	p.ID = id

	if p.ID == "" {
		fmt.Println("Creating project:")
		p.ID = prompt.Prompt("ID")
	} else {
		fmt.Println("Creating project: " + p.ID)
	}

	if p.ID == "" {
		return ErrInvalidID
	}

	if directory == "" {
		directory = p.ID
	}

	if err := tryProjectDirectory(directory); err != nil {
		return err
	}

	p.Name = prompt.Prompt("Name")
	p.CustomDomain = prompt.Prompt("Custom domain")

	return saveProject(p, directory)
}

func saveProject(p *projects.Project, directory string) error {
	var bin, err = json.MarshalIndent(p, "", "    ")

	if err != nil {
		return err
	}

	err = ioutil.WriteFile(filepath.Join(directory, "project.json"), bin, 0644)

	if err == nil {
		abs, ea := filepath.Abs(directory)

		if ea != nil {
			panic(ea)
		}

		fmt.Println("Project created at " + abs)
	}

	return err
}

func tryProjectDirectory(directory string) error {
	var err = os.MkdirAll(directory, 0775)

	if err != nil {
		return errwrap.Wrapf("Can't create project directory: {{err}}", err)
	}

	_, err = os.Stat(filepath.Join(directory, "project.json"))

	abs, eabs := filepath.Abs(directory)

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

func tryContainerDirectory(directory string) error {
	var err = os.MkdirAll(directory, 0775)

	if err != nil {
		return errwrap.Wrapf("Can't create container directory: {{err}}", err)
	}

	_, err = os.Stat(filepath.Join(directory, "container.json"))

	abs, eabs := filepath.Abs(directory)

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
