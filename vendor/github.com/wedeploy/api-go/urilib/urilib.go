package urilib

import "strings"

// ResolvePath concatenates URI paths
func ResolvePath(paths ...string) string {
	var final []string
	for _, path := range paths {
		path = strings.TrimPrefix(path, "/")
		path = strings.TrimSuffix(path, "/")

		if len(path) != 0 {
			final = append(final, path)
		}
	}

	return strings.Join(final, "/")
}
