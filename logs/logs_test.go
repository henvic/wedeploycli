package logs

import "testing"

type GetLevelProvider struct {
	in    string
	out   int
	valid bool
}

var GetLevelCases = []GetLevelProvider{
	{"0", 0, true},
	{"", 0, true},
	{"3", 3, true},
	{"critical", 2, true},
	{"error", 3, true},
	{"warning", 4, true},
	{"info", 6, true},
	{"debug", 7, true},
	{"foo", 0, false},
}

func TestGetLevel(t *testing.T) {
	for _, c := range GetLevelCases {
		out, err := GetLevel(c.in)
		valid := (c.valid == (err == nil))

		if out != c.out && valid {
			t.Errorf("Wanted level %v = (%v, valid: %v), got (%v, %v) instead",
				c.in,
				c.out,
				c.valid,
				out,
				err)
		}
	}
}
