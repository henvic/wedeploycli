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

type PackFiles map[string]*FileInfo

type TestPackProvider struct {
	MinSize        int64
	MaxSize        int64
	IgnoredList    []string
	NotIgnoredList []string
	RefuteIgnored  []string
}

func (*TestPackProvider) CheckSize(t *testing.T, size int64) {
	if size <= TestPackCase.MinSize || size >= TestPackCase.MaxSize {
		t.Errorf("Expected size to be around %v-%v bytes, got %v instead",
			TestPackCase.MinSize,
			TestPackCase.MaxSize,
			size)
	}
}

func (*TestPackProvider) CheckNotIgnored(t *testing.T, found PackFiles) {
	for _, k := range TestPackCase.NotIgnoredList {
		var f = found[k]
		var ref = Reference[k]

		if f == nil {
			t.Errorf("%v not found on tar", k)
		}

		if f.Dir != ref.Dir {
			t.Errorf("Wanted Dir: %v for %v, got %v instead",
				ref.Dir, k, f.Dir)
		}

		if !f.Dir && !f.Symlink && f.MD5 != ref.MD5 {
			t.Errorf("Wanted MD5: %v for %v, got %v instead",
				ref.MD5, k, f.MD5)
		}

		if f.Name != ref.Name {
			t.Errorf("Wanted Name: %v for %v, got %v instead",
				ref.Name, k, f.Name)
		}

		if f.Symlink != ref.Symlink {
			t.Errorf("Wanted Symlink: %v for %v, got %v instead",
				ref.Symlink, k, f.Symlink)
		}

		if f.Linkname != ref.Linkname {
			t.Errorf("Wanted Linkname: %v for %v, got %v instead",
				ref.Linkname, k, f.Linkname)
		}
	}
}

func (*TestPackProvider) CheckRefuteIgnored(t *testing.T, found PackFiles) {
	for _, k := range TestPackCase.RefuteIgnored {
		if found[k] != nil {
			t.Errorf("Expected file %v to be ignored.", k)
		}
	}
}

var TestPackCase = TestPackProvider{
	MinSize: 5,
	MaxSize: 20,
	IgnoredList: []string{
		"ignored",
		"dir/foo/another_ignored_dir",
		"ignored_dir",
		"**/complex/*",
		"*Ignored*.md",
		"!NotIgnored.md",
	},
	NotIgnoredList: []string{
		"doc",
		"dir/",
		"symlink_dir",
		"dir/placeholder",
		"dir/symlink_placeholder",
	},
	RefuteIgnored: []string{
		"ignored",
		"ignored_dir",
		"ignored_dir/placeholder",
		"dir/foo/another_ignored_dir",
		"dir/foo/another_ignored_dir/placeholder",
		"dir/sub/complex",
		"dir/sub/complex/placeholder",
		"dir/sub/complex/dir/placeholder",
	},
}

// Reference for files and directories on mocks/ref/
var Reference = PackFiles{
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
	// clean up package.tar.gz that might exist
	// to detect if it is not generated
	os.Remove("mocks/res/package.tar.gz")

	var size, err = Pack(
		"mocks/res/package.tar.gz",
		"mocks/ref",
		TestPackCase.IgnoredList,
		progress.New("mock"),
	)

	TestPackCase.CheckSize(t, size)

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

	var found = readPackFiles(t, tar.NewReader(gFile))
	TestPackCase.CheckNotIgnored(t, found)
	TestPackCase.CheckRefuteIgnored(t, found)

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

	var found = readPackFiles(t, tar.NewReader(gFile))

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

func readPackFiles(t *testing.T, r *tar.Reader) PackFiles {
	var found = PackFiles{}

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

	return found
}
