package cmdflagsfromhost

import (
	"context"
	"errors"
	"fmt"

	"github.com/hashicorp/errwrap"
	"github.com/spf13/cobra"
	"github.com/wedeploy/cli/config"
	"github.com/wedeploy/cli/defaults"
	"github.com/wedeploy/cli/fancy"
	"github.com/wedeploy/cli/flagsfromhost"
	"github.com/wedeploy/cli/login"
	"github.com/wedeploy/cli/metrics"
	"github.com/wedeploy/cli/services"
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
	Pattern             Pattern
	Requires            Requires
	UseServiceDirectory bool
	AllowMissingProject bool
	url                 string
	project             string
	service             string
	remote              string
	cmd                 *cobra.Command
	parsed              *flagsfromhost.FlagsFromHost
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

// InfrastructureDomain of the parsed flags or host
func (s *SetupHost) InfrastructureDomain() string {
	return config.Global.Remotes[s.remote].Infrastructure
}

// ServiceDomain of the parsed flags or host
func (s *SetupHost) ServiceDomain() string {
	return config.Global.Remotes[s.remote].Service
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

func (s *SetupHost) tryParseFlags() (*flagsfromhost.FlagsFromHost, error) {
	flags, err := s.parseFlags()

	switch err.(type) {
	case flagsfromhost.ErrorServiceWithNoProject:
		if s.Pattern&ServicePattern != 0 && s.AllowMissingProject && (s.url == "") {
			return flags, nil
		}
	}

	return flags, err
}

func (s *SetupHost) parseFlags() (*flagsfromhost.FlagsFromHost, error) {
	var remoteFlag = s.cmd.Flag("remote")
	var remoteFlagValue = s.remote
	var remoteFlagChanged bool

	if remoteFlag != nil && remoteFlag.Changed {
		remoteFlagChanged = true
	}

	if s.Requires.Local {
		return flagsfromhost.Parse(
			flagsfromhost.ParseFlags{
				Host:    s.url,
				Project: s.project,
				Service: s.service,
				Remote:  remoteFlagValue,
			})
	}

	return flagsfromhost.ParseWithDefaultCustomRemote(
		flagsfromhost.ParseFlagsWithDefaultCustomRemote{
			Host:          s.url,
			Project:       s.project,
			Service:       s.service,
			Remote:        remoteFlagValue,
			RemoteChanged: remoteFlagChanged,
		},
		config.Global.DefaultRemote)
}

// Process flags
func (s *SetupHost) Process() (err error) {
	s.parsed, err = s.tryParseFlags()

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
	cmd.Flags().StringVarP(&s.service, "service", "s", "", "Perform the operation for a specific service")
}

func (s *SetupHost) getServiceFromCurrentWorkingDirectory() (service string, err error) {
	if s.Pattern&ProjectPattern == 0 {
		return "", nil
	}

	sp, err := services.Read(".")

	if err != nil {
		return "", err
	}

	service = sp.ID

	switch {
	case err != nil && err != services.ErrServiceNotFound:
		return "", errwrap.Wrapf("Error reading current service: {{err}}", err)
	case err == services.ErrServiceNotFound:
		return "", nil
	}

	return service, nil
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

	if hasAuth := (config.Context.Token != ""); hasAuth {
		return nil
	}

	metrics.Rec(metrics.Event{
		Type: "required_auth_cmd_precondition_failure",
		Text: fmt.Sprintf(`Command "%v" requires authentication, but one is not set.`, s.cmd.CommandPath()),
		Extra: map[string]string{
			"cmd": s.cmd.CommandPath(),
		},
	})

	return authenticateOrCancel(s.cmd)
}

func authenticateOrCancel(cmd *cobra.Command) error {
	var options = fancy.Options{}
	options.Add("A", "Yes")
	options.Add("B", "Cancel")

	var choice, err = options.Ask(fmt.Sprintf(`You need to log in before performing "%s". Do you want to log in?`, cmd.UseLine()))

	if err != nil {
		return err
	}

	if choice == "B" {
		return login.CancelCommand("Login canceled.")
	}

	a := login.Authentication{
		NoLaunchBrowser: false,
	}

	return a.Run(context.Background())
}
