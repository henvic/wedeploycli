package pod

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/launchpad-project/cli/verbose"
	"github.com/sabhiram/go-git-ignore"
)

type pod struct {
	Source      string
	Writer      *zip.Writer
	ignoreRules *ignore.GitIgnore
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
func Compress(dest, src string, ignorePatterns []string) error {
	var file *os.File

	irules, err := ignore.CompileIgnoreLines(ignorePatterns...)

	if err == nil {
		file, err = os.Create(dest)
	}

	if err != nil {
		return err
	}

	verbose.Debug("Saving container to", file.Name())

	var pkg = &pod{
		Source:      src,
		Writer:      zip.NewWriter(file),
		ignoreRules: irules,
	}

	err = filepath.Walk(src, pkg.walkFunc)

	if err == nil {
		err = pkg.Writer.Close()
	}

	if err == nil {
		err = file.Close()
	}

	return err
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

	var header *zip.FileHeader
	header, err = zip.FileInfoHeader(fi)

	if err != nil {
		verbose.Debug("Can't retrieve file info for", path)
		return err
	}

	if p.ignoreRules.MatchesPath(relative) {
		if fi.IsDir() {
			return filepath.SkipDir
		}

		return nil
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
