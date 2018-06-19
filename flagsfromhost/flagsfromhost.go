package flagsfromhost

import (
	"fmt"
	"strings"

	"github.com/wedeploy/cli/defaults"
	"github.com/wedeploy/cli/remotes"
)

// ErrorRemoteFlagAndHost happens when both --remote and host remote are used together
type ErrorRemoteFlagAndHost struct{}

func (ErrorRemoteFlagAndHost) Error() string {
	return "incompatible use: --remote flag can not be used along host format with remote address"
}

// ErrorMultiMode happens when --project and --service are used with host URL flag
type ErrorMultiMode struct{}

func (ErrorMultiMode) Error() string {
	return "incompatible use: --project and --service are not allowed with host URL flag"
}

// ErrorServiceWithNoProject hapens when --service is used without --project
type ErrorServiceWithNoProject struct{}

func (ErrorServiceWithNoProject) Error() string {
	return "incompatible use: --service requires --project"
}

// ErrorLoadingRemoteList happens when the remote list is needed, but not found
type ErrorLoadingRemoteList struct{}

func (ErrorLoadingRemoteList) Error() string {
	return "error loading remotes list"
}

// ErrorFoundNoRemote happens when a remote isn't found for a given host address
type ErrorFoundNoRemote struct {
	Host string
}

func (e ErrorFoundNoRemote) Error() string {
	return fmt.Sprintf("found no remote for address %v", e.Host)
}

// ErrorNotFound happens when a remote isn't found
type ErrorNotFound struct {
	Remote string
}

func (e ErrorNotFound) Error() string {
	return fmt.Sprintf("remote %v not found", e.Remote)
}

// ErrorFoundMultipleRemote happens when multiple resolutions are found for a given host address
type ErrorFoundMultipleRemote struct {
	Host    string
	Remotes []string
}

func (e ErrorFoundMultipleRemote) Error() string {
	return fmt.Sprintf("found multiple remotes for address %v: %v",
		e.Host,
		strings.Join(e.Remotes, ", "))
}

// FlagsFromHost holds the project, service, and remote parsed
type FlagsFromHost struct {
	project          string
	service          string
	remote           string
	isRemoteFromHost bool
}

// Project of the parsed flags or host
func (f *FlagsFromHost) Project() string {
	return f.project
}

// Service of the parsed flags or host
func (f *FlagsFromHost) Service() string {
	return f.service
}

// Remote of the parsed flags or host
func (f *FlagsFromHost) Remote() string {
	return f.remote
}

// IsRemoteFromHost tells if the Remote was parsed from a host or not
func (f *FlagsFromHost) IsRemoteFromHost() bool {
	return f.isRemoteFromHost
}

// CommandFlagFromHost extractor
type CommandFlagFromHost struct {
	RemotesList *remotes.List
}

// New CommandFlagFromHost
func New(remotesList *remotes.List) CommandFlagFromHost {
	return CommandFlagFromHost{
		remotesList,
	}
}

// ParseFlags for the flags received by the CLI
type ParseFlags struct {
	Project string
	Service string
	Remote  string
	Host    string
}

// Parse host and flags
func (c CommandFlagFromHost) Parse(pf ParseFlags) (*FlagsFromHost, error) {
	pf.Host = strings.ToLower(pf.Host)
	pf.Project = strings.ToLower(pf.Project)
	pf.Service = strings.ToLower(pf.Service)
	pf.Remote = strings.ToLower(pf.Remote)

	var flagsFromHost, err = c.parse(pf.Host, pf.Project, pf.Service, pf.Remote)

	if err != nil {
		return nil, err
	}

	if flagsFromHost.Service() != "" && flagsFromHost.Project() == "" {
		err = ErrorServiceWithNoProject{}
	}

	return flagsFromHost, err
}

// ParseFlagsWithDefaultCustomRemote is similar to ParseFlags,
// but with 'default custom remote' support
type ParseFlagsWithDefaultCustomRemote struct {
	Project       string
	Service       string
	Remote        string
	Host          string
	RemoteChanged bool
}

// ParseWithDefaultCustomRemote parses the flags using a custom default remote value
func (c CommandFlagFromHost) ParseWithDefaultCustomRemote(pf ParseFlagsWithDefaultCustomRemote, customRemote string) (*FlagsFromHost, error) {
	if !pf.RemoteChanged {
		pf.Remote = ""
	}

	var f, err = c.Parse(ParseFlags{
		Project: pf.Project,
		Service: pf.Service,
		Remote:  pf.Remote,
		Host:    pf.Host,
	})

	if f != nil {
		switch {
		case f.remote == "" && (pf.RemoteChanged || f.IsRemoteFromHost()):
			f.remote = defaults.CloudRemote
		case !f.IsRemoteFromHost() && !pf.RemoteChanged:
			f.remote = customRemote
		}
	}

	return f, err
}

