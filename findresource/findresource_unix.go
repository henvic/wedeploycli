// +build !windows

package findresource

func setupOSRoot() {
	sysRoot = "/"
}

func isRootDelimiter(dir string) bool {
	return dir == sysRoot
}
