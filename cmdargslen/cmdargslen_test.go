package cmdargslen

import "testing"

type entry struct {
	args []string
	min  int
	max  int
	err  error
}

var cases = []entry{
	entry{[]string{}, 0, 0, nil},
	entry{[]string{"foo"}, 0, 1, nil},
	entry{[]string{"foo"}, 1, 1, nil},
	entry{[]string{"foo", "bar"}, 1, 2, nil},
	entry{[]string{"foo", "bar"}, 2, 4, nil},
	entry{[]string{"foo", "bar", "var"}, 2, 4, nil},
	entry{[]string{"foo", "bar"}, 5, 5, errTooFew},
	entry{[]string{"foo", "bar"}, 0, 1, errTooMany},
}

func TestValidate(t *testing.T) {
	for _, c := range cases {
		var gotErr = Validate(c.args, c.min, c.max)

		if gotErr != c.err {
			t.Errorf("Wanted Validate(%v, %v, %v) = %v, got %v instead",
				c.args,
				c.min,
				c.max,
				c.err,
				gotErr)
		}
	}
}

func TestValidateCmd(t *testing.T) {
	if err := ValidateCmd(1, 2)(nil, []string{"foo"}); err != nil {
		t.Errorf("Expected no error, got %v instead", err)
	}
}
