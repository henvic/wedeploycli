package remotes

import (
	"net"
	"sort"
	"strings"
	"sync"
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
type List struct {
	entries map[string]Entry
	m       sync.RWMutex
}

// Keys of the remote list
func (l List) Keys() []string {
	l.m.RLock()
	defer l.m.RUnlock()
	var keys = make([]string, 0, len(l.entries))

	for k := range l.entries {
		keys = append(keys, k)
	}

	sort.Strings(keys)
	return keys
}

// Has checks if a remote exists
func (l *List) Has(name string) bool {
	l.m.RLock()
	defer l.m.RUnlock()
	_, ok := l.entries[name]
	return ok
}

// Get a remote
func (l *List) Get(name string) Entry {
	l.m.RLock()
	defer l.m.RUnlock()
	return l.entries[name]
}

// Set a remote
func (l *List) Set(name string, entry Entry) {
	l.m.Lock()
	defer l.m.Unlock()

	if l.entries == nil {
		l.entries = map[string]Entry{}
	}

	l.entries[name] = entry
}

// Del deletes a remote by name
func (l *List) Del(name string) {
	l.m.Lock()
	defer l.m.Unlock()
	delete(l.entries, name)
}
