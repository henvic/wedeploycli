package stringlib

import (
	"strings"
	"testing"

	"github.com/kylelemons/godebug/diff"
)

// AssertSimilar strings by comparing its content after normalization
func AssertSimilar(t *testing.T, want string, got string) {
	if !Similar(want, got) {
		t.Errorf(
			"Strings doesn't match after normalization:\n%s",
			diff.Diff(Normalize(want), Normalize(got)))
	}
}

// Normalize string breaking lines with \n and removing extra spacing
// on the begining and end of strings
func Normalize(s string) string {
	var parts = strings.Split(s, "\n")
	var final = make([]string, 10*len(parts))

	var c = 0

	for p := range parts {
		var tp = strings.TrimSpace(parts[p])

		if tp != "" {
			final[c] = "\n"
			c++
		}

		final[c] = tp
		c++
	}

	return strings.TrimSpace(strings.Join(final, ""))
}

// Similar compares if two strings are similar after normalization
func Similar(x, y string) bool {
	return Normalize(x) == Normalize(y)
}
