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
	"github.com/wedeploy/cli/flagsfromhost"
	"github.com/wedeploy/cli/inspector"
	"github.com/wedeploy/cli/instances"
	"github.com/wedeploy/cli/isterm"
	"github.com/wedeploy/cli/links"
	"github.com/wedeploy/cli/list"
	"github.com/wedeploy/cli/listinstances"
	"github.com/wedeploy/cli/login"
	"github.com/wedeploy/cli/metrics"
	"github.com/wedeploy/cli/services"
)

// Requires configuration for the host and flags
type Requires struct {
	NoHost bool
	Auth   bool

	Region   bool
	Project  bool
	Service  bool
	Instance bool
}

// SetupHost is the structure for host and flags parsing
type SetupHost struct {
	ctx context.Context

	Pattern  Pattern
	Requires Requires

	UseProjectFromWorkingDirectory bool
	UseServiceDirectory            bool
	AutoSelectSingleInstance       bool

	PromptMissingProject  bool
	PromptMissingService  bool
	PromptMissingInstance bool

	AllowMissingProject        bool
	AllowCreateProjectOnPrompt bool

	HideServicesPrompt bool

	ListExtraDetails list.Pattern

	url      string
	region   string
	project  string
	service  string
	instance string
	remote   string

	cmd    *cobra.Command
	wectx  config.Context
	parsed *flagsfromhost.FlagsFromHost

	tmpEnv string
}

// Pattern for the host and flags
type Pattern int

const (
	missing Pattern = 1 << iota
	// RegionPattern takes only --region
	RegionPattern
	// RemotePattern takes only --remote
	RemotePattern
	// ServicePattern takes only --service
	ServicePattern
	// ProjectPattern takes only --project
	ProjectPattern
	// InstancePattern takes only --instance
	InstancePattern
	// ProjectAndServicePattern takes only --project and --service
	ProjectAndServicePattern = ProjectPattern | ServicePattern
	// ProjectAndRemotePattern takes only --project, and --remote
	ProjectAndRemotePattern = ProjectPattern | RemotePattern
	// FullHostPattern takes --project, --service, and --remote
	FullHostPattern = RemotePattern | ProjectAndServicePattern

	anyInstance = "any" // magic keyword for choosing any instance
)

// Region of the parsed flags
func (s *SetupHost) Region() string {
	return s.region
}

// Project of the parsed flags or host
func (s *SetupHost) Project() string {
	return s.project
}

// Environment of the parsed flags or host
func (s *SetupHost) Environment() string {
	if i := strings.Index(s.project, "-"); i != -1 {
		return s.project[i+1:]
	}

	return s.project
}

// Service of the parsed flags or host
func (s *SetupHost) Service() string {
	return s.service
}

// Instance of the parsed flag
func (s *SetupHost) Instance() string {
	return s.instance
}

// Remote of the parsed flags or host
func (s *SetupHost) Remote() string {
	return s.remote
}

// InfrastructureDomain of the parsed flags or host
func (s *SetupHost) InfrastructureDomain() string {
	var conf = s.wectx.Config()
	var params = conf.GetParams()
	var rl = params.Remotes
	return rl.Get(s.remote).Infrastructure
}

