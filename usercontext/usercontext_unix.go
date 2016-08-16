// +build !windows

package usercontext

func setupOSRoot() {
	sysRoot = "/"
}

func isRootDelimiter(dir string) bool {
	return dir == sysRoot
}
