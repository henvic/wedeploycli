package legal

import (
	"testing"
)

type formatLicense struct {
	in  string
	out string
}

var formatLicenseTests = []formatLicense{
	formatLicense{
		"oi",
		"`oi`",
	},
	formatLicense{
		"\"oi\"",
		"`\"oi\"`",
	},
	formatLicense{
		"abc`hello\nworld`def",
		"`abc` + \"`\" + `hello\nworld` + \"`\" + `def`",
	},
}

func TestFormatLicense(t *testing.T) {
	for _, tt := range formatLicenseTests {
		s := FormatLicense(tt.in)

		if s != tt.out {
			t.Errorf("FormatLicense(%q) => %q, want %q", tt.in, s, tt.out)
		}
	}
}
