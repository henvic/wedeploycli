package cmdflagsfromhost

import (
	"errors"
	"fmt"
	"strings"

	"github.com/hashicorp/errwrap"
	"github.com/spf13/cobra"
	"github.com/wedeploy/cli/config"
	"github.com/wedeploy/cli/containers"
	"github.com/wedeploy/cli/defaults"
	"github.com/wedeploy/cli/flagsfromhost"
	"github.com/wedeploy/cli/projects"
	"github.com/wedeploy/cli/remoteuriparser"
	"github.com/wedeploy/cli/wdircontext"
)

// SetLocal the context
func SetLocal() {
	config.Context.Remote = "local"
	config.Context.Username = "no-reply@wedeploy.com"
	config.Context.Password = "cli-tool-password"
	config.Context.Token = ""
	config.Context.Endpoint = config.Global.LocalEndpoint
	config.Context.RemoteAddress = "wedeploy.me"
}

// Requires configuration for the host and flags
type Requires struct {
	NoHost    bool
	Auth      bool
	Remote    bool
	Local     bool
	Project   bool
	Container bool
}

// SetupHost is the structure for host and flags parsing
type SetupHost struct {
	Pattern                         Pattern
	Requires                        Requires
	UseProjectDirectory             bool
	UseProjectDirectoryForContainer bool
	UseContainerDirectory           bool
	url                             string
	project                         string
	container                       string
	remote                          string
	cmd                             *cobra.Command
	parsed                          *flagsfromhost.FlagsFromHost
}

// Pattern for the host and flags
type Pattern int

const (
	missing Pattern = 1 << iota
	// RemotePattern takes only --remote
	RemotePattern
	// ContainerPattern takes only --container
	ContainerPattern
	// ProjectPattern takes only --project
	ProjectPattern
	// ProjectAndContainerPattern takes only --project and --container
	ProjectAndContainerPattern = ProjectPattern | ContainerPattern
	// ProjectAndRemotePattern takes only --project, and --remote
	ProjectAndRemotePattern = ProjectPattern | RemotePattern
	// FullHostPattern takes --project, --container, and --remote
	FullHostPattern = RemotePattern | ProjectAndContainerPattern
)

// Project of the parsed flags or host
func (s *SetupHost) Project() string {
	return s.project
}

// Container of the parsed flags or host
func (s *SetupHost) Container() string {
	return s.container
}

// Remote of the parsed flags or host
func (s *SetupHost) Remote() string {
	return s.remote
}

// RemoteAddress of the parsed flags or host
func (s *SetupHost) RemoteAddress() string {
	return config.Global.Remotes[s.remote].URL
}

// Init flags on a given command
func (s *SetupHost) Init(cmd *cobra.Command) {
	var none = true
	s.cmd = cmd

	if !s.Requires.NoHost && (s.Pattern&RemotePattern != 0 || s.Pattern&ContainerPattern != 0) {
		s.addURLFlag(cmd)
	}

	if s.Pattern&RemotePattern != 0 {
		s.addRemoteFlag(cmd)
		none = false
	}

	if s.Pattern&ProjectPattern != 0 {
		s.addProjectFlag(cmd)
		none = false
	}

	if s.Pattern&ContainerPattern != 0 {
		s.addContainerFlag(cmd)
		none = false
	}

	if none {
		panic("Missing or unsupported host pattern")
	}
}

func (s *SetupHost) parseFlags() (*flagsfromhost.FlagsFromHost, error) {
	var remoteFlag = s.cmd.Flag("remote")
	var remoteFlagValue = s.remote
	var remoteFlagChanged bool

	if remoteFlag != nil && remoteFlag.Changed {
		remoteFlagChanged = true
	}

	if s.Requires.Local {
		return s.applyParseFlagsFilters(flagsfromhost.Parse(
			flagsfromhost.ParseFlags{
				Host:      s.url,
				Project:   s.project,
				Container: s.container,
				Remote:    remoteFlagValue,
			}))
	}

	return s.applyParseFlagsFilters(flagsfromhost.ParseWithDefaultCustomRemote(
		flagsfromhost.ParseFlagsWithDefaultCustomRemote{
			Host:          s.url,
			Project:       s.project,
			Container:     s.container,
			Remote:        remoteFlagValue,
			RemoteChanged: remoteFlagChanged,
		},
		config.Global.DefaultRemote))
}

func (s *SetupHost) applyParseFlagsFilters(f *flagsfromhost.FlagsFromHost, err error) (
	*flagsfromhost.FlagsFromHost, error) {
	if (s.UseProjectDirectory || s.UseProjectDirectoryForContainer) && err != nil {
		switch err.(type) {
		case flagsfromhost.ErrorContainerWithNoProject:
			err = nil
		}
	}

	return f, err
}

// Process flags
func (s *SetupHost) Process() (err error) {
	s.parsed, err = s.parseFlags()

	if err != nil {
		return err
	}

	if err = s.loadValues(); err != nil {
		return err
	}

	if err := s.verifyCmdReqAuth(); err != nil {
		return err
	}

	return nil
}

