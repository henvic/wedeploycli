package deployment

import (
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/hashicorp/errwrap"
	"github.com/henvic/wedeploycli/deployment/internal/ignore"
	"github.com/henvic/wedeploycli/verbose"
)

func (d *Deploy) copyServiceFiles(path string) (err error) {
	c := copyServiceFiles{
		deploy:      d,
		servicePath: path,
		copyPath:    filepath.Join(d.workDir, filepath.Base(path)),
	}

	if err = filepath.Walk(path, c.walkFn); err != nil {
		return err
	}

	verbose.Debug("Adding service " + path)
	return nil
}

type copyServiceFiles struct {
	deploy *Deploy

	servicePath string
	copyPath    string
}

func (c *copyServiceFiles) walkFn(path string, info os.FileInfo, ef error) (err error) {
	if ef != nil {
		return errwrap.Wrapf("can't read file "+path+" {err}}", ef)
	}

	if info.Name() == ".git" || ignore.Match(info.Name()) {
		if info.IsDir() {
			return filepath.SkipDir
		}

		return nil
	}

	if _, has := c.deploy.ignored[path]; has {
		return nil
	}

	var toTmp = filepath.Join(c.copyPath, strings.TrimPrefix(path, c.servicePath))
	var mode = info.Mode()

	if info.IsDir() {
		return os.MkdirAll(toTmp, mode)
	}

	from, openErr := os.Open(path) // #nosec
	to, createErr := os.OpenFile(toTmp, os.O_RDWR|os.O_CREATE, mode)

	if openErr != nil {
		return openErr
	}

	if createErr != nil {
		return createErr
	}

	_, err = io.Copy(to, from)
	eff := from.Close()
	eft := to.Close()

	if err != nil {
		return err
	}

	if eff != nil {
		return eff
	}

	return eft
}
