package cmdflagsfromhost

import (
	"errors"
	"strings"

	"github.com/hashicorp/errwrap"
	"github.com/spf13/cobra"
	"github.com/wedeploy/cli/apihelper"
	"github.com/wedeploy/cli/config"
	"github.com/wedeploy/cli/containers"
	"github.com/wedeploy/cli/flagsfromhost"
	"github.com/wedeploy/cli/projects"
	"github.com/wedeploy/cli/remoteuriparser"
	"github.com/wedeploy/cli/wdircontext"
)

// SetLocal the context
func SetLocal() {
	config.Context.Token = apihelper.DefaultToken
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
	project                         string
	container                       string
	remote                          string
	remoteAddress                   string
	cmdName                         string
	parsed                          *flagsfromhost.FlagsFromHost
}

// Pattern for the host and flags
type Pattern int

const (
	missing Pattern = iota
	// RemotePattern takes only --remote
	RemotePattern
	// ProjectPattern takes only --project
	ProjectPattern
	// ProjectAndContainerPattern takes only --project and --container
	ProjectAndContainerPattern
	// FullHostPattern takes --project, --container, and --remote
	FullHostPattern
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
	return s.remoteAddress
}

// Init flags on a given command
func (s *SetupHost) Init(cmd *cobra.Command) {
	s.cmdName = cmd.Name()

	switch s.Pattern {
	case RemotePattern:
		s.addRemoteFlag(cmd)
	case ProjectPattern:
		s.addProjectFlag(cmd)
	case ProjectAndContainerPattern:
		s.addProjectAndContainerFlags(cmd)
	case FullHostPattern:
		s.addRemoteFlag(cmd)
		s.addProjectAndContainerFlags(cmd)
	default:
		panic("Missing host pattern")
	}
}

// Process host and flags
func (s *SetupHost) Process(args []string) error {
	if s.Requires.NoHost && len(args) != 0 {
		return errors.New("Values by host style is disabled for this command")
	}

	switch len(args) {
	case 0:
		return s.process("")
	case 1:
		return s.process(args[0])
	default:
		return errors.New("Wrong number of arguments (expected only host)")
	}
}

func (s *SetupHost) parseFlags(host string) (f *flagsfromhost.FlagsFromHost, err error) {
	f, err = flagsfromhost.Parse(host, s.project, s.container, s.remote)

	if (s.UseProjectDirectory || s.UseProjectDirectoryForContainer) && err != nil {
		switch err.(type) {
		case flagsfromhost.ErrorContainerWithNoProject:
			err = nil
		}
	}

	return f, err
}

func (s *SetupHost) process(host string) (err error) {
	s.parsed, err = s.parseFlags(host)

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

func (s *SetupHost) addRemoteFlag(cmd *cobra.Command) {
	cmd.Flags().StringVarP(
		&s.remote,
		"remote", "r", "", "Remote to use")
}

func (s *SetupHost) addProjectFlag(cmd *cobra.Command) {
	cmd.Flags().StringVar(&s.project, "project", "", "Project ID")
}

func (s *SetupHost) addProjectAndContainerFlags(cmd *cobra.Command) {
	s.addProjectFlag(cmd)
	cmd.Flags().StringVar(&s.container, "container", "", "Container ID")
}

func (s *SetupHost) getContainerFromCurrentWorkingDirectory() (container string, err error) {
	if s.Pattern != ProjectPattern && s.Pattern != ProjectAndContainerPattern && s.Pattern != FullHostPattern {
		return "", nil
	}

	container, err = wdircontext.GetContainerID()

	switch {
	case err != nil && err != containers.ErrContainerNotFound:
		return "", errwrap.Wrapf("Error extracting container ID from current directory's context: {{err}}", err)
	case err == containers.ErrContainerNotFound:
		return "", nil
	}

	return container, nil
}

func (s *SetupHost) getProjectFromCurrentWorkingDirectory() (project string, err error) {
	if s.Pattern != ProjectPattern && s.Pattern != ProjectAndContainerPattern && s.Pattern != FullHostPattern {
		return "", nil
	}

	project, err = wdircontext.GetProjectID()

	switch {
	case err != nil && err != projects.ErrProjectNotFound:
		return "", errwrap.Wrapf("Error extracting project ID from current directory's context: {{err}}", err)
	case err == projects.ErrProjectNotFound:
		return "", errwrap.Wrapf("{{err}} or local project.json context",
			flagsfromhost.ErrorContainerWithNoProject{})
	}

	return project, nil
}

func (s *SetupHost) loadValues() (err error) {
	var container = s.parsed.Container()
	var project = s.parsed.Project()
	var remote = s.parsed.Remote()

	if container == "" && s.UseContainerDirectory {
		container, err = s.getContainerFromCurrentWorkingDirectory()
		if err != nil {
			return err
		}
	}

	if project == "" && (s.UseProjectDirectory || (s.UseProjectDirectoryForContainer && container != "")) {
		project, err = s.getProjectFromCurrentWorkingDirectory()
		if err != nil {
			return err
		}
	}

	if (s.Pattern != ProjectAndContainerPattern && s.Pattern != FullHostPattern) && container != "" {
		return errors.New("Container parameter is not allowed for this command")
	}

	if s.Requires.Container && container == "" {
		return errors.New("Container and project are required")
	}

	if s.Requires.Project && project == "" {
		return errors.New("Project is required")
	}

	if s.Requires.Remote && remote == "" {
		return errors.New("Remote is required")
	}

	if s.Requires.Local && remote != "" {
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

	if s.parsed.Remote() == "" {
		return nil
	}

	var g = config.Global

	var hasAuth = (g.Token != "") || (g.Username != "" && g.Password != "")

	if hasAuth {
		return nil
	}

	return errors.New(`Please run "we login" before using "we ` + s.cmdName + `".`)
}

func (s *SetupHost) setEndpoint() error {
	if s.parsed.Remote() == "" {
		SetLocal()
		return nil
	}

	return s.setRemote()
}

func (s *SetupHost) setRemote() (err error) {
	var r = config.Global.Remotes[s.parsed.Remote()]
	config.Context.Remote = s.parsed.Remote()
	config.Context.RemoteAddress = getRemoteAddress(r.URL)
	config.Context.Endpoint = remoteuriparser.Parse(r.URL)
	config.Context.Username = config.Global.Username
	config.Context.Password = config.Global.Password
	config.Context.Token = config.Global.Token
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
