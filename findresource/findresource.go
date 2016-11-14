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
	switch stat, err := os.Stat(delimiter); {
	case os.IsNotExist(err) || !stat.IsDir():
		return "", os.ErrNotExist
	case err != nil:
		return "", errwrap.Wrapf("Can't read delimiter "+delimiter+"directory: {{err}}", err)
	}

	return walkToRootDirectory(dir, delimiter, file)
}

var errReachedDirectoryTreeRoot = errors.New("Reached directory tree root")

func walkToRootDirectory(dir, delimiter, file string) (string, error) {
	// sysRoot = / = upper-bound / The Power of Ten rule 2
	for {
		stat, err := os.Stat(dir)

		switch {
		case os.IsNotExist(err) || !stat.IsDir():
			return "", os.ErrNotExist
		case err != nil:
			return "", errwrap.Wrapf("Error reading directory "+dir+": {{err}}", err)
		}

		switch _, err := os.Stat(filepath.Join(dir, file)); {
		case os.IsNotExist(err):
			if dir == delimiter {
				return "", os.ErrNotExist
			}

			newDir := filepath.Join(dir, "..")

			if dir == newDir {
				return "", errReachedDirectoryTreeRoot
			}

			dir = newDir

			if !isRootDelimiter(dir) && dir != delimiter {
				continue
			}

			return "", os.ErrNotExist
		case err != nil:
			return "", errwrap.Wrapf("Error walking filesystem trying to find resouce "+file+": {{err}}", err)
		}

		return dir, err
	}
}
