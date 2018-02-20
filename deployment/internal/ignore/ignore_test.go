package ignore

import (
	"os"
	"path/filepath"
	"testing"
)

var Matches = []string{
	".DS_Store",
	"/.DS_Store",
	"/foo/bar/.DS_Store",
	".swp",
	"/foo/bar/.swp",
	"/foo/bar/_swp",
	"/foo/bar/_swp",
	"/foo/bar/.kdev4",
}

var NoMatches = []string{
	"file",
	"swp.swp",
	"swp.swp",
	"/foo/DS_Store",
	"DS_Store", // missing "."
}

func TestMatch(t *testing.T) {
	for _, m := range Matches {
		if !Match(m) {
			t.Errorf("Expected %v to be matched.", m)
		}
	}

	for _, nm := range NoMatches {
		if Match(nm) {
			t.Errorf("Expected %v to not be matched.", nm)
		}
	}
}

type walker struct {
	b *testing.B
}

func (w *walker) walkFn(path string, info os.FileInfo, err error) error {
	if err != nil {
		return err
	}

	path, _ = filepath.Abs(path)
	_ = Match(path)

	return nil
}

func BenchmarkMatchRepository(b *testing.B) {
	// Remind that the tendency of any repo is to grow.
	// Therefore, comparing benchmark generated at different times
	// is futile. Instead, focus on comparing different algorithms against the same data.
	var w = walker{b}

	for i := 0; i < b.N; i++ {
		if err := filepath.Walk("../../", w.walkFn); err != nil {
			b.Errorf("Error walking: %v", err)
		}
	}
}
