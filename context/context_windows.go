// +build windows

package context

import (
	"os"
	"path/filepath"
)

// isDriveLetter returns true if path is Windows drive letter (like "c:").
// (from Go's filepath)
func isDriveLetter(path string) bool {
	return len(path) == 3 && path[1] == ':' && path[2] == '\\'
}

func setupOSRoot() {
	// we need to delimit to <drive>:\
	var cwd, err = os.Getwd()

	if err != nil {
		panic(err)
	}

	var drive = cwd

	for !isDriveLetter(drive) {
		drive = filepath.Join(drive, "..")
	}

	sysRoot = drive
}

func isRootDelimiter(dir string) bool {
	if dir == sysRoot {
		return true
	}

	return isDriveLetter(dir)
}
