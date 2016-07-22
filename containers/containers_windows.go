// +build windows

package containers

func normalizePath(s string) string {
	return normalizePathToUnix(s)
}
