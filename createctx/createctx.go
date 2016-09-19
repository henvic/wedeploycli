package createctx

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/hashicorp/errwrap"
	wordwrap "github.com/mitchellh/go-wordwrap"
	"github.com/wedeploy/cli/color"
	"github.com/wedeploy/cli/containers"
	"github.com/wedeploy/cli/projects"
	"github.com/wedeploy/cli/prompt"
)

var (
	// ErrContainerPath indicates an invalid container location
	ErrContainerPath = errors.New("A container immediate parent dir must be the root of a project")
)

type containerCreator struct {
	Container *containers.Container
	Register  containers.Register
	Directory string
}

// NewContainer creates a new container directory
func NewContainer(container, project, directory string) error {
	if err := tryCreateDirectory(directory); err != nil {
		return err
	}

	return (&containerCreator{
		Directory: directory,
		Container: &containers.Container{
			ID: container,
		},
	}).run()
}

func (cc *containerCreator) checkParentDirIsProject() error {
	switch _, err := projects.Read(filepath.Join(cc.Directory, "..")); err {
	case nil:
	case projects.ErrProjectNotFound:
		return errwrap.Wrapf("Parent directory is not a project", err)
	default:
		return errwrap.Wrapf("Error trying to find project on parent dir: {{err}}", err)
	}

	return nil
}

func (cc *containerCreator) run() error {
	if err := cc.checkParentDirIsProject(); err != nil {
		return err
	}

	if err := cc.getContainersRegister(); err != nil {
		return err
	}

	if err := cc.chooseContainerOptions(); err != nil {
		return err
	}

	return cc.saveContainer()
}

func (cc *containerCreator) getContainersRegister() error {
	registry, err := containers.GetRegistry(context.Background())

	if err != nil {
		return errwrap.Wrapf("Can't get the registry: {{err}}", err)
	}

	for pos, r := range registry {
		ne := fmt.Sprintf("%d) %v", pos+1, r.ID)

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

	var err = tryCreateDirectory(cc.Directory)

	if err != nil {
		return err
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

// NewProject creates a new project directory
func NewProject(project, directory string) error {
	var p = &projects.Project{}

	p.ID = project

	if p.ID == "" {
		fmt.Println("Creating project:")
		p.ID = prompt.Prompt("ID")
	} else {
		fmt.Println("Creating project: " + p.ID)
	}

	if p.ID == "" {
		return errors.New("Empty value for project ID is invalid")
	}

	if directory == "" {
		directory = p.ID
	}

	if err := tryCreateDirectory(directory); err != nil {
		return err
	}

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

func checkDirectoryIsFree(directory string) error {
	switch _, err := os.Stat(filepath.Join(directory, "project.json")); {
	case os.IsNotExist(err):
	case err == nil:
		return errors.New("Project already exists in " + directory)
	default:
		return errwrap.Wrapf("Error trying to read project on "+
			directory+": {{err}}", err)
	}

	switch _, err := os.Stat(filepath.Join(directory, "container.json")); {
	case os.IsNotExist(err):
		return nil
	case err == nil:
		return errors.New("Container already exists in " + directory)
	default:
		return errwrap.Wrapf("Error trying to read container on "+
			directory+": {{err}}", err)
	}
}

func tryCreateDirectory(directory string) error {
	if err := checkDirectoryIsFree(directory); err != nil {
		return err
	}

	var err = os.MkdirAll(directory, 0775)

	if err != nil {
		return errwrap.Wrapf("Can't create directory: {{err}}", err)
	}

	return err
}

func pad(space int) string {
	return strings.Join(make([]string, space), " ")
}
