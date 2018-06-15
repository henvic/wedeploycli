package prettyjson

import (
	"github.com/tidwall/pretty"
	"github.com/wedeploy/cli/color"
	"github.com/wedeploy/cli/isterm"
)

// Pretty prettifies JSON
func Pretty(b []byte) []byte {
	res := pretty.PrettyOptions(b, &pretty.Options{
		Width:  20,
		Indent: "    ",
	})

	if !color.NoColor && isterm.Check() {
		res = pretty.Color(res, nil)
	}

	return res
}
