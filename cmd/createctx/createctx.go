package cmdcreate

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/crypto/ssh/terminal"

	"github.com/hashicorp/errwrap"
	wordwrap "github.com/mitchellh/go-wordwrap"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/wedeploy/cli/cmdflagsfromhost"
	"github.com/wedeploy/cli/color"
	"github.com/wedeploy/cli/containers"
	"github.com/wedeploy/cli/projects"
	"github.com/wedeploy/cli/prompt"
)

var (
	projectCustomDomain string
	containerType       string
)

var createRunner = runner{}

// CreateCmd creates a project or container
var CreateCmd = &cobra.Command{
	Use:   "create <host>",
	Short: "Creates a project or container",
	Long: `Use "we create" to create projects and containers.
You can create a project anywhere on your machine and on the cloud.
Containers can only be created from inside projects and
are stored on the first subdirectory of its project.`,
	PreRunE: preRun,
	RunE:    createRunner.Run,
	Example: `we create projector.cinema.wedeploy.io
we create --project cinema --container projector room`,
}

var setupHost = cmdflagsfromhost.SetupHost{
	Pattern: cmdflagsfromhost.FullHostPattern,
}

const (
	customDomainForProjectMessage = "Custom domain for project"
	containerTypeMessage          = "Container type"
)

func init() {
	if !terminal.IsTerminal(int(os.Stdin.Fd())) {
		setupHost.Requires.Project = true
	}

	setupHost.Init(CreateCmd)
	createRunner.cmd = CreateCmd

	CreateCmd.Flags().StringVar(
		&createRunner.base,
		"directory",
		"",
		"Overrides current working directory")

	CreateCmd.Flags().StringVar(
		&projectCustomDomain,
		"project-custom-domain",
		"",
		customDomainForProjectMessage)

	CreateCmd.Flags().StringVar(
		&containerType,
		"container-type",
		"",
		containerTypeMessage)
}

func preRun(cmd *cobra.Command, args []string) error {
	return setupHost.Process(args)
}

func shouldPromptToCreateContainer() (bool, error) {
	if !terminal.IsTerminal(int(os.Stdin.Fd())) {
		return false, errors.New("Project is required when detached from terminal")
	}

	fmt.Println("Do you want to create:")
	fmt.Println("1) only project")
	fmt.Println("2) container (or project and container)")

	index, err := prompt.SelectOption(2)

	if err != nil {
		return false, err
	}

	const offset = 1
	return index != 1-offset, nil
}

func promptProject() (project string, err error) {
	project, err = prompt.Prompt("Project")

	if err != nil {
		return "", err
	}

	if project == "" {
		return "", errors.New("Project is required")
	}

	return project, nil
}

func checkContainerDirectory(container, path string) error {
	switch containerExists, err := exists(filepath.Join(path, "container.json")); {
	case containerExists:
		return fmt.Errorf("Container %v already exists in:\n%v",
			color.Format(color.FgBlue, container), path)
	default:
		return err
	}
}

func getUsedFlagsPrefixList(cmd *cobra.Command, prefix string) (list []string) {
	cmd.Flags().Visit(func(f *pflag.Flag) {
		if !strings.HasPrefix(f.Name, prefix+"-") {
			return
		}

		if f.Changed {
			list = append(list, "--"+f.Name)
		}
	})

	return list
}

func (r *runner) checkNoContainerFlagsWhenContainerIsNotCreated() error {
	var list = getUsedFlagsPrefixList(r.cmd, "container")

	if len(list) == 0 {
		return nil
	}

	return fmt.Errorf("%v: flags requires --container directive", strings.Join(list, ", "))
}

func checkNoProjectFlagsWhenProjectAlreadyExists(cmd *cobra.Command) error {
	var list = getUsedFlagsPrefixList(cmd, "project")

	if len(list) == 0 {
		return nil
	}

	return fmt.Errorf("%v: flags used when project already exists", strings.Join(list, ", "))
}

type runner struct {
	base            string
	project         string
	container       string
	askWithPrompt   bool
	createContainer bool
	cmd             *cobra.Command
}

func (r *runner) setBase() (err error) {
	if r.base, err = filepath.Abs(r.base); err != nil {
		return errwrap.Wrapf("Can't get absolute path: {{err}}", err)
	}

	e, err := exists(r.base)

	if err == nil && !e {
		return fmt.Errorf("Directory not exists: %v", r.base)
	}

	return err
}

func (r *runner) setup() {
	r.project = setupHost.Project()
	r.container = setupHost.Container()

	if r.project == "" {
		r.askWithPrompt = true
	}
}

func (r *runner) Run(cmd *cobra.Command, args []string) (err error) {
	if err = r.setBase(); err != nil {
		return err
	}

	r.setup()

	if err = r.handlePrompts(); err != nil {
		return err
	}

	if err = r.handleProject(); err != nil {
		return err
	}

	if r.createContainer {
		fmt.Println("")
		return r.handleCreateContainer()
	}

	return nil
}

