package formatter

import (
	"bytes"
	"fmt"
	"testing"
)

func TestPadding(t *testing.T) {
	t.Run("testMachineFriendlyFormat", testMachineFriendlyFormat)
	Human = true
	defer func() {
		Human = false
	}()
	t.Run("testHumanFriendlyFormat", testHumanFriendlyFormat)
}

func testMachineFriendlyFormat(t *testing.T) {
	if Human {
		t.Errorf("Expected padding to be machine-friendly by default")
	}

	var b bytes.Buffer

	var tw = NewTabWriter(&b)
	var text = "ABC\tDEFGH\tIJKLMNOPQRSTUVWXYZ\n012345\t6789\n"
	_, err := fmt.Fprint(tw, text)

	if err != nil {
		t.Errorf("Error printing: %v", err)
	}

	if err := tw.Flush(); err != nil {
		t.Errorf("Error flushing: %v", err)
	}

	var got = b.String()

	if got != text {
		t.Errorf(`Expected text to be original "%v", got "%v" instead`, text, got)
	}
}

func testHumanFriendlyFormat(t *testing.T) {
	if !Human {
		t.Errorf("Expected padding to be human-friendly")
	}

	var b bytes.Buffer

	var tw = NewTabWriter(&b)
	var text = "ABC\tDEFGH\tIJKLMNOPQRSTUVWXYZ\n012345\t6789\n"
	_, err := fmt.Fprint(tw, text)

	if err != nil {
		t.Errorf("Error printing: %v", err)
	}

	if err := tw.Flush(); err != nil {
		t.Errorf("Error flushing: %v", err)
	}

	var want = `ABC       DEFGH    IJKLMNOPQRSTUVWXYZ
012345    6789
`
	var got = b.String()

	if got != want {
		t.Errorf(`Expected text to be "%v", got "%v" instead`, want, got)
	}
}
