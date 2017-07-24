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
	"github.com/wedeploy/cli/cmdargslen"
	"github.com/wedeploy/cli/cmdflagsfromhost"
	"github.com/wedeploy/cli/color"
	"github.com/wedeploy/cli/projects"
	"github.com/wedeploy/cli/prompt"
	"github.com/wedeploy/cli/services"
	"github.com/wedeploy/cli/verbose"
)

var (
	projectCustomDomain string
	serviceImage        string
)

var generateRunner = runner{}

// GenerateCmd generates a project or service
var GenerateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generates a project or service",
	Long: `Use "we generate" to generate projects and services.
  You can generate a project anywhere on your machine and on the cloud.
  Services can only be generated from inside projects and are stored on the first subdirectory of its project.

  --directory should point to either the parent dir of a project directory to be generated or to a existing project directory.`,
	PreRunE: generateRunner.PreRun,
	RunE:    generateRunner.Run,
	Example: `we generate --project cinema --service projector room`,
}

var setupHost = cmdflagsfromhost.SetupHost{
	Pattern: cmdflagsfromhost.ProjectAndServicePattern,
	Requires: cmdflagsfromhost.Requires{
		NoHost: true,
		Local:  true,
	},
}

const (
	serviceImageMessage = "Service Image"
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
		&serviceImage,
		"service-image",
		"",
		serviceImageMessage)

	GenerateCmd.Flags().BoolVar(
		&generateRunner.boilerplate,
		"service-boilerplate",
		true,
		"Generate service boilerplate")

	GenerateCmd.Hidden = true
}

func shouldPromptToGenerateService() (bool, error) {
	if !terminal.IsTerminal(int(os.Stdin.Fd())) {
		return false, errors.New("Project is required when detached from terminal")
	}

	fmt.Println("Generate:")
	fmt.Println("1) a project")
	fmt.Print("2) a project and a service inside it\n\n")
	fmt.Println("Select: ")
	index, err := prompt.SelectOption(2, nil)

	if err != nil {
		return false, err
	}

	const offset = 1
	return index != 1-offset, nil
}

func promptProject() (project string, err error) {
	fmt.Println("Project: ")
	project, err = prompt.Prompt()

	if err != nil {
		return "", err
	}

	if project == "" {
		return "", errors.New("Project is required")
	}

	return project, nil
}

