package findresource

import (
	"os"
	"path/filepath"
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
func GetRootDirectory(delimiter, file string) (dir string, err error) {
	dir, err = os.Getwd()

	if err != nil {
		return "", err
	}

	stat, err := os.Stat(delimiter)

	if err != nil || !stat.IsDir() {
		return "", os.ErrNotExist
	}

	return walkToRootDirectory(dir, delimiter, file)
}

func walkToRootDirectory(dir, delimiter, file string) (string, error) {
	// sysRoot = / = upper-bound / The Power of Ten rule 2
	for !isRootDelimiter(dir) && dir != delimiter {
		stat, err := os.Stat(filepath.Join(dir, file))

		if stat == nil {
			dir = filepath.Join(dir, "..")
			continue
		}

		return dir, err
	}

	return "", os.ErrNotExist
}
