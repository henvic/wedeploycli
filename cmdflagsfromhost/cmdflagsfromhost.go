package cmdflagsfromhost

import (
	"errors"
	"fmt"

	"github.com/hashicorp/errwrap"
	"github.com/spf13/cobra"
	"github.com/wedeploy/cli/config"
	"github.com/wedeploy/cli/defaults"
	"github.com/wedeploy/cli/flagsfromhost"
	"github.com/wedeploy/cli/projects"
	"github.com/wedeploy/cli/services"
	"github.com/wedeploy/cli/wdircontext"
)

// Requires configuration for the host and flags
type Requires struct {
	NoHost  bool
	Auth    bool
	Remote  bool
	Local   bool
	Project bool
	Service bool
}

// SetupHost is the structure for host and flags parsing
type SetupHost struct {
	Pattern                       Pattern
	Requires                      Requires
	UseProjectDirectory           bool
	UseProjectDirectoryForService bool
	UseServiceDirectory           bool
	url                           string
	project                       string
	service                       string
	remote                        string
	cmd                           *cobra.Command
	parsed                        *flagsfromhost.FlagsFromHost
}

// Pattern for the host and flags
type Pattern int

const (
	missing Pattern = 1 << iota
	// RemotePattern takes only --remote
	RemotePattern
	// ServicePattern takes only --service
	ServicePattern
	// ProjectPattern takes only --project
	ProjectPattern
	// ProjectAndServicePattern takes only --project and --service
	ProjectAndServicePattern = ProjectPattern | ServicePattern
	// ProjectAndRemotePattern takes only --project, and --remote
	ProjectAndRemotePattern = ProjectPattern | RemotePattern
	// FullHostPattern takes --project, --service, and --remote
	FullHostPattern = RemotePattern | ProjectAndServicePattern
)

// Project of the parsed flags or host
func (s *SetupHost) Project() string {
	return s.project
}

// Service of the parsed flags or host
func (s *SetupHost) Service() string {
	return s.service
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

	if !s.Requires.NoHost && (s.Pattern&RemotePattern != 0 || s.Pattern&ServicePattern != 0) {
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

	if s.Pattern&ServicePattern != 0 {
		s.addServiceFlag(cmd)
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
				Host:    s.url,
				Project: s.project,
				Service: s.service,
				Remote:  remoteFlagValue,
			}))
	}

	return s.applyParseFlagsFilters(flagsfromhost.ParseWithDefaultCustomRemote(
		flagsfromhost.ParseFlagsWithDefaultCustomRemote{
			Host:          s.url,
			Project:       s.project,
			Service:       s.service,
			Remote:        remoteFlagValue,
			RemoteChanged: remoteFlagChanged,
		},
		config.Global.DefaultRemote))
}

func (s *SetupHost) applyParseFlagsFilters(f *flagsfromhost.FlagsFromHost, err error) (
	*flagsfromhost.FlagsFromHost, error) {
	if (s.UseProjectDirectory || s.UseProjectDirectoryForService) && err != nil {
		switch err.(type) {
		case flagsfromhost.ErrorServiceWithNoProject:
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
	cmd.Flags().StringVarP(&s.url, "url", "u", "", "Perform the operation for a specific URL (host)")
}

func (s *SetupHost) addRemoteFlag(cmd *cobra.Command) {
	cmd.Flags().StringVarP(
		&s.remote,
		"remote", "r", "current", "Perform the operation for a specific remote")
}

func (s *SetupHost) addProjectFlag(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&s.project, "project", "p", "", "Perform the operation for a specific project")
}

func (s *SetupHost) addServiceFlag(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&s.service, "service", "c", "", "Perform the operation for a specific service")
}

func (s *SetupHost) getServiceFromCurrentWorkingDirectory() (service string, err error) {
	if s.Pattern&ProjectPattern == 0 {
		return "", nil
	}

	service, err = wdircontext.GetServiceID()

	switch {
	case err != nil && err != services.ErrServiceNotFound:
		return "", errwrap.Wrapf("Error reading current service: {{err}}", err)
	case err == services.ErrServiceNotFound:
		return "", nil
	}

	return service, nil
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
		if !s.Requires.Service || s.Pattern&ServicePattern == 0 {
			return "", errwrap.Wrapf(
				"Use flag --project or call this from inside a project directory", err)
		}

		if !s.UseServiceDirectory {
			return "", errwrap.Wrapf("Use flags --project and --service", err)
		}

		return "", errwrap.Wrapf(
			"Use flags --project and --service or call this from inside a project service directory", err)
	}

	return project, nil
}

func (s *SetupHost) loadValues() (err error) {
	var service = s.parsed.Service()
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

	if s.Pattern&ServicePattern == 0 && service != "" {
		return errors.New("Service is not allowed for this command")
	}

	if project == "" && (s.UseProjectDirectory || (s.UseProjectDirectoryForService && service != "")) {
		project, err = s.getProjectFromCurrentWorkingDirectory()
		if err != nil {
			return err
		}
	}

	if service == "" && s.UseServiceDirectory && s.parsed.Project() == "" {
		service, err = s.getServiceFromCurrentWorkingDirectory()
		if err != nil {
			return err
		}
	}

	if (s.Pattern&ProjectPattern == 0 && s.Pattern&ServicePattern == 0) && service != "" {
		return errors.New("Service parameter is not allowed for this command")
	}

	if s.Requires.Service && service == "" {
		return errors.New("Service and project are required")
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

	s.service = service
	s.project = project
	s.remote = remote

	return config.SetEndpointContext(s.Remote())
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

	return fmt.Errorf(`not logged on remote "%v". Please run "we login" before using "we %v"`,
		s.Remote(),
		s.cmd.Name())
}