func (c CommandFlagFromHost) parse(host, project, service, remote string) (*FlagsFromHost, error) {
	if host != "" {
		if project != "" || service != "" {
			return nil, ErrorMultiMode{}
		}

		return c.parseWithHost(host, remote)
	}

	if remote != "" {
		if _, ok := (*c.RemotesList)[remote]; !ok {
			return nil, ErrorNotFound{remote}
		}
	}

	return &FlagsFromHost{
		project: project,
		service: service,
		remote:  remote,
	}, nil
}

func (c CommandFlagFromHost) parseWithHost(host, remoteFromFlag string) (*FlagsFromHost, error) {
	remote, err := c.ParseRemoteAddress(host)

	if err == nil {
		if remote != "" && remoteFromFlag != "" {
			return nil, ErrorRemoteFlagAndHost{}
		}

		return &FlagsFromHost{
			remote:           remote,
			isRemoteFromHost: true,
		}, nil
	}

	switch err.(type) {
	case ErrorFoundMultipleRemote:
		return nil, err
	}

	flagsFromHost, err := c.parseHost(host)

	if err != nil {
		return nil, err
	}

	if flagsFromHost.remote != "" && remoteFromFlag != "" {
		return nil, ErrorRemoteFlagAndHost{}
	}

	if remoteFromFlag != "" {
		if _, ok := (*c.RemotesList)[remoteFromFlag]; !ok {
			return nil, ErrorNotFound{remoteFromFlag}
		}

		flagsFromHost.remote = remoteFromFlag
	}

	return flagsFromHost, err
}

func splitHyphenedHostPart(s string) []string {
	var p = strings.Index(s, "-")

	if p == -1 {
		return []string{s}
	}

	return []string{
		s[0:p],
		s[p+1:],
	}
}

func (c CommandFlagFromHost) parseHost(host string) (*FlagsFromHost, error) {
	var (
		parseDot    = strings.SplitN(host, ".", 2)
		parseHyphen = splitHyphenedHostPart(parseDot[0])
		project     string
		service     string
	)

	if len(parseDot) == 1 {
		if len(parseHyphen) == 1 {
			return &FlagsFromHost{
				service: parseHyphen[0],
			}, nil
		}

		return &FlagsFromHost{
			service: parseHyphen[0],
			project: parseHyphen[1],
		}, nil
	}

	switch len(parseHyphen) {
	case 1:
		project = parseHyphen[0]
	default:
		service = parseHyphen[0]
		project = parseHyphen[1]
	}

	return c.parseHostWithRemote(project, service, host, parseDot[1])
}

func (c CommandFlagFromHost) parseHostWithRemote(project, service, host, remoteHost string) (*FlagsFromHost, error) {
	var remote, err = c.ParseRemoteAddress(remoteHost)
	// notice that the logic above implies we MUST NOT
	// have a immediate foo.bar if bar is already a remote address
	// or it is going to have an ambiguity and always choose the longest host
	// testing for this edge case is either nice to have or overkill (undecided)
	// and using such a architecture would be problematic in many ways for
	// things such as DNS resolution, understanding which service is which, etc.

	// rewrite host if the error is "ErrorFoundNoRemote"
	switch err.(type) {
	case ErrorFoundNoRemote:
		err = ErrorFoundNoRemote{host}
	}

	if err != nil {
		return nil, err
	}

	var flagsFromHost = &FlagsFromHost{
		project: project,
		service: service,
		remote:  remote,
	}

	if remote != "" || strings.HasSuffix(host, "."+defaults.CloudRemote) {
		flagsFromHost.isRemoteFromHost = true
	}

	return flagsFromHost, nil
}

// ParseRemoteAddress to get related remote
func (c CommandFlagFromHost) ParseRemoteAddress(remoteAddress string) (remote string, err error) {
	if remoteAddress == "" {
		return "", nil
	}

	if c.RemotesList == nil {
		return "", ErrorLoadingRemoteList{}
	}

	switch found := c.parseRemoteAddress(remoteAddress); len(found) {
	case 0:
		return "", ErrorFoundNoRemote{remoteAddress}
	case 1:
		return found[0], nil
	default:
		return "", ErrorFoundMultipleRemote{remoteAddress, found}
	}
}

func (c CommandFlagFromHost) parseRemoteAddress(remoteAddress string) (found []string) {
	for _, k := range c.RemotesList.Keys() {
		var v = (*c.RemotesList)[k]

		var sameHTTP = matchStringWithPrefix("http://", remoteAddress, v.Service)
		var sameHTTPS = matchStringWithPrefix("https://", remoteAddress, v.Service)

		if sameHTTP || sameHTTPS || remoteAddress == v.Service {
			found = append(found, k)
		}
	}

	return found
}

func matchStringWithPrefix(prefix, ref, s string) bool {
	return strings.HasPrefix(s, prefix) &&
		ref == strings.TrimPrefix(s, prefix)
}
