package findresource

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/hashicorp/errwrap"
)

var sysRoot string

func init() {
	setupOSRoot()
}

// SetSysRoot sets the delimiter to stop searching for a resource
func SetSysRoot(sr string) {
	sysRoot = sr
}

// GetSysRoot returns the delimiter to stop searching for a resource
func GetSysRoot() string {
	return sysRoot
}

// GetRootDirectory for a given file source
func GetRootDirectory(dir, delimiter, file string) (string, error) {
	stat, err := os.Stat(delimiter)

	if err != nil || !stat.IsDir() {
		return "", os.ErrNotExist
	}

	return walkToRootDirectory(dir, delimiter, file)
}

var errReachedDirectoryTreeRoot = errors.New("Reached directory tree root")

func walkToRootDirectory(dir, delimiter, file string) (string, error) {
	// sysRoot = / = upper-bound / The Power of Ten rule 2
	for !isRootDelimiter(dir) && dir != delimiter {
		_, err := os.Stat(filepath.Join(dir, file))

		switch {
		case os.IsNotExist(err):
			newDir := filepath.Join(dir, "..")

			if dir == newDir {
				return "", errReachedDirectoryTreeRoot
			}

			dir = newDir
			continue
		case err != nil:
			return "", errwrap.Wrapf("Error walking filesystem trying to find resouce "+file+": {{err}}", err)
		}

		return dir, err
	}

	return "", os.ErrNotExist
}