func (s *SetupHost) addURLFlag(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&s.url, "url", "u", "", "URL host for resource")
}

func (s *SetupHost) addRemoteFlag(cmd *cobra.Command) {
	cmd.Flags().StringVarP(
		&s.remote,
		"remote", "r", "current", "Remote to use")
}

func (s *SetupHost) addProjectFlag(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&s.project, "project", "p", "", "Project ID")
}

func (s *SetupHost) addContainerFlag(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&s.container, "container", "c", "", "Container ID")
}

func (s *SetupHost) getContainerFromCurrentWorkingDirectory() (container string, err error) {
	if s.Pattern&ProjectPattern == 0 {
		return "", nil
	}

	container, err = wdircontext.GetContainerID()

	switch {
	case err != nil && err != containers.ErrContainerNotFound:
		return "", errwrap.Wrapf("Error reading current container: {{err}}", err)
	case err == containers.ErrContainerNotFound:
		return "", nil
	}

	return container, nil
}

func (s *SetupHost) getProjectFromCurrentWorkingDirectory() (project string, err error) {
	if s.Pattern&ProjectPattern == 0 {
		return "", nil
	}

	project, err = wdircontext.GetProjectID()

	switch {
	case err != nil && err != projects.ErrProjectNotFound:
		return "", errwrap.Wrapf("Error reading current project: {{err}}", err)
	case err == projects.ErrProjectNotFound:
		if !s.Requires.Container || s.Pattern&ContainerPattern == 0 {
			return "", errwrap.Wrapf(
				"Use flag --project or call this from inside a project directory", err)
		}

		if !s.UseContainerDirectory {
			return "", errwrap.Wrapf("Use flags --project and --container", err)
		}

		return "", errwrap.Wrapf(
			"Use flags --project and --container or call this from inside a project container directory", err)
	}

	return project, nil
}

func (s *SetupHost) loadValues() (err error) {
	var container = s.parsed.Container()
	var project = s.parsed.Project()
	var remote = s.parsed.Remote()

	if remote == "" {
		remote = defaults.LocalRemote
	}

	if s.Pattern&RemotePattern == 0 && remote != defaults.LocalRemote {
		return errors.New("Remote is not allowed for this command")
	}

	if s.Pattern&ProjectPattern == 0 && project != "" {
		return errors.New("Project is not allowed for this command")
	}

	if s.Pattern&ContainerPattern == 0 && container != "" {
		return errors.New("Container is not allowed for this command")
	}

	if project == "" && (s.UseProjectDirectory || (s.UseProjectDirectoryForContainer && container != "")) {
		project, err = s.getProjectFromCurrentWorkingDirectory()
		if err != nil {
			return err
		}
	}

	if container == "" && s.UseContainerDirectory && s.parsed.Project() == "" {
		container, err = s.getContainerFromCurrentWorkingDirectory()
		if err != nil {
			return err
		}
	}

	if (s.Pattern&ProjectPattern == 0 && s.Pattern&ContainerPattern == 0) && container != "" {
		return errors.New("Container parameter is not allowed for this command")
	}

	if s.Requires.Container && container == "" {
		return errors.New("Container and project are required")
	}

	if s.Requires.Project && project == "" {
		return errors.New("Project is required")
	}

	if s.Requires.Remote && remote == defaults.LocalRemote {
		return errors.New(`Remote is required and can not be "local"`)
	}

	if s.Requires.Local && remote != defaults.LocalRemote {
		return errors.New("Remote parameter is not allowed for this command")
	}

	s.container = container
	s.project = project
	s.remote = remote

	return s.setEndpoint()
}

func (s *SetupHost) verifyCmdReqAuth() error {
	if !s.Requires.Auth {
		return nil
	}

	if s.Remote() == defaults.LocalRemote {
		return nil
	}

	var c = config.Context

	var hasAuth = (c.Token != "") || (c.Username != "" && c.Password != "")

	if hasAuth {
		return nil
	}

	return errors.New(`Please run "we login" before using "we ` + s.cmd.Name() + `".`)
}

func (s *SetupHost) setEndpoint() error {
	if s.Remote() == defaults.LocalRemote {
		SetLocal()
		return nil
	}

	return SetRemote(s.Remote())
}

// SetRemote sets the remote for the current context
func SetRemote(remote string) (err error) {
	var r, ok = config.Global.Remotes[remote]

	if !ok {
		return fmt.Errorf(`Error loading selected remote "%v"`, remote)
	}

	config.Context.Remote = remote
	config.Context.RemoteAddress = getRemoteAddress(r.URL)
	config.Context.Endpoint = remoteuriparser.Parse(r.URL)
	config.Context.Username = r.Username
	config.Context.Password = r.Password
	config.Context.Token = r.Token
	return nil
}

func getRemoteAddress(address string) string {
	var removePrefixes = []string{
		"http://",
		"https://",
	}

	for _, prefix := range removePrefixes {
		if strings.HasPrefix(address, prefix) {
			return strings.TrimPrefix(address, prefix)
		}
	}

	return address
}
