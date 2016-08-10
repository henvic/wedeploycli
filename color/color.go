// Heavily modified version of
// https://github.com/fatih/color by Fatih Arslan (2013, MIT license)
// with minimal public interface:
// Format and Escape functions only

package color

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"golang.org/x/crypto/ssh/terminal"
)

// NoColor defines if the output is colorized or not.
// It is set based on the stdout's file descriptor by default.
var NoColor = !terminal.IsTerminal(int(os.Stdout.Fd()))

// Attribute defines a single SGR Code
type Attribute int

// Base attributes
const (
	Reset Attribute = iota
	Bold
	Faint
	Italic
	Underline
	BlinkSlow
	BlinkRapid
	ReverseVideo
	Concealed
	CrossedOut
)

// Foreground text colors
const (
	FgBlack Attribute = iota + 30
	FgRed
	FgGreen
	FgYellow
	FgBlue
	FgMagenta
	FgCyan
	FgWhite
)

// Foreground Hi-Intensity text colors
const (
	FgHiBlack Attribute = iota + 90
	FgHiRed
	FgHiGreen
	FgHiYellow
	FgHiBlue
	FgHiMagenta
	FgHiCyan
	FgHiWhite
)

// Background text colors
const (
	BgBlack Attribute = iota + 40
	BgRed
	BgGreen
	BgYellow
	BgBlue
	BgMagenta
	BgCyan
	BgWhite
)

// Background Hi-Intensity text colors
const (
	BgHiBlack Attribute = iota + 100
	BgHiRed
	BgHiGreen
	BgHiYellow
	BgHiBlue
	BgHiMagenta
	BgHiCyan
	BgHiWhite
)

const (
	escape   = "\x1b"
	unescape = "\\x1b"
)

// Format text for terminal
func Format(s ...interface{}) string {
	var out = make([]interface{}, 0)
	var params = []Attribute{}

	for _, v := range s {
		switch v.(type) {
		case []Attribute:
			params = append(params, v.([]Attribute)...)
		case Attribute:
			params = append(params, v.(Attribute))
		default:
			out = append(out, v)
		}
	}

	return wrap(params, sprintf(out...))
}

// Escape text for terminal
func Escape(s string) string {
	return strings.Replace(s, escape, unescape, -1)
}

func sprintf(s ...interface{}) string {
	switch len(s) {
	case 0:
		return ""
	case 1:
		return fmt.Sprintf("%v", s[0])
	}

	format := s[0]
	return fmt.Sprintf(fmt.Sprintf("%v", format), s[1:]...)
}

// sequence returns a formated SGR sequence to be plugged into a "\x1b[...m"
// an example output might be: "1;36" -> bold cyan.
func sequence(params []Attribute) string {
	format := make([]string, len(params))
	for i, v := range params {
		format[i] = strconv.Itoa(int(v))
	}

	return strings.Join(format, ";")
}

// wrap wraps the s string with the colors attributes.
func wrap(params []Attribute, s string) string {
	if NoColor {
		return s
	}

	return fmt.Sprintf("%s[%sm%s%s[%dm", escape, sequence(params), s, escape, Reset)
}
