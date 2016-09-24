package flagsfromhost

import (
	"fmt"
	"strings"

	"github.com/wedeploy/cli/remotes"
)

// ErrorRemoteFlagAndHost happens when both --remote and host remote are used together
type ErrorRemoteFlagAndHost struct{}

func (ErrorRemoteFlagAndHost) Error() string {
	return "Incompatible use: --remote flag can't be used along host format with remote address"
}

// ErrorMultiMode happens when --project and --container are used with host format
type ErrorMultiMode struct{}

func (ErrorMultiMode) Error() string {
	return "Incompatible use: --project and --container are not allowed with host format"
}

// ErrorContainerWithNoProject hapens when --container is used without --project
type ErrorContainerWithNoProject struct{}

func (ErrorContainerWithNoProject) Error() string {
	return "Incompatible use: --container requires --project"
}

// ErrorLoadingRemoteList happens when the remote list is needed, but not injected into the module
type ErrorLoadingRemoteList struct{}

func (ErrorLoadingRemoteList) Error() string {
	return "Error loading remotes list"
}

// ErrorFoundNoRemote happens when a remote isn't found for a given host address
type ErrorFoundNoRemote struct {
	Host string
}

func (e ErrorFoundNoRemote) Error() string {
	return fmt.Sprintf("Found no remote for address %v", e.Host)
}

// ErrorFoundNoRemote happens when a remote isn't found
type ErrorNotFound struct {
	Remote string
}

func (e ErrorNotFound) Error() string {
	return fmt.Sprintf("Remote %v not found", e.Remote)
}

// ErrorFoundMultipleRemote happens when multiple resolutions are found for a given host address
type ErrorFoundMultipleRemote struct {
	Host    string
	Remotes []string
}

func (e ErrorFoundMultipleRemote) Error() string {
	return fmt.Sprintf("Found multiple remotes for address %v: %v",
		e.Host,
		strings.Join(e.Remotes, ", "))
}

// FlagsFromHost holds the project, container, and remote parsed
type FlagsFromHost struct {
	project   string
	container string
	remote    string
}

// Project of the parsed flags or host
func (f *FlagsFromHost) Project() string {
	return f.project
}

// Container of the parsed flags or host
func (f *FlagsFromHost) Container() string {
	return f.container
}

// Remote of the parsed flags or host
func (f *FlagsFromHost) Remote() string {
	return f.remote
}

var remotesList *remotes.List
var localRemoteAddress = "wedeploy.me"

// InjectRemotes list into the flagsfromhost module
func InjectRemotes(list *remotes.List) {
	remotesList = list
}

// Parse host and flags
func Parse(host, project, container, remote string) (*FlagsFromHost, error) {
	var flagsFromHost, err = parse(host, project, container, remote)

	if err != nil {
		return nil, err
	}

	if container != "" && project == "" {
		return nil, ErrorContainerWithNoProject{}
	}

	return flagsFromHost, err
}

func parse(host, project, container, remote string) (*FlagsFromHost, error) {
	if host != "" {
		if project != "" || container != "" {
			return nil, ErrorMultiMode{}
		}

		return parseWithHost(host, remote)
	}

	return &FlagsFromHost{
		project:   project,
		container: container,
		remote:    remote,
	}, nil
}

func parseWithHost(host, remoteFromFlag string) (*FlagsFromHost, error) {
	var flagsFromHost, err = parseHost(host)

	if err != nil {
		return nil, err
	}

	if flagsFromHost.remote != "" && remoteFromFlag != "" {
		return nil, ErrorRemoteFlagAndHost{}
	}

	if remoteFromFlag != "" {
		if _, ok := (*remotesList)[remoteFromFlag]; !ok {
			return nil, ErrorNotFound{remoteFromFlag}
		}

		flagsFromHost.remote = remoteFromFlag
	}

	return flagsFromHost, nil
}

func parseHost(host string) (*FlagsFromHost, error) {
	var parseWithoutContainer = strings.SplitN(host, ".", 2)
	var parseWithContainer = strings.SplitN(host, ".", 3)

	switch len(parseWithContainer) {
	case 1:
		return &FlagsFromHost{
			project: parseWithContainer[0],
		}, nil
	case 2:
		return &FlagsFromHost{
			project:   parseWithContainer[1],
			container: parseWithContainer[0],
		}, nil
	}

	// a host "a.b.c.d" might translate into either
	// project: "a", container: "" (empty), remote address: "b.c.d"
	var (
		project     = parseWithContainer[0]
		container   = ""
		remote, err = ParseRemoteAddress(parseWithoutContainer[1])
	)

	// or project: "a", container: "b", remote address: "c.d"
	if err != nil {
		project = parseWithContainer[1]
		container = parseWithContainer[0]
		remote, err = ParseRemoteAddress(parseWithContainer[2])
	}

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

	return &FlagsFromHost{
		project:   project,
		container: container,
		remote:    remote,
	}, nil
}

// ParseRemoteAddress to get related remote
func ParseRemoteAddress(remoteAddress string) (remote string, err error) {
	if remoteAddress == "" || remoteAddress == localRemoteAddress {
		return "", nil
	}

	if remotesList == nil {
		return "", ErrorLoadingRemoteList{}
	}

	switch found := parseRemoteAddress(remoteAddress); len(found) {
	case 0:
		return "", ErrorFoundNoRemote{remoteAddress}
	case 1:
		return found[0], nil
	default:
		return "", ErrorFoundMultipleRemote{remoteAddress, found}
	}
}

func parseRemoteAddress(remoteAddress string) (found []string) {
	for _, k := range remotesList.Keys() {
		var v = (*remotesList)[k]

		var sameHTTP = matchStringWithPrefix("http://", remoteAddress, v.URL)
		var sameHTTPS = matchStringWithPrefix("http://", remoteAddress, v.URL)

		if sameHTTP || sameHTTPS || remoteAddress == v.URL {
			found = append(found, k)
		}
	}

	return found
}

func matchStringWithPrefix(prefix, ref, s string) bool {
	return strings.HasPrefix(s, prefix) &&
		ref == strings.TrimPrefix(s, prefix)
}
