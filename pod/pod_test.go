package pod

import (
	"archive/zip"
	"fmt"
	"math/rand"
	"os"
	"reflect"
	"testing"

	"github.com/launchpad-project/cli/progress"
)

type FileInfo struct {
	Name    string
	CRC32   uint32
	Dir     bool
	Symlink bool
}

// Reference for files and directories on mocks/ref/
var Reference = map[string]*FileInfo{
	"dir/": &FileInfo{
		Name:  "dir/",
		CRC32: 0,
		Dir:   true,
	},
	"symlink_dir": &FileInfo{
		Name:    "symlink_dir",
		CRC32:   3131800080,
		Dir:     false,
		Symlink: true,
	},
	"dir/placeholder": &FileInfo{
		Name:  "dir/placeholder",
		CRC32: 258472330,
		Dir:   false,
	},
	"dir/symlink_placeholder": &FileInfo{
		Name:    "dir/symlink_placeholder",
		CRC32:   4125531906,
		Dir:     false,
		Symlink: true,
	},
	"dir2/NotIgnored.md": &FileInfo{
		Name:  "dir2/NotIgnored.md",
		CRC32: 2286427926,
		Dir:   false,
	},
	"doc": &FileInfo{
		Name:  "doc",
		CRC32: 3402152999,
		Dir:   false,
	},
	"ignored": &FileInfo{
		Name:  "ignored",
		CRC32: 763821341,
		Dir:   false,
	},
}

func TestCompress(t *testing.T) {
	var ignoredList = []string{
		"ignored",
		"dir/foo/another_ignored_dir",
		"ignored_dir",
		"**/complex/*",
		"*Ignored*.md",
		"!NotIgnored.md",
	}

	// clean up compress.zip that might exist
	// to detect if it is not generated
	os.Remove("mocks/res/compress.zip")

	var size, err = Compress(
		"mocks/res/compress.zip",
		"mocks/ref",
		ignoredList,
		progress.New("mock"),
	)

	var minSize int64 = 500
	var maxSize int64 = 2000

	if size <= minSize || size >= maxSize {
		t.Errorf("Expected size to be around %v-%v bytes, got %v instead",
			minSize,
			maxSize,
			size)
	}

	if err != nil {
		t.Errorf("Expected compress to end without errors, got %v error instead", err)
	}

	r, err := zip.OpenReader("mocks/res/compress.zip")

	if err != nil {
		t.Errorf("Wanted no errors opening compressed file, got %v instead", err)
	}

	var found = map[string]*FileInfo{}

	for _, f := range r.File {
		found[f.Name] = &FileInfo{
			Name:    f.Name,
			CRC32:   f.CRC32,
			Dir:     f.FileInfo().IsDir(),
			Symlink: f.Mode()&os.ModeSymlink == os.ModeSymlink,
		}
	}

	var assertNotIgnored = []string{
		"doc",
		"dir/",
		"symlink_dir",
		"dir/placeholder",
		"dir/symlink_placeholder",
	}

	var refuteIgnored = []string{
		"ignored",
		"ignored_dir",
		"ignored_dir/placeholder",
		"dir/foo/another_ignored_dir",
		"dir/foo/another_ignored_dir/placeholder",
		"dir/sub/complex",
		"dir/sub/complex/placeholder",
		"dir/sub/complex/dir/placeholder",
	}

	for _, k := range assertNotIgnored {
		if found[k] == nil {
			t.Errorf("%v not found on zip", k)
		}

		if !reflect.DeepEqual(found[k], Reference[k]) {
			t.Errorf("Zipped %v headers doesn't match expected values", k)
		}
	}

	for _, k := range refuteIgnored {
		if found[k] != nil {
			t.Errorf("Expected file %v to be ignored.", k)
		}
	}

	r.Close()

	// clean up compress.zip to avoid false positives for other
	// tests that misses adding a detection step
	os.Remove("mocks/res/compress.zip")
}

func TestCompressInvalidDestination(t *testing.T) {
	var invalid = fmt.Sprintf("mocks/res/invalid-dest-%d/foo.pod", rand.Int())
	var size, err = Compress(invalid, "mocks/ref", nil, progress.New("invalid"))

	if size != 0 {
		t.Errorf("Expected size to be zero on invalid destination")
	}

	if !os.IsNotExist(err) {
		t.Errorf("Wanted error te be due directory not found, got %v instead", err)
	}
}

func BenchmarkCompress(b *testing.B) {
	var ignoredList = []string{
		"arch",
		"arm",
		"blackfin",
		"drivers",
		"firmware",
		"fs",
		"include",
		"mips",
		"sound",
		"tools",
	}

	if _, err := os.Stat("mocks/benchmark/linux-4.6-rc2/"); os.IsNotExist(err) {
		b.Skip(`Test skipped due to missing test data. To install it run
pod/mocks/benchmark/install.sh`)
	}

	// clean up any old compress.zip that might exist
	os.Remove("mocks/res/benchmark.zip")

	var size, err = Compress(
		"mocks/res/benchmark.zip",
		"mocks/benchmark",
		ignoredList,
		progress.New("mock"),
	)

	var minSize int64 = 19000000
	var maxSize int64 = 26000000

	if size <= minSize || size >= maxSize {
		b.Errorf("Expected size to be around %v-%v bytes, got %v instead",
			minSize,
			maxSize,
			size)
	}

	if err != nil {
		b.Errorf("Expected compress to end without errors, got %v error instead", err)
	}

	r, err := zip.OpenReader("mocks/res/benchmark.zip")

	if err != nil {
		b.Errorf("Wanted no errors opening compressed file, got %v instead", err)
	}

	r.Close()

	// clean up any old compress.zip that might exist
	os.Remove("mocks/res/benchmark.zip")
}
