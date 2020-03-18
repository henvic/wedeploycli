package prettyjson

import (
	"github.com/henvic/wedeploycli/color"
	"github.com/henvic/wedeploycli/isterm"
	"github.com/tidwall/pretty"
)

// Pretty prettifies JSON
func Pretty(b []byte) []byte {
	res := pretty.PrettyOptions(b, &pretty.Options{
		Width:  20,
		Indent: "    ",
	})

	if !color.NoColor && isterm.Stderr() && isterm.Stdout() {
		res = pretty.Color(res, nil)
	}

	return res
}
