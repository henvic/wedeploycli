package deployment

import (
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/hashicorp/errwrap"
)

func (d *Deploy) copyServiceFiles(path string) (copyPath string, err error) {
	c := copyServiceFiles{
		deploy:      d,
		servicePath: path,
		copyPath:    filepath.Join(d.tmpWorkDir, filepath.Base(path)),
	}

	if err = filepath.Walk(path, c.walkFn); err != nil {
		return "", err
	}

	return c.copyPath, nil
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

	if info.Name() == ".git" {
		return filepath.SkipDir
	}

	if _, has := c.deploy.ignoreList[path]; has {
		return nil
	}

	var toTmp = filepath.Join(c.copyPath, strings.TrimPrefix(path, c.servicePath))
	var mode = info.Mode()

	if info.IsDir() {
		return os.MkdirAll(toTmp, mode)
	}

	from, openErr := os.Open(path)
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
