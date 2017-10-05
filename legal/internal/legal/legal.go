package legal

import (
	"fmt"
	"strings"
)

// FormatLicense gets a fmt.Sprint'ed %q and replaces
// single-quoted character literal safely escaped with Go Syntax
// and replaces the representation using backtick instead, in a safe manner
func FormatLicense(s string) string {
	s = strings.Replace(fmt.Sprintf("%q", s), "\\n", "\n", -1)
	s = strings.Replace(s, `\"`, `"`, -1)
	s = strings.Replace(s, "`", "` + \"`\" + `", -1)
	s = "`" + s[1:len(s)-1] + "`"
	return s
}
