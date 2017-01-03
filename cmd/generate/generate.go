package cmdgenerate

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
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
	"github.com/wedeploy/cli/verbose"
)

var (
	projectCustomDomain string
	containerType       string
)

var generateRunner = runner{}

// GenerateCmd generates a project or container
var GenerateCmd = &cobra.Command{
	Use:   "generate <host>",
	Short: "Generates a project or container",
	Long: `Use "we generate" to generate projects and containers.
You can generate a project anywhere on your machine and on the cloud.
Containers can only be generated from inside projects and
are stored on the first subdirectory of its project.

--directory should point to either the parent dir of a project directory to be
generated or to a existing project directory.`,
	PreRunE: generateRunner.PreRun,
	RunE:    generateRunner.Run,
	Example: `we generate projector.cinema.wedeploy.io
we generate --project cinema --container projector room`,
}

var setupHost = cmdflagsfromhost.SetupHost{
	Pattern: cmdflagsfromhost.ProjectAndContainerPattern,
}

const (
	customDomainForProjectMessage = "Custom domain for project"
	containerTypeMessage          = "Container type"
)

func init() {
	if !terminal.IsTerminal(int(os.Stdin.Fd())) {
		setupHost.Requires.Project = true
	}

	setupHost.Init(GenerateCmd)
	generateRunner.cmd = GenerateCmd

	GenerateCmd.Flags().StringVar(
		&generateRunner.base,
		"directory",
		"",
		"Overrides current working directory")

	GenerateCmd.Flags().StringVar(
		&projectCustomDomain,
		"project-custom-domain",
		"",
		customDomainForProjectMessage)

	GenerateCmd.Flags().StringVar(
		&containerType,
		"container-type",
		"",
		containerTypeMessage)

	GenerateCmd.Flags().BoolVar(
		&generateRunner.boilerplate,
		"container-boilerplate",
		true,
		"Generate container boilerplate")
}