func checkServiceDirectory(service, path string) error {
	switch serviceExists, err := exists(filepath.Join(path, "wedeploy.json")); {
	case serviceExists:
		return fmt.Errorf("Service %v already exists in:\n%v",
			color.Format(color.FgBlue, service), path)
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

func (r *runner) checkNoServiceFlagsWhenServiceIsNotGenerated() error {
	var list = getUsedFlagsPrefixList(r.cmd, "service")

	if len(list) == 0 {
		return nil
	}

	return fmt.Errorf("%v: Flag --service is required by --service-image", strings.Join(list, ", "))
}

func checkNoProjectFlagsWhenProjectAlreadyExists(cmd *cobra.Command) error {
	var list = getUsedFlagsPrefixList(cmd, "project")

	if len(list) == 0 {
		return nil
	}

	return fmt.Errorf("Project flags (%v) can only be used on new projects", strings.Join(list, ", "))
}

type runner struct {
	base            string
	project         string
	projectBase     string
	service         string
	askWithPrompt   bool
	generateService bool
	boilerplate     bool
	cmd             *cobra.Command
	baseIsProject   bool
	flagsErr        error
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
		return r.setupServiceOnProject()
	}

	return errwrap.Wrapf("{{err}} unless on a project directory", r.flagsErr)
}

func (r *runner) setupProject() {
	r.project = setupHost.Project()
	r.service = setupHost.Service()

	if r.project == "" {
		r.askWithPrompt = true
	}
}

func (r *runner) setupServiceOnProject() error {
	var ec error
	r.service, ec = r.cmd.Flags().GetString("service")

	if ec != nil {
		return errwrap.Wrapf("Can not get service generated within project: {{err}}", ec)
	}

	return nil
}

func (r *runner) PreRun(cmd *cobra.Command, args []string) (err error) {
	if err := cmdargslen.Validate(args, 0, 0); err != nil {
		return err
	}

	r.flagsErr = setupHost.Process()
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

	if r.generateService {
		return r.handleGenerateService()
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

	r.generateService = true
	return nil
}

func (r *runner) handleGenerateWhatPrompts() (err error) {
	if r.service != "" {
		r.generateService = true
	} else if r.project == "" && r.service == "" {
		if r.generateService, err = shouldPromptToGenerateService(); err != nil {
			return err
		}
	}

	if !r.generateService {
		if err := r.checkNoServiceFlagsWhenServiceIsNotGenerated(); err != nil {
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
		if !r.generateService {
			return fmt.Errorf("Project %v already exists in:\n%v",
				color.Format(color.FgBlue, r.project), r.projectBase)
		}

		fmt.Fprintf(os.Stderr, "Jumping creation of project %v (already exists)\n",
			color.Format(color.FgBlue, r.project))

		return checkNoProjectFlagsWhenProjectAlreadyExists(r.cmd)
	}

	return r.newProject()
}

func (r *runner) handleGenerateService() error {
	if r.service != "" {
		if err := checkServiceDirectory(r.service,
			filepath.Join(r.projectBase, r.service)); err != nil {
			return err
		}
	}

	var cc = &serviceCreator{
		ProjectDirectory: r.projectBase,
		ServicePackage: &services.ServicePackage{
			ID: r.service,
		},
		boilerplate:            r.boilerplate,
		boilerplateFlagChanged: r.cmd.Flags().Changed("service-boilerplate"),
	}

	return cc.run()
}

type serviceCreator struct {
	ServicePackage         *services.ServicePackage
	Registry               []services.Register
	Register               services.Register
	ProjectDirectory       string
	ServiceDirectory       string
	boilerplate            bool
	boilerplateFlagChanged bool
	boilerplateGenerated   bool
}

func (cc *serviceCreator) run() error {
	if err := cc.handleServiceImage(); err != nil {
		return err
	}

	if err := cc.chooseServiceID(); err != nil {
		return err
	}

	cc.ServiceDirectory = filepath.Join(cc.ProjectDirectory, cc.ServicePackage.ID)

	if err := cc.handleBoilerplate(); err != nil {
		return err
	}

	// 1. mkdir repo; 2. git clone u@h:/p or git scheme://404 repo =>
	// On error git clone actually removes the existing directory. Odd. Weird.
	if !cc.boilerplateGenerated {
		if err := tryGenerateDirectory(cc.ServiceDirectory); err != nil {
			return err
		}
	}

	return cc.saveServicePackage()
}

func (cc *serviceCreator) handleServiceImage() error {
	registry, err := services.GetRegistry(context.Background())

	if err != nil {
		return errwrap.Wrapf("Can not get the registry to create service: {{err}}", err)
	}

	cc.Registry = registry

	if serviceImage == "" {
		return cc.chooseServiceType()
	}

	for _, r := range cc.Registry {
		if serviceImage == r.Image {
			cc.ServicePackage.Image = r.Image
			return nil
		}
	}

	// if matching for the exact type is not possible, try to find it
	// by getting only possible matches from WeDeploy, without versions
	for _, r := range cc.Registry {
		if serviceImage == getBoilerplateServiceImage(r.Image) {
			cc.ServicePackage.Image = r.Image
			return nil
		}
	}

	return errors.New("Service image not found on register")
}

func (cc *serviceCreator) chooseServiceType() error {
	fmt.Println(serviceImageMessage + ":")

	var mapSelectOptions = map[string]int{}

	for pos, r := range cc.Registry {
		mapSelectOptions[r.ID] = pos + 1
		mapSelectOptions[strings.TrimPrefix(r.ID, "wedeploy-")] = pos + 1
		ne := fmt.Sprintf("%d) %v", pos+1, r.ID)

		p := 80 - len(ne) - len(r.Image) + 1

		if p < 1 {
			p = 1
		}

		fmt.Fprintf(os.Stdout, "%v%v%v\n", ne, pad(p), r.Image)
		fmt.Fprintf(os.Stdout, "%v\n\n", color.Format(color.FgHiBlack, wordwrap.WrapString(r.Description, 80)))
	}

	fmt.Fprintf(os.Stdout, "\nSelect from 1..%d: ", len(cc.Registry))
	option, err := prompt.SelectOption(len(cc.Registry), mapSelectOptions)

	if err != nil {
		return err
	}

	cc.Register = cc.Registry[option]
	cc.ServicePackage.Image = cc.Register.Image
	return nil
}

func (cc *serviceCreator) chooseServiceID() (err error) {
	var service = cc.ServicePackage.ID

	if service == "" {
		fmt.Println("Service ID [default: " + cc.Register.ID + "]: ")
		service, err = prompt.Prompt()

		if err != nil {
			return err
		}

		if service == "" {
			service = cc.Register.ID
		}

		err := checkServiceDirectory(service,
			filepath.Join(cc.ProjectDirectory, service))

		if err != nil {
			return err
		}
	}

	cc.ServicePackage.ID = service
	return nil
}

func getBoilerplateServiceImage(cImage string) string {
	cImage = strings.TrimPrefix(cImage, "wedeploy/")

	if strings.Contains(cImage, ":") {
		ws := strings.SplitN(cImage, ":", 2)

		if len(ws) > 1 {
			cImage = ws[0]
		}
	}

	return cImage
}

func (cc *serviceCreator) checkIfDirectoryEmptyForInstallingBoilerplate() (empty bool, err error) {
	notEmpty, err := exists(cc.ServiceDirectory)

	if err != nil {
		return false, err
	}

	if notEmpty {
		if !cc.boilerplateFlagChanged {
			fmt.Fprintf(os.Stderr,
				"Service directory already exists (bypassing installing boilerplate).\n")
			return false, nil
		}

		return false, errors.New("service directory already exists. Can not install boilerplate")
	}

	return true, nil
}

func (cc *serviceCreator) handleBoilerplate() (err error) {
	if !cc.boilerplate {
		return nil
	}

	if empty, err := cc.checkIfDirectoryEmptyForInstallingBoilerplate(); err != nil || !empty {
		return err
	}

	var (
		service = cc.ServicePackage.ID
		cImage  = cc.ServicePackage.Image
	)

	var boilerplateImage = getBoilerplateServiceImage(cImage)
	var boilerplateAddress = fmt.Sprintf(
		"https://github.com/wedeploy/boilerplate-%v.git",
		boilerplateImage)

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

	var boilerplateDotGit = filepath.Join(cc.ServiceDirectory, ".boilerplate-git")

	cmd := exec.Command("git",
		"clone",
		"--depth=1",
		"--quiet",
		"--separate-git-dir="+boilerplateDotGit,
		boilerplateAddress,
		cc.ServiceDirectory)

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
	if err = os.Remove(filepath.Join(cc.ServiceDirectory, ".git")); err != nil {
		return errwrap.Wrapf("Error removing .git ref file for service's boilerplate: {{err}}", err)
	}

	if cc.ServicePackage, err = services.Read(cc.ServiceDirectory); err != nil {
		return errwrap.Wrapf("Can not read boilerplate's service file: {{err}}", err)
	}

	cc.ServicePackage.ID = service
	cc.ServicePackage.Image = cImage
	return nil
}

func (cc *serviceCreator) saveServicePackage() error {
	bin, err := json.MarshalIndent(cc.ServicePackage, "", "    ")

	if err != nil {
		return err
	}

	err = ioutil.WriteFile(filepath.Join(cc.ServiceDirectory, "wedeploy.json"),
		bin, 0644)

	if err == nil {
		abs, aerr := filepath.Abs(cc.ServiceDirectory)

		if aerr != nil {
			fmt.Fprintf(os.Stderr, "Error getting absolute path: %v\n", aerr)
		}

		fmt.Fprintf(os.Stdout, `Service generated at %v
Go to the service directory to keep hacking! :)
Some tips:
	Run the service on your local machine with: "we link"
	Check wedeploy.json for additional configuration.
`, abs)
	}

	return err
}

func (r *runner) newProject() (err error) {
	var p = &projects.Project{
		ProjectID: r.project,
	}

	return r.saveProject(p)
}

func (r *runner) saveProject(p *projects.Project) error {
	if err := tryGenerateDirectory(r.projectBase); err != nil {
		return err
	}

	bin, err := json.MarshalIndent(projects.ProjectPackage{
		ID: p.ProjectID,
	}, "", "    ")

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
	Generate a service there with: "we generate"
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
	var err = os.MkdirAll(directory, 0700)

	if err != nil {
		return errwrap.Wrapf("Can not generate directory: {{err}}", err)
	}

	return err
}

func pad(space int) string {
	return strings.Join(make([]string, space), " ")
}
