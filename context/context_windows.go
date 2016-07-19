// +build windows

package context

// isDriveLetter returns true if path is Windows drive letter (like "c:").
// (from Go's filepath)
func isDriveLetter(path string) bool {
	return len(path) == 3 && path[1] == ':' && path[2] == '\\'
}

func setupOSRoot() {
	// not necessary to set sysRoot
	// unless we want to delimit it
	// and we don't need to delimit to C:\
}

func isRootDelimiter(dir string) bool {
	if dir == sysRoot {
		return true
	}

	return isDriveLetter(dir)
}
