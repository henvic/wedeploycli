package pod

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/dustin/go-humanize"
	"github.com/launchpad-project/cli/progress"
	"github.com/launchpad-project/cli/verbose"
	"github.com/sabhiram/go-git-ignore"
)

type pod struct {
	Source           string
	Writer           *zip.Writer
	NumberFiles      int
	NumberDirs       int
	NumberPathsZip   int
	UncompressedSize int64
	ignoreRules      *ignore.GitIgnore
	progress         *progress.Bar
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

// Compress pod
func Compress(dest,
	src string,
	ignorePatterns []string,
	pb *progress.Bar) (int64, error) {
	var file *os.File
	var size int64

	irules, err := ignore.CompileIgnoreLines(ignorePatterns...)

	if err == nil {
		file, err = os.Create(dest)
	}

	if err != nil {
		return 0, err
	}

	verbose.Debug("Saving container to", file.Name())

	var pkg = &pod{
		Source:      src,
		Writer:      zip.NewWriter(file),
		ignoreRules: irules,
		progress:    pb,
	}

	// Add 'counting files' progress bar (ya know, user feedback is important)
	pb.Reset("Counting files", "")
	err = filepath.Walk(src, pkg.countWalkFunc)

	if err != nil {
		return 0, err
	}

	pb.Reset("Zipping", "")

	err = filepath.Walk(src, pkg.walkFunc)

	if err != nil {
		return 0, err
	}

	pb.Set(progress.Total)
	pb.Append = "(Complete)"
	err = pkg.Writer.Close()

	if err != nil {
		return 0, err
	}

	var stat os.FileInfo

	stat, err = file.Stat()

	if err != nil {
		return 0, err
	}

	err = file.Close()

	if stat != nil {
		size = stat.Size()
	}

	return size, err
}

func (p *pod) countWalkFunc(path string, fi os.FileInfo, ierr error) error {
	if ierr != nil {
		verbose.Debug("Error reading", path, "when counting")
		return ierr
	}

	p.progress.Flow()

	var relative, err = filepath.Rel(p.Source, path)

	if err != nil {
		return err
	}

	// Pod, Jar is a .tar bomb, err... a .zip bomb!
	if relative == "." {
		return nil
	}

	if p.ignoreRules.MatchesPath(relative) {
		if fi.IsDir() {
			return filepath.SkipDir
		}

		return nil
	}

	if fi.IsDir() {
		p.NumberDirs++
	} else {
		p.NumberFiles++
		p.UncompressedSize += fi.Size()
	}

	p.progress.Append = fmt.Sprintf(`%s (%d dirs, %d files)`,
		humanize.Bytes(uint64(p.UncompressedSize)),
		p.NumberDirs,
		p.NumberFiles,
	)

	return nil
}

func (p *pod) walkFunc(path string, fi os.FileInfo, ierr error) error {
	if ierr != nil {
		verbose.Debug("Error reading", path)
		return ierr
	}

	var relative, err = filepath.Rel(p.Source, path)

	if err != nil {
		return err
	}

	// Pod, Jar is a .tar bomb, err... a .zip bomb!
	if relative == "." {
		return nil
	}

	if p.ignoreRules.MatchesPath(relative) {
		if fi.IsDir() {
			return filepath.SkipDir
		}

		return nil
	}

	p.NumberPathsZip++

	var totalPaths = p.NumberDirs + p.NumberFiles
	var perc = int(progress.Total * p.NumberPathsZip / totalPaths)

	p.progress.Set(perc)
	p.progress.Append = fmt.Sprintf(
		"%d/%d %v",
		p.NumberPathsZip,
		totalPaths,
		miniPath(relative))

	var header *zip.FileHeader
	header, err = zip.FileInfoHeader(fi)

	if err != nil {
		verbose.Debug("Can't retrieve file info for", path)
		return err
	}

	header.Name = relative

	if fi.IsDir() {
		header.Name += "/"
	} else {
		header.Method = zip.Deflate
	}

	var writer io.Writer
	writer, err = p.Writer.CreateHeader(header)

	if err != nil {
		verbose.Debug("Failure to create zip header for", path)
		return err
	}

	if fi.IsDir() {
		return nil
	}

	return copy(writer, path, relative)
}

func miniPath(s string) string {
	if len(s) <= 30 {
		return s
	}

	return "..." + s[len(s)-22:len(s)]
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
