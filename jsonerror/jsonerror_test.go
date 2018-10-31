package jsonerror

import (
	"encoding/json"
	"testing"
)

// Mock for the test cases.
type Mock struct {
	ID           string            `json:"id,omitempty"`
	Scale        int               `json:"scale,omitempty"`
	Enabled      bool              `json:"enabled,omitempty"`
	Env          map[string]string `json:"env,omitempty"`
	Dependencies []string          `json:"dependencies,omitempty"`
}

type testcase struct {
	in   []byte
	want string
}

var cases = []testcase{
	testcase{
		in:   []byte(`{"id": "valid"}`),
		want: "",
	},
	testcase{
		in:   []byte(""),
		want: "unexpected end of JSON input",
	},
	testcase{
		in:   []byte(`true`),
		want: `json: cannot decode bool as jsonerror.Mock`,
	},
	testcase{
		in:   []byte(`false`),
		want: `json: cannot decode bool as jsonerror.Mock`,
	},
	testcase{
		in:   []byte(`1`),
		want: `json: cannot decode number as jsonerror.Mock`,
	},
	testcase{
		in:   []byte(`"x"`),
		want: `json: cannot decode string as jsonerror.Mock`,
	},
	testcase{
		in:   []byte(`""`),
		want: `json: cannot decode string as jsonerror.Mock`,
	},
	testcase{
		in:   []byte(`[]`),
		want: `json: cannot decode array as jsonerror.Mock`,
	},
	testcase{
		in:   []byte("1"),
		want: `json: cannot decode number as jsonerror.Mock`,
	},
	testcase{
		in: []byte(`{
			"id": "foo",
			"env": {
				"environment": "production",
				"size": 3
			}
		}`),
		want: `json: cannot decode number as Mock "env" string field`,
	},
	testcase{
		in:   []byte(`{"id": 400}`),
		want: `json: cannot decode number as Mock "id" string field`,
	},
	testcase{
		in:   []byte(`{"id": {"a": true}}`),
		want: `json: cannot decode object as Mock "id" string field`,
	},
	testcase{
		in:   []byte(`{"scale": "foo"}`),
		want: `json: cannot decode string as Mock "scale" int field`,
	},
	testcase{
		in:   []byte(`{"dependencies": "foo"}`),
		want: `json: cannot decode string as Mock "dependencies" []string field`,
	},
}

func TestFriendlyUnmarshal(t *testing.T) {
	for _, c := range cases {
		s := Mock{}
		err := FriendlyUnmarshal(json.Unmarshal(c.in, &s))

		if (err == nil) != (c.want == "") {
			t.Errorf("Expected error for %v to be %v, got %v instead", string(c.in), c.want, err)
		}

		if (err == nil && c.want != "") || (err != nil && err.Error() != c.want) {
			t.Errorf("Expected error to be %v, got %v instead", c.want, err)
		}
	}
}
