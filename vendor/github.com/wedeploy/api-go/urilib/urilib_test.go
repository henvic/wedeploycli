package urilib

import "testing"

var dataProvider = []struct {
	paths    []string
	expected string
}{
	{
		[]string{"", "foo"},
		"foo",
	},
	{
		[]string{"", "foo/"},
		"foo",
	},
	{
		[]string{"foo", ""},
		"foo",
	},
	{
		[]string{"foo/", ""},
		"foo",
	},
	{
		[]string{"foo", "/bar"},
		"foo/bar",
	},
	{
		[]string{"foo", "bar"},
		"foo/bar",
	},
	{
		[]string{"foo/", "/bar"},
		"foo/bar",
	},
	{
		[]string{"foo/", "bar"},
		"foo/bar",
	},
	{
		[]string{"foo", "/bar", "bazz"},
		"foo/bar/bazz",
	},
	{
		[]string{"http://localhost:123", ""},
		"http://localhost:123",
	},
	{
		[]string{"http://localhost:123", "/foo", "bah"},
		"http://localhost:123/foo/bah",
	},
}

func TestResolvePath(t *testing.T) {
	for _, test := range dataProvider {
		got := ResolvePath(test.paths...)
		if got != test.expected {
			t.Errorf("For %q got %q; expected %q", test.paths, got, test.expected)
		}
	}
}
