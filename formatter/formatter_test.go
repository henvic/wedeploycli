package formatter

import "testing"

func TestPadding(t *testing.T) {
	t.Run("testPaddingMachineFriendly", testPaddingMachineFriendly)
	Human = true
	defer func() {
		Human = false
	}()
	t.Run("testPaddingHumanFriendly", testPaddingHumanFriendly)
}

func testPaddingMachineFriendly(t *testing.T) {
	if Human {
		t.Errorf("Expected padding to be machine-friendly by default")
	}

	if CondPad("dog", 10) != "\t" {
		t.Errorf("Expected conditional padding to be tab, got something else instead")
	}
}

type padProvider struct {
	word      string
	threshold int
	want      string
}

var padCases = []padProvider{
	padProvider{"dog", -1, " "},
	padProvider{"dog", 0, " "},
	padProvider{"cat", 1, " "},
	padProvider{"fox", 2, " "},
	padProvider{"rex", 3, " "},
	padProvider{"ted", 4, " "},
	padProvider{"cup", 5, "  "},
	padProvider{"mom", 6, "   "},
	padProvider{"pop", 7, "    "},
	padProvider{"token", 3, " "},
	padProvider{"crop", 10, "      "},
}

func testPaddingHumanFriendly(t *testing.T) {
	for _, c := range padCases {
		if s := CondPad(c.word, c.threshold); s != c.want {
			t.Errorf(`Expected conditional padding of "%v" to be "%v", got "%v" instead`, c.word, c.want, s)
		}
	}
}
