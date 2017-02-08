package formatter

import "strings"

// Human tells if the default formatting strategy should be human or machine friendly
var Human = false

// CondPad is a conditional padding function
func CondPad(word string, threshold int) string {
	if !Human {
		return "\t"
	}

	var wl = len(word)
	var space = " "

	if threshold > wl {
		space = pad(threshold - wl + 1)
	}

	return space
}

func pad(space int) string {
	return strings.Join(make([]string, space), " ")
}
