package pod

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/dustin/go-humanize"
	"github.com/sabhiram/go-git-ignore"
	"github.com/wedeploy/cli/progress"
	"github.com/wedeploy/cli/verbose"
)

type pack struct {
	File       *os.File
	GzipWriter *gzip.Writer
	TarWriter  *tar.Writer
}

type pod struct {
	Dest               string
	Source             string
	NumberFiles        int
	NumberDirs         int
	NumberPathsPackage int
	PackageSize        int64
	pack               *pack
	ignoreRules        *ignore.GitIgnore
	progress           *progress.Bar
}

// CommonIgnorePatterns is a list of useful ignored lines
var CommonIgnorePatterns = []string{
	".DS_Store",
	".directory",
	".Trashes",
	".project",
	".settings",
	".idea",
	"*.pod",
}

// PackParams are package params for the packing process
type PackParams struct {
	RelDest        string
	RelSource      string
	IgnorePatterns []string
}

// Pack pod
func Pack(pp PackParams, pb *progress.Bar) (size int64, err error) {
	var pkg = pod{
		progress: pb,
	}

	err = pkg.start(pp.RelDest, pp.RelSource, pp.IgnorePatterns)

	if err != nil {
		return 0, err
	}

	return pkg.do()
}

func (p *pod) countWalkFunc(path string, fi os.FileInfo, ierr error) error {
	if ierr != nil {
		verbose.Debug("Error reading", path, "when counting")
		return ierr
	}

	p.progress.Flow()

	var relative, err = filepath.Rel(p.Source, path)
	var abs string

	if err == nil {
		abs, err = filepath.Abs(path)
	}

	if err != nil {
		return err
	}

	var wi, wiErr = p.testWalkIgnore(fi, relative, abs)

	if wi {
		return wiErr
	}

	if fi.IsDir() {
		p.NumberDirs++
	} else {
		p.NumberFiles++
		p.PackageSize += fi.Size()
	}

	p.progress.Append = fmt.Sprintf(`%s (%d dirs, %d files)`,
		humanize.Bytes(uint64(p.PackageSize)),
		p.NumberDirs,
		p.NumberFiles,
	)

	return nil
}

func (p *pod) createPack(dest string) error {
	var file, err = os.Create(dest)

	if err != nil {
		return err
	}

	var gFile = gzip.NewWriter(file)

	p.pack = &pack{
		File:       file,
		GzipWriter: gFile,
		TarWriter:  tar.NewWriter(gFile),
	}

	return err
}

func (p *pod) do() (size int64, err error) {
	err = p.runPacking()

	if err != nil {
		return 0, err
	}

	p.progress.Set(progress.Total)
	p.progress.Append = "(Complete)"

	size, err = p.pack.getSize()

	if err != nil {
		return 0, err
	}

	err = p.pack.Close()

	if err != nil {
		return 0, err
	}

	verbose.Debug("Saving container to", p.Dest)
	return size, err
}

func (p *pod) loadIgnorePatterns(ignorePatterns []string) error {
	var rules, err = ignore.CompileIgnoreLines(ignorePatterns...)
	p.ignoreRules = rules
	return err
}

func (p *pod) runPacking() (err error) {
	// Add 'counting files' progress bar (ya know, user feedback is important)
	p.progress.Reset("Counting files", "")
	err = p.walkSource(p.countWalkFunc)

	if err != nil {
		return err
	}

	p.progress.Reset("Packing", "")
	err = p.walkSource(p.walkFunc)

	return err
}

func (p *pod) setAbsPaths(dest, src string) (err error) {
	p.Dest, err = filepath.Abs(dest)

	if err != nil {
		return err
	}

	p.Source, err = filepath.Abs(src)
	return err
}

func (p *pod) start(dest, src string, ignorePatterns []string) error {
	var err = p.loadIgnorePatterns(ignorePatterns)

	if err != nil {
		return err
	}

	err = p.setAbsPaths(dest, src)

	if err != nil {
		return err
	}

	err = p.createPack(dest)

	return err
}

func (p *pod) walkSource(
	f func(path string, fi os.FileInfo, ierr error) error) error {
	return filepath.Walk(p.Source, f)
}

func (p *pod) walkFunc(path string, fi os.FileInfo, ierr error) error {
	if ierr != nil {
		verbose.Debug("Error reading", path)
		return ierr
	}

	var relative, err = filepath.Rel(p.Source, path)
	var abs string

	if err == nil {
		abs, err = filepath.Abs(path)
	}

	if err != nil {
		return err
	}

	var wi, wiErr = p.testWalkIgnore(fi, relative, abs)

	if wi {
		return wiErr
	}

	p.NumberPathsPackage++

	var totalPaths = p.NumberDirs + p.NumberFiles
	var perc = int(progress.Total * p.NumberPathsPackage / totalPaths)

	p.progress.Set(perc)
	p.progress.Append = fmt.Sprintf(
		"%d/%d %v",
		p.NumberPathsPackage,
		totalPaths,
		miniPath(relative))

	var header *tar.Header
	header, err = tar.FileInfoHeader(fi, relative)

	if err != nil {
		verbose.Debug("Can't retrieve file info for", path)
		return err
	}

	header.Name = relative

	if fi.IsDir() {
		header.Name += "/"
	}

	if fi.Mode()&os.ModeSymlink == os.ModeSymlink {
		var linkDest string

		linkDest, err = os.Readlink(path)

		if err != nil {
			return err
		}

		header.Linkname = linkDest
	}

	err = p.pack.TarWriter.WriteHeader(header)

	if err != nil {
		verbose.Debug("Failure to create package header for", path)
		return err
	}

	if fi.IsDir() || fi.Mode()&os.ModeSymlink == os.ModeSymlink {
		return nil
	}

	return copy(p.pack.TarWriter, path, relative)
}

func (p *pod) testWalkIgnore(fi os.FileInfo, relative, abs string) (
	ignore bool, err error) {
	// Pod, Jar is a gzipped tar bomb!
	// avoid packing itself 'til starvation also
	if relative == "." || abs == p.Dest {
		return true, nil
	}

	if p.ignoreRules.MatchesPath(relative) {
		if fi.IsDir() {
			return true, filepath.SkipDir
		}

		return true, nil
	}

	return false, nil
}

func miniPath(s string) string {
	if len(s) <= 30 {
		return s
	}

	return "..." + s[len(s)-22:]
}

func verboseCopyInfo(relative string, file *os.File) {
	if verbose.Enabled {
		stat, _ := file.Stat()
		verbose.Debug(fmt.Sprintf("%v (%v bytes)", relative, stat.Size()))
	}
}

func copy(writer io.Writer, path, relative string) error {
	var file, err = os.Open(path)

	if err == nil {
		verboseCopyInfo(relative, file)
		_, err = io.Copy(writer, file)
	}

	if err == nil {
		err = file.Close()
	}

	return err
}

func (pack *pack) Close() error {
	var err = pack.TarWriter.Close()

	if err != nil {
		return err
	}

	err = pack.GzipWriter.Close()

	if err != nil {
		return err
	}

	return pack.File.Close()
}

func (pack *pack) getSize() (int64, error) {
	var fi, err = pack.File.Stat()

	if err != nil {
		return 0, err
	}

	return fi.Size(), err
}
