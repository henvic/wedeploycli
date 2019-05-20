package deployment

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/wedeploy/cli/services"
	"github.com/wedeploy/cli/verbose"
)

func (d *Deploy) overwriteServicePackage(path string, content []byte) error {
	var overwrite = filepath.Join(d.workDir, path)
	var mode = os.FileMode(0644)

	if stat, err := os.Stat(overwrite); err == nil {
		mode = stat.Mode()
	}

	return ioutil.WriteFile(overwrite, content, mode)
}

func (d *Deploy) prepareAndModifyServicePackage(s services.ServiceInfo) error {
	// ignore service package contents because it is strict (see note below)
	var _, err = services.Read(s.Location)

	switch err {
	case nil:
	case services.ErrServiceNotFound:
		verbose.Debug(fmt.Sprintf(`LCP.json not found for service "%s"`, s.ServiceID))
		err = nil
	default:
		return err
	}

	// LCP.json is actively modified. Therefore, by using a map instead of relying on the
	// Package struct we avoid any issues regarding synchronization and we future-proof the structure.

	c := changes{
		ServiceID: s.ServiceID,
		Image:     d.Image,
	}

	bin, err := getPreparedServicePackage(c, s.Location)

	if err != nil {
		return err
	}

	path := filepath.Join(filepath.Base(s.Location), "LCP.json")

	return d.overwriteServicePackage(path, bin)
}
