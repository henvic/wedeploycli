package color

import (
	"reflect"
	"testing"
)

func TestFormat(t *testing.T) {
	var defaultNoColor = NoColor
	NoColor = false

	want := "\x1b[102;95mHello World\x1b[0m"
	got := Format(BgHiGreen, FgHiMagenta, "Hello World")

	if got != want {
		t.Errorf("Expecting %s, got '%s'\n", want, got)
	}

	NoColor = defaultNoColor
}

func TestFormatLong(t *testing.T) {
	var defaultNoColor = NoColor
	NoColor = false

	want := "\x1b[102;95mHello World, my little 92%!(EXTRA string=robot)\x1b[0m"
	got := Format(BgHiGreen, FgHiMagenta, "Hello World, %s %s %v", "my", "little", FgHiGreen, "robot")

	if got != want {
		t.Errorf("Expecting %s, got '%s'\n", want, got)
	}

	NoColor = defaultNoColor
}

func TestMalformedFormat(t *testing.T) {
	var defaultNoColor = NoColor
	NoColor = false

	want := "\x1b[102mHello World%!(EXTRA color.Attribute=95)\x1b[0m"
	got := Format(BgHiGreen, "Hello World", FgHiMagenta)

	if got != want {
		t.Errorf("Expecting %s, got '%s'\n", want, got)
	}

	NoColor = defaultNoColor
}

func TestFormatArray(t *testing.T) {
	var defaultNoColor = NoColor
	NoColor = false

	var format = []Attribute{BgHiGreen, FgHiMagenta}

	want := "\x1b[102;95mHello World\x1b[0m"
	got := Format(format, "Hello World")

	if got != want {
		t.Errorf("Expecting %s, got '%s'\n", want, got)
	}

	NoColor = defaultNoColor
}

func TestEmpty(t *testing.T) {
	var defaultNoColor = NoColor
	NoColor = false

	want := "\x1b[m\x1b[0m"
	got := Format()

	if got != want {
		t.Errorf("Expecting %s, got '%s'\n", want, got)
	}

	NoColor = defaultNoColor
}

func TestEmptyColorString(t *testing.T) {
	var defaultNoColor = NoColor
	NoColor = false

	want := "\x1b[40m\x1b[0m"
	got := Format(BgBlack)

	if got != want {
		t.Errorf("Expecting %s, got '%s'\n", want, got)
	}

	NoColor = defaultNoColor
}

func TestNoFormat(t *testing.T) {
	var defaultNoColor = NoColor
	NoColor = false

	want := "\x1b[mHello World\x1b[0m"
	got := Format("Hello World")

	if got != want {
		t.Errorf("Expecting %s, got '%s'\n", want, got)
	}

	NoColor = defaultNoColor
}

func TestFormatStartingWithNumber(t *testing.T) {
	var defaultNoColor = NoColor
	NoColor = false

	want := "\x1b[102;95m100 forks\x1b[0m"
	number := 100

	if reflect.TypeOf(number).String() != "int" {
		t.Errorf("Must be integer; not a similar like Attribute")
	}

	got := Format(BgHiGreen, FgHiMagenta, "%v forks", number)

	if got != want {
		t.Errorf("Expecting %s, got '%s'\n", want, got)
	}

	NoColor = defaultNoColor
}

func TestFormatAsSprintf(t *testing.T) {
	var defaultNoColor = NoColor
	NoColor = false

	want := "\x1b[102;95mHello World\x1b[0m"
	got := Format(BgHiGreen, FgHiMagenta, "%v", "Hello World")

	if got != want {
		t.Errorf("Expecting %s, got '%s'\n", want, got)
	}

	NoColor = defaultNoColor
}

func TestNoColor(t *testing.T) {
	var defaultNoColor = NoColor
	NoColor = true

	want := "Hello World"
	got := Format(BgHiGreen, FgHiMagenta, "Hello World")

	if got != want {
		t.Errorf("Expecting %s, got '%s'\n", want, got)
	}

	NoColor = defaultNoColor
}

func TestEscape(t *testing.T) {
	unescaped := "\x1b[32mGreen"
	escaped := "\\x1b[32mGreen"
	got := Escape(unescaped)

	if got != escaped {
		t.Errorf("Expecting %s, got '%s'\n", escaped, got)
	}
}
