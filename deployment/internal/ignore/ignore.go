package ignore

import (
	"path/filepath"
)

// These ignore rules were extracted and adapted from github.com/github/gitignore
// Though it contains lots of files, it isn't an exhaustive list.

// Match if a filename matches the ignore list
func Match(path string) bool {
	for _, p := range Patterns {
		// github.com/gobwas/glob has better performance, but
		// premature optimization is the root of all evil
		if ok, _ := filepath.Match(p, filepath.Base(path)); ok {
			return true
		}
	}

	return false
}
