package gogit

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/hashicorp/errwrap"
	"github.com/wedeploy/cli/deployment/internal/ignore"
	"github.com/wedeploy/cli/verbose"
	"gopkg.in/src-d/go-billy.v4/osfs"
	git "gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing/format/gitignore"
)

type file struct {
	path  string
	isDir bool
}

type ignoreChecker struct {
	path  string
	files []file
}

// ignoreChecker gets what file should be ignored.
func (i *ignoreChecker) Process() (map[string]struct{}, error) {
	var ps, err = gitignore.ReadPatterns(osfs.New(i.path), nil)

	if err != nil {
		return nil, errwrap.Wrapf("error processing .gitignore: {{err}}", err)
	}

	var files = []file{}

	if err = filepath.Walk(i.path, i.walkIgnored); err != nil {
		return nil, err
	}

	var ignored = map[string]struct{}{}

	for _, pattern := range ps {
		for _, f := range files {
			if _, is := ignored[f.path]; is {
				continue
			}

			status := pattern.Match(strings.Split(f.path, string(filepath.Separator)), f.isDir)

			if status == gitignore.Exclude {
				ignored[f.path] = struct{}{}
			}
		}
	}

	if len(ignored) != 0 {
		verbose.Debug(fmt.Sprintf(
			"Ignoring %d files and directories found on .gitignore files",
			len(ignored)))

	}

	return ignored, nil
}

func (i *ignoreChecker) walkIgnored(path string, info os.FileInfo, err error) error {
	if err != nil {
		return err
	}

	path = strings.TrimPrefix(path, i.path+string(os.PathSeparator))

	if info.Name() == git.GitDirName || ignore.Match(info.Name()) {
		if info.IsDir() {
			return filepath.SkipDir
		}

		return nil
	}

	i.files = append(i.files, file{
		path,
		info.IsDir(),
	})

	return nil
}
