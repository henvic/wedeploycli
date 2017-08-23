package remotes

import "sort"

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
