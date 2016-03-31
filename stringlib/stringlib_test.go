package stringlib

import (
	"testing"

	"github.com/kylelemons/godebug/diff"
)

type AssertSimilarProvider struct {
	In   string
	Out  string
	Pass bool
}

type TestStringProvider struct {
	In   string
	Want string
}

var TestNormalizeCases = []TestStringProvider{
	{
		In:   "\n\n Hi 1 2 3\n\n\n\n 4\n\r\n\n\n (xyz) \n\n\r\n\n\r\nHi\n\n",
		Want: "Hi 1 2 3\n4\n(xyz)\nHi",
	},
}

var AssertSimilarCases = []AssertSimilarProvider{
	{
		In:   "Hello World \n",
		Out:  "Hello World xyz",
		Pass: false,
	},
	{
		In:   "Hello World \n",
		Out:  "Hello World",
		Pass: true,
	},
	{
		In:   "Hello World \n",
		Out:  "Hello World",
		Pass: true,
	},
}

var TestSimilarCases = []TestStringProvider{
	{
		In:   "Hello World \n",
		Want: "Hello World",
	},
	{
		In:   "Hello World \n\n",
		Want: "Hello World",
	},
	{
		In:   "Hello World \n",
		Want: "Hello World",
	},
	{
		In:   "  \nHello  World \n",
		Want: " Hello  World",
	},
	{
		In:   "\n\n Hello World \n\n",
		Want: "Hello World",
	},
	{
		In:   "\n\n Hello World        \nHi",
		Want: "Hello World    \nHi",
	},
	{
		In:   "\n\n Hello World\n\n\n\r\r\n\n \n\r\n\n\n (xyz) \n\n\r\n\n\n\r\n\r\nHi\n\n",
		Want: "Hello World\n(xyz)\nHi",
	},
	{
		In:   "\n\n Hello World \nHow are you doing?",
		Want: "Hello World\nHow are you doing?",
	},
	{
		In:   "\n\n Hello World \nHow are you doing?\n\n\n",
		Want: "Hello World\nHow are you doing?",
	},
}

func TestAssertSimilar(t *testing.T) {
	for _, c := range AssertSimilarCases {
		var mockTest = &testing.T{}
		AssertSimilar(mockTest, c.In, c.Out)

		if mockTest.Failed() == c.Pass {
			t.Errorf("Mock test did not meet passing status = %v assertion", c.Pass)
		}
	}
}

func TestNormalize(t *testing.T) {
	for _, c := range TestNormalizeCases {
		var got = Normalize(c.In)

		if got != c.Want {
			t.Errorf("Wanted string %v, got %v instead", c.Want, got)
		}
	}
}

func TestSimilar(t *testing.T) {
	for _, c := range TestSimilarCases {
		if !Similar(c.In, c.Want) {
			t.Errorf(
				"Strings doesn't match after normalization:\n%s",
				diff.Diff(Normalize(c.Want), Normalize(c.In)))
		}
	}
}
