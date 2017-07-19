// +build windows

package services

func normalizePath(s string) string {
	return normalizePathToUnix(s)
}
