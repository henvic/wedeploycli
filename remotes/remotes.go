package remotes

import (
	"sort"
	"strings"
)

// Entry for a remote
type Entry struct {
	URL        string
	URLComment string
	Comment    string
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
func (l List) Set(name string, url string, comment ...string) {
	// make sure to use # by default, instead of ;
	if len(comment) != 0 {
		comment = append([]string{"#"}, comment...)
	}

	l[name] = Entry{
		URL:     url,
		Comment: strings.Join(comment, " "),
	}
}

// Del deletes a remote by name
func (l List) Del(name string) {
	delete(l, name)
}
