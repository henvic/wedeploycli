package pod

import (
	"archive/tar"
	"compress/gzip"
	"crypto/md5"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"os"
	"testing"

	"github.com/launchpad-project/cli/progress"
)

type FileInfo struct {
	Name     string
	MD5      string
	Dir      bool
	Symlink  bool
	Linkname string
}

// Reference for files and directories on mocks/ref/
var Reference = map[string]*FileInfo{
	"dir/": {
		Name: "dir/",
		Dir:  true,
	},
	"symlink_dir": {
		Name:     "symlink_dir",
		Dir:      false,
		Symlink:  true,
		Linkname: "dir",
	},
	"dir/placeholder": {
		Name: "dir/placeholder",
		MD5:  "8996e816a871cf692f1063c93a18bd1b",
		Dir:  false,
	},
	"dir/symlink_placeholder": {
		Name:     "dir/symlink_placeholder",
		Dir:      false,
		Symlink:  true,
		Linkname: "placeholder",
	},
	"dir2/NotIgnored.md": {
		Name: "dir2/NotIgnored.md",
		MD5:  "788fe1134468d020cd3da5bce85eded1",
		Dir:  false,
	},
	"doc": {
		Name: "doc",
		MD5:  "ce1a09a0a60f13dfb8a718bd45e130fc",
		Dir:  false,
	},
	"ignored": {
		Name: "ignored",
		MD5:  "7b20a20f0c6ac0878c2a2018ee90a24c",
		Dir:  false,
	},
}

func TestPack(t *testing.T) {
	var ignoredList = []string{
		"ignored",
		"dir/foo/another_ignored_dir",
		"ignored_dir",
		"**/complex/*",
		"*Ignored*.md",
		"!NotIgnored.md",
	}

	// clean up package.tar.gz that might exist
	// to detect if it is not generated
	os.Remove("mocks/res/package.tar.gz")

	var size, err = Pack(
		"mocks/res/package.tar.gz",
		"mocks/ref",
		ignoredList,
		progress.New("mock"),
	)

	var minSize int64 = 5
	var maxSize int64 = 20

	if size <= minSize || size >= maxSize {
		t.Errorf("Expected size to be around %v-%v bytes, got %v instead",
			minSize,
			maxSize,
			size)
	}

	if err != nil {
		t.Errorf("Expected pack to end without errors, got %v error instead", err)
	}

	file, err := os.Open("mocks/res/package.tar.gz")

	if err != nil {
		t.Error(err)
	}

	gFile, err := gzip.NewReader(file)

	if err != nil {
		t.Error(err)
	}

	r := tar.NewReader(gFile)

	var found = map[string]*FileInfo{}

	for {
		f, err := r.Next()

		if err == io.EOF {
			break
		}

		if err != nil {
			t.Error(err)
		}

		rdbuf := make([]byte, 8)
		h := md5.New()

		if _, err := io.CopyBuffer(h, r, rdbuf); err != nil {
			t.Error(err)
		}

		found[f.Name] = &FileInfo{
			Name:     f.Name,
			MD5:      fmt.Sprintf("%x", h.Sum(nil)),
			Dir:      f.FileInfo().IsDir(),
			Symlink:  f.FileInfo().Mode()&os.ModeSymlink == os.ModeSymlink,
			Linkname: f.Linkname,
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
			t.Errorf("%v not found on tar", k)
		}

		if found[k].Dir != Reference[k].Dir {
			t.Errorf("Wanted Dir: %v for %v, got %v instead",
				Reference[k].Dir, k, found[k].Dir)
		}

		if !found[k].Dir && !found[k].Symlink && found[k].MD5 != Reference[k].MD5 {
			t.Errorf("Wanted MD5: %v for %v, got %v instead",
				Reference[k].MD5, k, found[k].MD5)
		}

		if found[k].Name != Reference[k].Name {
			t.Errorf("Wanted Name: %v for %v, got %v instead",
				Reference[k].Name, k, found[k].Name)
		}

		if found[k].Symlink != Reference[k].Symlink {
			t.Errorf("Wanted Symlink: %v for %v, got %v instead",
				Reference[k].Symlink, k, found[k].Symlink)
		}

		if found[k].Linkname != Reference[k].Linkname {
			t.Errorf("Wanted Linkname: %v for %v, got %v instead",
				Reference[k].Linkname, k, found[k].Linkname)
		}
	}

	for _, k := range refuteIgnored {
		if found[k] != nil {
			t.Errorf("Expected file %v to be ignored.", k)
		}
	}

	gFile.Close()
	file.Close()

	// clean up package.tar.gz to avoid false positives for other
	// tests that misses adding a detection step
	os.Remove("mocks/res/package.tar.gz")
}

func TestNotSelfPack(t *testing.T) {
	d := []byte("temporary placeholder")
	if err := ioutil.WriteFile("mocks/self/package.tar.gz", d, 0644); err != nil {
		panic(err)
	}

	var _, err = Pack(
		"mocks/self/package.tar.gz",
		"mocks/self",
		nil,
		progress.New("mock"),
	)

	file, err := os.Open("mocks/self/package.tar.gz")

	if err != nil {
		t.Error(err)
	}

	gFile, err := gzip.NewReader(file)

	if err != nil {
		t.Error(err)
	}

	r := tar.NewReader(gFile)

	var found = map[string]*FileInfo{}

	for {
		f, err := r.Next()

		if err == io.EOF {
			break
		}

		if err != nil {
			t.Error(err)
		}

		found[f.Name] = &FileInfo{
			Name:    f.Name,
			Dir:     f.FileInfo().IsDir(),
			Symlink: f.FileInfo().Mode()&os.ModeSymlink == os.ModeSymlink,
		}
	}

	if found["placeholder"] == nil {
		t.Errorf("Expected placeholder to be found")
	}

	if found["package.tar.gz"] != nil {
		t.Errorf("package.tar.gz should not be packed")
	}

	gFile.Close()
	file.Close()

	// clean up package.tar.gz to avoid false positives for other
	// tests that misses adding a detection step
	os.Remove("mocks/self/package.tar.gz")
}

func TestPackInvalidDestination(t *testing.T) {
	var invalid = fmt.Sprintf("mocks/res/invalid-dest-%d/foo.pod", rand.Int())
	var size, err = Pack(invalid, "mocks/ref", nil, progress.New("invalid"))

	if size != 0 {
		t.Errorf("Expected size to be zero on invalid destination")
	}

	if !os.IsNotExist(err) {
		t.Errorf("Wanted error te be due directory not found, got %v instead", err)
	}
}

func BenchmarkPack(b *testing.B) {
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

	// clean up any old package.tar.gz that might exist
	os.Remove("mocks/res/benchmark.tar.gz")

	var size, err = Pack(
		"mocks/res/benchmark.tar.gz",
		"mocks/benchmark",
		ignoredList,
		progress.New("mock"),
	)

	var minSize int64 = 14000000
	var maxSize int64 = 22000000

	if size <= minSize || size >= maxSize {
		b.Errorf("Expected size to be around %v-%v bytes, got %v instead",
			minSize,
			maxSize,
			size)
	}

	if err != nil {
		b.Errorf("Expected pack to end without errors, got %v error instead", err)
	}

	file, err := os.Open("mocks/res/benchmark.tar.gz")

	if err != nil {
		b.Error(err)
	}

	gFile, err := gzip.NewReader(file)

	if err != nil {
		b.Error(err)
	}

	tar.NewReader(gFile)
	gFile.Close()
	file.Close()

	// clean up any old package.tar.gz that might exist
	os.Remove("mocks/res/benchmark.tar.gz")
}