// ServiceDomain of the parsed flags or host
func (s *SetupHost) ServiceDomain() string {
	var conf = s.wectx.Config()
	var params = conf.GetParams()
	var rl = params.Remotes
	return rl.Get(s.remote).Service
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

	if s.Pattern&RegionPattern != 0 {
		s.addRegionFlag(cmd)
	}

	if s.Pattern&ProjectPattern != 0 {
		s.addProjectFlag(cmd)
		none = false
	}

	if s.Pattern&ServicePattern != 0 {
		s.addServiceFlag(cmd)
		none = false
	}

	if s.Pattern&ServicePattern == 0 && s.Pattern&InstancePattern != 0 {
		panic("Instance pattern requires service pattern")
	}

	if s.Pattern&InstancePattern != 0 {
		s.addInstanceFlag(cmd)
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

func (s *SetupHost) tryInstance() error {
	if s.instance == "" || s.instance == anyInstance || (s.service != "" && s.project != "") {
		return nil
	}

	instancesClient := instances.New(s.wectx)

	l, err := instancesClient.List(s.ctx, instances.Filter{
		InstanceID: s.instance,
		ServiceID:  s.service,
		ProjectID:  s.project,
	})

	if err != nil {
		return errwrap.Wrapf("cannot list available instances: {{err}}", err)
	}

	if len(l) == 0 {
		return fmt.Errorf(`no instance found starting with "%s"`, s.instance)
	}

	if len(l) > 1 {
		return fmt.Errorf(`"%s" is not distinct enough: %d instances found`, s.instance, len(l))
	}

	instance := l[0]

	s.instance = instance.InstanceID
	s.service = instance.ServiceID
	s.project = instance.ProjectID

	return nil
}

func (s *SetupHost) parseFlags() (*flagsfromhost.FlagsFromHost, error) {
	var remoteFlag = s.cmd.Flag("remote")
	var remoteFlagValue = s.remote
	var remoteFlagChanged bool

	if remoteFlag != nil && remoteFlag.Changed {
		remoteFlagChanged = true
	}

	var conf = s.wectx.Config()
	var params = conf.GetParams()
	var rl = params.Remotes
	var cffh = flagsfromhost.New(rl)

	if err := s.injectEnvFlagInProject(); err != nil {
		return nil, err
	}

	var instanceFlag = s.cmd.Flag("instance")

	if instanceFlag != nil && instanceFlag.Changed {
		s.instance = instanceFlag.Value.String()
	}

	return cffh.ParseWithDefaultCustomRemote(
		flagsfromhost.ParseFlagsWithDefaultCustomRemote{
			Host:          s.url,
			Project:       s.project,
			Service:       s.service,
			Remote:        remoteFlagValue,
			RemoteChanged: remoteFlagChanged,
		},
		params.DefaultRemote)
}

func (s *SetupHost) injectEnvFlagInProject() error {
	if s.tmpEnv == "" || s.project == "" && s.tmpEnv == "" {
		return nil
	}

	if s.project == "" && s.tmpEnv != "" {
		return errors.New("incompatible use: --environment requires --project")
	}

	if strings.Contains(s.project, "-") && s.tmpEnv != "" {
		return errors.New("incompatible use: environment value cannot be passed both on --project and --environment")
	}

	s.project = s.project + "-" + s.tmpEnv
	return nil
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

	if strings.HasSuffix(wectx.InfrastructureDomain(), ".liferay.cloud") {
		links.SetDXP()
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

func (s *SetupHost) addRegionFlag(cmd *cobra.Command) {
	cmd.Flags().StringVar(
		&s.region,
		"region", "", "Perform the operation for a specific region")
}

func (s *SetupHost) addProjectFlag(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&s.project, "project", "p", "", "Perform the operation for a specific project")
	cmd.Flags().StringVarP(&s.tmpEnv, "environment", "e", "", "Perform the operation for a specific environment")
}

func (s *SetupHost) addServiceFlag(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&s.service, "service", "s", "", "Perform the operation for a specific service")
}

func (s SetupHost) addInstanceFlag(cmd *cobra.Command) {
	cmd.Flags().StringVar(&s.instance, "instance", "", "Perform the operation for a specific instance")
}

func (s *SetupHost) getProjectFromCurrentWorkingDirectory() (project string, err error) {
	var overview = inspector.ContextOverview{}

	err = overview.Load(".")
	return overview.ProjectID, err
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
		return errors.New("remote is not allowed for this command")
	}

	if s.Pattern&ProjectPattern == 0 && s.project != "" {
		return errors.New("project is not allowed for this command")
	}

	if s.Pattern&ServicePattern == 0 && s.service != "" {
		return errors.New("service is not allowed for this command")
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

	if err = s.wectx.SetEndpoint(s.Remote()); err != nil {
		return err
	}

	if (s.Pattern&ProjectPattern == 0 && s.Pattern&ServicePattern == 0) && s.service != "" {
		return errors.New("service parameter is not allowed for this command")
	}

	if s.Pattern&InstancePattern != 0 {
		if err = s.tryInstance(); err != nil {
			return err
		}
	}

	if err = s.maybePromptMissing(); err != nil {
		return err
	}

	if s.Requires.Instance && s.instance == "" {
		return errors.New(`instance, service, and project are required (try "--instance any")`)
	}

	if s.instance == anyInstance {
		s.instance = ""
	}

	if s.Requires.Service && s.service == "" {
		return errors.New("service and project are required")
	}

	if s.Requires.Project && s.project == "" {
		return errors.New("project is required")
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

		if err = s.promptMissingProjectOrService(); err != nil {
			return err
		}
	}

	if s.PromptMissingInstance && s.instance == "" {
		if err = s.promptMissingInstanceF(); err != nil {
			return err
		}
	}

	return nil
}

func (s *SetupHost) promptMissingInstanceF() (err error) {
	var li = listinstances.New(s.project, s.service)
	var f = li.Prompt

	if s.AutoSelectSingleInstance {
		f = li.AutoSelectOrPrompt
	}

	s.instance, err = f(s.ctx, s.wectx)
	return err
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

	if s.project == "" {
		s.project = selection.Project
	}

	if s.service == "" {
		s.service = selection.Service
	}

	return nil
}

func (s *SetupHost) verifyCmdReqAuth() error {
	if hasAuth := (s.wectx.Token() != ""); hasAuth {
		return nil
	}

	metrics.Rec(s.wectx.Config(), metrics.Event{
		Type: "required_auth_failure",
		Text: fmt.Sprintf(`command "%v" requires authentication`, s.cmd.CommandPath()),
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
