package remotes

import (
	"net"
	"sort"
	"strings"
)

// Entry for a remote
type Entry struct {
	Infrastructure        string
	InfrastructureComment string
	ServiceComment        string
	Service               string
	Username              string
	UsernameComment       string
	Token                 string
	TokenComment          string
	Comment               string
}

// InfrastructureServer to connect with
func (e *Entry) InfrastructureServer() string {
	if !isHTTPLocalhost(e.Infrastructure) {
		return "https://api." + e.Infrastructure
	}

	return e.Infrastructure
}

func isHTTPLocalhost(address string) bool {
	address = strings.TrimPrefix(address, "http://")
	var h, _, err = net.SplitHostPort(address)

	if err != nil {
		return false
	}

	return h == "localhost"
}

// List of remotes
type List map[string]Entry

// Keys of the remote list
func (l List) Keys() []string {
	var keys = make([]string, 0, len(l))

	for k := range l {
		keys = append(keys, k)
	}

	sort.Strings(keys)
	return keys
}

// Set a remote
func (l List) Set(name string, entry Entry) {
	l[name] = entry
}

// Del deletes a remote by name
func (l List) Del(name string) {
	delete(l, name)
}