func (r *runner) handlePrompts() (err error) {
	if r.container != "" {
		r.createContainer = true
	} else if r.project == "" && r.container == "" {
		if r.createContainer, err = shouldPromptToCreateContainer(); err != nil {
			return err
		}
	}

	if !r.createContainer {
		if err := r.checkNoContainerFlagsWhenContainerIsNotCreated(); err != nil {
			return err
		}
	}

	if r.project == "" {
		if r.project, err = promptProject(); err != nil {
			return err
		}
	}

	return nil
}

func (r *runner) handleProject() error {
	projectExists, err := exists(filepath.Join(r.base, r.project, "project.json"))

	if err != nil {
		return err
	}

	if projectExists {
		if !r.createContainer {
			return fmt.Errorf("Project %v already exists in:\n%v",
				color.Format(color.FgBlue, r.project), filepath.Join(r.base, r.project))
		}

		fmt.Fprintf(os.Stderr, "Jumping creating project %v (already exists)\n",
			color.Format(color.FgBlue, r.project))

		return checkNoProjectFlagsWhenProjectAlreadyExists(r.cmd)
	}

	return r.newProject()
}

func (r *runner) handleCreateContainer() error {
	if r.container == "" {
		r.askWithPrompt = true
	} else if err := checkContainerDirectory(
		r.container,
		filepath.Join(r.base, r.project, r.container)); err != nil {
		return err
	}

	var cc = &containerCreator{
		ProjectDirectory: filepath.Join(r.base, r.project),
		Container: &containers.Container{
			ID: r.container,
		},
	}

	return cc.run()
}

type containerCreator struct {
	Container        *containers.Container
	Register         containers.Register
	ProjectDirectory string
}

func (cc *containerCreator) run() error {
	if err := cc.chooseContainerType(); err != nil {
		return err
	}

	if err := cc.chooseContainerID(); err != nil {
		return err
	}

	return cc.saveContainer()
}

func (cc *containerCreator) chooseContainerType() error {
	fmt.Println(containerTypeMessage + ":")
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

	option, err := prompt.SelectOption(len(registry))

	if err != nil {
		return err
	}

	cc.Register = registry[option]
	cc.Container.Type = cc.Register.Type
	return nil
}

func (cc *containerCreator) chooseContainerID() (err error) {
	var container = cc.Container.ID

	if container == "" {
		container, err = prompt.Prompt("Container ID [default: " + cc.Register.ID + "]")

		if err != nil {
			return err
		}

		if container == "" {
			container = cc.Register.ID
		}

		err := checkContainerDirectory(container,
			filepath.Join(cc.ProjectDirectory, container))

		if err != nil {
			return err
		}
	}

	cc.Container.ID = container
	return nil
}

func (cc *containerCreator) saveContainer() error {
	var containerDirectory = filepath.Join(cc.ProjectDirectory, cc.Container.ID)

	if err := tryCreateDirectory(containerDirectory); err != nil {
		return err
	}

	bin, err := json.MarshalIndent(cc.Container, "", "    ")

	if err != nil {
		return err
	}

	err = ioutil.WriteFile(filepath.Join(containerDirectory, "container.json"),
		bin, 0644)

	if err == nil {
		abs, err := filepath.Abs(filepath.Join(containerDirectory))

		if err != nil {
			return err
		}

		fmt.Println("Container created at " + abs)
	}

	return err
}

func (r *runner) newProject() (err error) {
	if r.askWithPrompt {
		if projectCustomDomain, err = prompt.Prompt(customDomainForProjectMessage); err != nil {
			return err
		}
	}

	var directory = filepath.Join(r.base, r.project)
	var p = &projects.Project{
		ID:           r.project,
		CustomDomain: projectCustomDomain,
	}

	p.ID = r.project

	return r.saveProject(p, directory)
}

func (r *runner) saveProject(p *projects.Project, directory string) error {
	if err := tryCreateDirectory(directory); err != nil {
		return err
	}

	bin, err := json.MarshalIndent(p, "", "    ")

	if err != nil {
		return err
	}

	err = ioutil.WriteFile(filepath.Join(directory, "project.json"), bin, 0644)

	if err == nil {
		abs, err := filepath.Abs(directory)

		if err != nil {
			return err
		}

		fmt.Println("Project created at " + abs)
	}

	return err
}

func exists(file string) (bool, error) {
	switch _, err := os.Stat(file); {
	case os.IsNotExist(err):
		return false, nil
	case err == nil:
		return true, nil
	default:
		return false, errwrap.Wrapf("Error trying to read "+
			file+": {{err}}", err)
	}
}

func tryCreateDirectory(directory string) error {
	var err = os.MkdirAll(directory, 0775)

	if err != nil {
		return errwrap.Wrapf("Can't create directory: {{err}}", err)
	}

	return err
}

func pad(space int) string {
	return strings.Join(make([]string, space), " ")
}
