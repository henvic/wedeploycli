// +build !windows

package context

func setupOSRoot() {
	sysRoot = "/"
}

func isRootDelimiter(dir string) bool {
	return dir == sysRoot
}
