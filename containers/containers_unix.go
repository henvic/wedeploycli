// +build !windows

package containers

// Unix paths are already normalized by its very nature
func normalizePath(s string) string {
	return s
}