func shouldPromptToGenerateContainer() (bool, error) {
	if !terminal.IsTerminal(int(os.Stdin.Fd())) {
		return false, errors.New("Project is required when detached from terminal")
	}

	fmt.Println("Generate:")
	fmt.Println("1) a project")
	fmt.Println("2) a project and a container inside it")

	index, err := prompt.SelectOption(2, nil)

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

func (r *runner) checkNoContainerFlagsWhenContainerIsNotGenerated() error {
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
	base              string
	project           string
	projectBase       string
	container         string
	askWithPrompt     bool
	generateContainer bool
	boilerplate       bool
	cmd               *cobra.Command
	baseIsProject     bool
	flagsErr          error
}

func (r *runner) setBase() (err error) {
	if err = r.setBaseDirectory(); err != nil {
		return err
	}

	r.baseIsProject, err = exists(filepath.Join(r.base, "project.json"))
	return err
}

func (r *runner) setBaseDirectory() (err error) {
	if r.base, err = filepath.Abs(r.base); err != nil {
		return errwrap.Wrapf("Can not get absolute path: {{err}}", err)
	}

	e, err := exists(r.base)

	if err == nil && !e {
		return fmt.Errorf("Directory not exists: %v", r.base)
	}

	return err
}

func (r *runner) setup() error {
	if r.flagsErr == nil {
		r.setupProject()
		return nil
	}

	if r.baseIsProject {
		return r.setupContainerOnProject()
	}

	return errwrap.Wrapf("{{err}} unless on a project directory", r.flagsErr)
}

func (r *runner) setupProject() {
	r.project = setupHost.Project()
	r.container = setupHost.Container()

	if r.project == "" {
		r.askWithPrompt = true
	}
}

func (r *runner) setupContainerOnProject() error {
	var ec error
	r.container, ec = r.cmd.Flags().GetString("container")

	if ec != nil {
		return errwrap.Wrapf("Can not get container generated within project: {{err}}", ec)
	}

	return nil
}

func (r *runner) PreRun(cmd *cobra.Command, args []string) (err error) {
	r.flagsErr = setupHost.Process(args)
	return nil
}

func (r *runner) selectProject() (err error) {
	switch r.baseIsProject {
	case true:
		r.projectBase = r.base
		if err = r.handleProjectBase(); err != nil {
			return err
		}
	default:
		if err = r.handleGenerateWhatPrompts(); err != nil {
			return err
		}

		if err = r.handleProject(); err != nil {
			return err
		}
	}

	return nil
}

func (r *runner) Run(cmd *cobra.Command, args []string) (err error) {
	if err = r.setBase(); err != nil {
		return err
	}

	if err = r.setup(); err != nil {
		return err
	}

	if err = r.selectProject(); err != nil {
		return err
	}

	if r.generateContainer {
		return r.handleGenerateContainer()
	}

	return nil
}

func (r *runner) handleProjectBase() error {
	if r.project != "" {
		return fmt.Errorf("Can not use project flag (value: \"%v\") from inside a project:\n%v",
			color.Format(color.FgBlue, r.project), r.base)
	}

	if err := checkNoProjectFlagsWhenProjectAlreadyExists(r.cmd); err != nil {
		return err
	}

	r.generateContainer = true
	return nil
}

func (r *runner) handleGenerateWhatPrompts() (err error) {
	if r.container != "" {
		r.generateContainer = true
	} else if r.project == "" && r.container == "" {
		if r.generateContainer, err = shouldPromptToGenerateContainer(); err != nil {
			return err
		}
	}

	if !r.generateContainer {
		if err := r.checkNoContainerFlagsWhenContainerIsNotGenerated(); err != nil {
			return err
		}
	}

	if r.project == "" {
		if r.project, err = promptProject(); err != nil {
			return err
		}
	}

	r.projectBase = filepath.Join(r.base, r.project)
	return nil
}

func (r *runner) handleProject() error {
	projectExists, err := exists(filepath.Join(r.projectBase, "project.json"))

	if err != nil {
		return err
	}

	if projectExists {
		if !r.generateContainer {
			return fmt.Errorf("Project %v already exists in:\n%v",
				color.Format(color.FgBlue, r.project), r.projectBase)
		}

		fmt.Fprintf(os.Stderr, "Jumping creation of project %v (already exists)\n",
			color.Format(color.FgBlue, r.project))

		return checkNoProjectFlagsWhenProjectAlreadyExists(r.cmd)
	}

	return r.newProject()
}

func (r *runner) handleGenerateContainer() error {
	if r.container != "" {
		if err := checkContainerDirectory(r.container,
			filepath.Join(r.projectBase, r.container)); err != nil {
			return err
		}
	}

	var cc = &containerCreator{
		ProjectDirectory: r.projectBase,
		Container: &containers.Container{
			ID: r.container,
		},
		boilerplate:            r.boilerplate,
		boilerplateFlagChanged: r.cmd.Flags().Changed("container-boilerplate"),
	}

	return cc.run()
}

type containerCreator struct {
	Container              *containers.Container
	Registry               []containers.Register
	Register               containers.Register
	ProjectDirectory       string
	ContainerDirectory     string
	boilerplate            bool
	boilerplateFlagChanged bool
	boilerplateGenerated   bool
}

func (cc *containerCreator) run() error {
	if err := cc.handleContainerType(); err != nil {
		return err
	}

	if err := cc.chooseContainerID(); err != nil {
		return err
	}

	cc.ContainerDirectory = filepath.Join(cc.ProjectDirectory, cc.Container.ID)

	if err := cc.handleBoilerplate(); err != nil {
		return err
	}

	// 1. mkdir repo; 2. git clone u@h:/p or git scheme://404 repo =>
	// On error git clone actually removes the existing directory. Odd. Weird.
	if !cc.boilerplateGenerated {
		if err := tryGenerateDirectory(cc.ContainerDirectory); err != nil {
			return err
		}
	}

	return cc.saveContainer()
}

func (cc *containerCreator) handleContainerType() error {
	registry, err := containers.GetRegistry(context.Background())

	if err != nil {
		return errwrap.Wrapf("Can not get the registry: {{err}}", err)
	}

	cc.Registry = registry

	if containerType == "" {
		return cc.chooseContainerType()
	}

	for _, r := range cc.Registry {
		if containerType == r.Type {
			cc.Container.Type = r.Type
			return nil
		}
	}

	// if matching for the exact type is not possible, try to find it
	// by getting only possible matches from WeDeploy, without versions
	for _, r := range cc.Registry {
		if containerType == getBoilerplateContainerType(r.Type) {
			cc.Container.Type = r.Type
			return nil
		}
	}

	return errors.New("Container type not found on register")
}

func (cc *containerCreator) chooseContainerType() error {
	fmt.Println(containerTypeMessage + ":")

	var mapSelectOptions = map[string]int{}

	for pos, r := range cc.Registry {
		mapSelectOptions[r.ID] = pos + 1
		mapSelectOptions[strings.TrimPrefix(r.ID, "wedeploy-")] = pos + 1
		ne := fmt.Sprintf("%d) %v", pos+1, r.ID)

		p := 80 - len(ne) - len(r.Type) + 1

		if p < 1 {
			p = 1
		}

		fmt.Fprintf(os.Stdout, "%v%v%v\n", ne, pad(p), r.Type)
		fmt.Fprintf(os.Stdout, "%v\n\n", color.Format(color.FgHiBlack, wordwrap.WrapString(r.Description, 80)))
	}

	option, err := prompt.SelectOption(len(cc.Registry), mapSelectOptions)

	if err != nil {
		return err
	}

	cc.Register = cc.Registry[option]
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

func getBoilerplateContainerType(cType string) string {
	cType = strings.TrimPrefix(cType, "wedeploy/")

	if strings.Contains(cType, ":") {
		ws := strings.SplitN(cType, ":", 2)

		if len(ws) > 1 {
			cType = ws[0]
		}
	}

	return cType
}

func (cc *containerCreator) checkIfDirectoryEmptyForInstallingBoilerplate() (empty bool, err error) {
	notEmpty, err := exists(cc.ContainerDirectory)

	if err != nil {
		return false, err
	}

	if notEmpty {
		if !cc.boilerplateFlagChanged {
			fmt.Fprintf(os.Stderr,
				"Container directory already exists (bypassing installing boilerplate).\n")
			return false, nil
		}

		return false, errors.New("Container directory already exists. Can not install boilerplate.")
	}

	return true, nil
}

func (cc *containerCreator) handleBoilerplate() (err error) {
	if !cc.boilerplate {
		return nil
	}

	if empty, err := cc.checkIfDirectoryEmptyForInstallingBoilerplate(); err != nil || !empty {
		return err
	}

	var (
		container = cc.Container.ID
		cType     = cc.Container.Type
	)

	var boilerplateType = getBoilerplateContainerType(cType)
	var boilerplateAddress = fmt.Sprintf(
		"https://github.com/wedeploy/boilerplate-%v.git",
		boilerplateType)

	// There isn't a way to simply curl | unzip here.
	// the separate-git-dir is used just as a safeguard against removing
	// legit .git in case of an atypical failure.
	// os.Remove is used to remove .git
	// if it is a symlink to the separate set git-dir, it is going to remove
	// if not, it is going to fail.
	// This is a dirty / quick approach and should be replaced with a proper
	// unzipping of a release whenever a reliable and tested solution, such as
	// a library or a new UnzipPackage zip/archive method, is readily available.
	// This would, at most, cause destruction of a link, instead of the data
	// it points to, in the worst case.

	var boilerplateDotGit = filepath.Join(cc.ContainerDirectory, ".boilerplate-git")

	cmd := exec.Command("git",
		"clone",
		"--depth=1",
		"--quiet",
		"--separate-git-dir="+boilerplateDotGit,
		boilerplateAddress,
		cc.ContainerDirectory)

	var cmdStderr = new(bytes.Buffer)

	if verbose.Enabled {
		cmd.Stderr = io.MultiWriter(cmdStderr, os.Stderr)
	} else {
		cmd.Stderr = cmdStderr
	}

	if err = cmd.Run(); err != nil {
		if !cc.boilerplateFlagChanged {
			fmt.Fprintf(os.Stderr, "Jumping boilerplate creation (not available).\n")
			return nil
		}

		fmt.Fprintf(os.Stderr, "%v\n", cmdStderr.String())
		return errwrap.Wrapf("Can not get boilerplate: {{err}}", err)
	}

	cc.boilerplateGenerated = true

	if err = os.RemoveAll(boilerplateDotGit); err != nil {
		return errwrap.Wrapf("Can not remove boilerplate .git hidden dir: {{err}}", err)
	}

	// never use os.RemoveAll here, see block comment on generateBoilerplate()
	if err = os.Remove(filepath.Join(cc.ContainerDirectory, ".git")); err != nil {
		return errwrap.Wrapf("Error removing .git ref file for container's boilerplate: {{err}}", err)
	}

	if cc.Container, err = containers.Read(cc.ContainerDirectory); err != nil {
		return errwrap.Wrapf("Can not read boilerplate's container file: {{err}}", err)
	}

	cc.Container.ID = container
	cc.Container.Type = cType
	return nil
}

func (cc *containerCreator) saveContainer() error {
	bin, err := json.MarshalIndent(cc.Container, "", "    ")

	if err != nil {
		return err
	}

	err = ioutil.WriteFile(filepath.Join(cc.ContainerDirectory, "container.json"),
		bin, 0644)

	if err == nil {
		abs, aerr := filepath.Abs(cc.ContainerDirectory)

		if aerr != nil {
			fmt.Fprintf(os.Stderr, "Error getting absolute path: %v\n", aerr)
		}

		fmt.Fprintf(os.Stdout, `Container generated at %v
Go to the container directory to keep hacking! :)
Some tips:
	Run the container on your local machine with: "we link"
	Check container.json for additional configuration.
`, abs)
	}

	return err
}

func (r *runner) newProject() (err error) {
	if r.askWithPrompt {
		if projectCustomDomain, err = prompt.Prompt(customDomainForProjectMessage); err != nil {
			return err
		}
	}

	var p = &projects.Project{
		ID: r.project,
	}

	if projectCustomDomain != "" {
		p.CustomDomains = []string{projectCustomDomain}
	}

	return r.saveProject(p)
}

func (r *runner) saveProject(p *projects.Project) error {
	if err := tryGenerateDirectory(r.projectBase); err != nil {
		return err
	}

	bin, err := json.MarshalIndent(p, "", "    ")

	if err != nil {
		return err
	}

	err = ioutil.WriteFile(filepath.Join(r.projectBase, "project.json"), bin, 0644)

	if err == nil {
		abs, err := filepath.Abs(r.projectBase)

		if err != nil {
			return err
		}

		fmt.Fprintf(os.Stdout, `Project generated at %v
Go to the project directory and happy hacking! :)
Some tips:
	Run the project on your local machine with: "we link"
	Generate a container there with: "we generate"
	Check project.json for additional configuration.
`, abs)
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

func tryGenerateDirectory(directory string) error {
	var err = os.MkdirAll(directory, 0775)

	if err != nil {
		return errwrap.Wrapf("Can not generate directory: {{err}}", err)
	}

	return err
}

func pad(space int) string {
	return strings.Join(make([]string, space), " ")
}
