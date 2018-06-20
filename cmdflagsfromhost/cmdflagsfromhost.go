package cmdflagsfromhost

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/hashicorp/errwrap"
	"github.com/spf13/cobra"
	"github.com/wedeploy/cli/config"
	"github.com/wedeploy/cli/defaults"
	"github.com/wedeploy/cli/fancy"
	"github.com/wedeploy/cli/flagsfromhost"
	"github.com/wedeploy/cli/inspector"
	"github.com/wedeploy/cli/isterm"
	"github.com/wedeploy/cli/list"
	"github.com/wedeploy/cli/login"
	"github.com/wedeploy/cli/metrics"
	"github.com/wedeploy/cli/services"
)

// Requires configuration for the host and flags
type Requires struct {
	NoHost  bool
	Auth    bool
	Project bool
	Service bool
}

// SetupHost is the structure for host and flags parsing
type SetupHost struct {
	ctx context.Context

	Pattern  Pattern
	Requires Requires

	UseProjectFromWorkingDirectory bool
	UseServiceDirectory            bool

	PromptMissingProject       bool
	PromptMissingService       bool
	AllowMissingProject        bool
	AllowCreateProjectOnPrompt bool
	HideServicesPrompt         bool

	ListExtraDetails list.Pattern

	url     string
	project string
	service string
	remote  string
	cmd     *cobra.Command
	wectx   config.Context
	parsed  *flagsfromhost.FlagsFromHost
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
	return s.wectx.Config().Remotes[s.remote].Infrastructure
}

// ServiceDomain of the parsed flags or host
func (s *SetupHost) ServiceDomain() string {
	return s.wectx.Config().Remotes[s.remote].Service
}

// Host returns the host for a given service or partial host for a given project
func (s *SetupHost) Host() (host string) {
	if s.service != "" {
		host = s.service + "-"
	}

	host += s.project + "." + s.ServiceDomain()
	return host
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

	if none || s.Pattern == missing {
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

	var conf = s.wectx.Config()
	var cffh = flagsfromhost.New(&conf.Remotes)

	return cffh.ParseWithDefaultCustomRemote(
		flagsfromhost.ParseFlagsWithDefaultCustomRemote{
			Host:          s.url,
			Project:       s.project,
			Service:       s.service,
			Remote:        remoteFlagValue,
			RemoteChanged: remoteFlagChanged,
		},
		s.wectx.Config().DefaultRemote)
}

// Process flags
func (s *SetupHost) Process(ctx context.Context, wectx config.Context) (err error) {
	s.ctx = ctx
	s.wectx = wectx
	s.parsed, err = s.tryParseFlags()

	if err != nil {
		return err
	}

	if err = s.loadValues(); err != nil {
		return err
	}

	if !s.Requires.Auth {
		return nil
	}

	return s.verifyCmdReqAuth()
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

func (s *SetupHost) getProjectFromCurrentWorkingDirectory() (project string, err error) {
	var overview = inspector.ContextOverview{}

	if err = overview.Load("."); err != nil {
		return "", err
	}

	if overview.ProjectID == "" {
		return "", nil
	}

	if !isterm.Check() {
		return overview.ProjectID, nil
	}

	fmt.Println(`A reference to a project ID was found within your services.`)

	yes, err := fancy.Boolean(fmt.Sprintf(`Use inferred project ID "%s"?`, overview.ProjectID))

	if err != nil {
		return "", err
	}

	if !yes {
		return "", nil
	}

	return overview.ProjectID, nil
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
		return "", errwrap.Wrapf("error reading current service: {{err}}", err)
	case err == services.ErrServiceNotFound:
		return "", nil
	}

	return service, nil
}

func (s *SetupHost) loadValues() (err error) {
	s.service = s.parsed.Service()
	s.project = s.parsed.Project()
	s.remote = s.parsed.Remote()

	if s.remote == "" {
		s.remote = defaults.CloudRemote
	}

	if s.Pattern&RemotePattern == 0 {
		return errors.New("Remote is not allowed for this command")
	}

	if s.Pattern&ProjectPattern == 0 && s.project != "" {
		return errors.New("Project is not allowed for this command")
	}

	if s.Pattern&ServicePattern == 0 && s.service != "" {
		return errors.New("Service is not allowed for this command")
	}

	if s.project == "" && s.UseProjectFromWorkingDirectory {
		s.project, err = s.getProjectFromCurrentWorkingDirectory()
		if err != nil {
			return err
		}
	}

	if s.service == "" && s.project == "" &&
		s.UseServiceDirectory && !s.PromptMissingService {
		s.service, err = s.getServiceFromCurrentWorkingDirectory()
		if err != nil {
			return err
		}
	}

	if err := s.wectx.SetEndpoint(s.Remote()); err != nil {
		return err
	}

	if (s.Pattern&ProjectPattern == 0 && s.Pattern&ServicePattern == 0) && s.service != "" {
		return errors.New("Service parameter is not allowed for this command")
	}

	if err = s.maybePromptMissing(); err != nil {
		return err
	}

	if s.Requires.Service && s.service == "" {
		return errors.New("Service and project are required")
	}

	if s.Requires.Project && s.project == "" {
		return errors.New("Project is required")
	}

	return nil
}

func (s *SetupHost) maybePromptMissing() (err error) {
	if !isterm.Check() {
		return nil
	}

	if (s.PromptMissingProject && s.project == "") ||
		(s.PromptMissingService && s.service == "") {
		if err = s.verifyCmdReqAuth(); err != nil {
			return err
		}

		if err := s.promptMissingProjectOrService(); err != nil {
			return err
		}
	}

	return nil
}

func (s *SetupHost) promptMissingProjectOrService() (err error) {
	var filter = list.Filter{
		Project: s.project,
	}

	if s.service != "" {
		filter.Services = []string{s.service}
	}

	var l = list.New(filter)
	l.AllowCreateProjectOnPrompt = s.AllowCreateProjectOnPrompt
	l.Details = s.ListExtraDetails

	var selection *list.Selection

	switch {
	case s.Pattern&ProjectPattern != 0 && s.Pattern&ServicePattern != 0 &&
		s.Requires.Service && !s.HideServicesPrompt:
		selection, err = l.PromptService(s.ctx, s.wectx)
	case s.Pattern&ProjectPattern != 0 && s.Pattern&ServicePattern != 0 &&
		s.Requires.Project && !s.HideServicesPrompt:
		selection, err = l.PromptProjectOrService(s.ctx, s.wectx)
	case s.Pattern&ProjectPattern != 0 && s.Requires.Project:
		selection, err = l.PromptProject(s.ctx, s.wectx)
	default:
		return errors.New("not implemented")
	}

	if err != nil {
		return err
	}

	s.project = selection.Project
	s.service = selection.Service
	return nil
}

func (s *SetupHost) verifyCmdReqAuth() error {
	if hasAuth := (s.wectx.Token() != ""); hasAuth {
		return nil
	}

	metrics.Rec(s.wectx.Config(), metrics.Event{
		Type: "required_auth_cmd_precondition_failure",
		Text: fmt.Sprintf(`Command "%v" requires authentication, but one is not set.`, s.cmd.CommandPath()),
		Extra: map[string]string{
			"cmd": s.cmd.CommandPath(),
		},
	})

	return s.authenticateOrCancel()
}

func (s *SetupHost) authenticateOrCancel() error {
	fmt.Printf("You need to log in before using \"%s\".\n",
		strings.TrimSuffix(s.cmd.UseLine(), " [flags]"))

	a := login.Authentication{
		NoLaunchBrowser: false,
	}

	return a.Run(context.Background(), s.wectx)
}
