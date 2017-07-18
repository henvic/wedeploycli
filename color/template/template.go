package colortemplate

import (
	"text/template"

	"github.com/wedeploy/cli/color"
)

var templateColorFunctions = template.FuncMap{
	"Reset":        func() color.Attribute { return color.Reset },
	"Bold":         func() color.Attribute { return color.Bold },
	"Faint":        func() color.Attribute { return color.Faint },
	"Italic":       func() color.Attribute { return color.Italic },
	"Underline":    func() color.Attribute { return color.Underline },
	"BlinkSlow":    func() color.Attribute { return color.BlinkSlow },
	"BlinkRapid":   func() color.Attribute { return color.BlinkRapid },
	"ReverseVideo": func() color.Attribute { return color.ReverseVideo },
	"Concealed":    func() color.Attribute { return color.Concealed },
	"CrossedOut":   func() color.Attribute { return color.CrossedOut },
	"FgBlack":      func() color.Attribute { return color.FgBlack },
	"FgRed":        func() color.Attribute { return color.FgRed },
	"FgGreen":      func() color.Attribute { return color.FgGreen },
	"FgYellow":     func() color.Attribute { return color.FgYellow },
	"FgBlue":       func() color.Attribute { return color.FgBlue },
	"FgMagenta":    func() color.Attribute { return color.FgMagenta },
	"FgCyan":       func() color.Attribute { return color.FgCyan },
	"FgWhite":      func() color.Attribute { return color.FgWhite },
	"FgHiBlack":    func() color.Attribute { return color.FgHiBlack },
	"FgHiRed":      func() color.Attribute { return color.FgHiRed },
	"FgHiGreen":    func() color.Attribute { return color.FgHiGreen },
	"FgHiYellow":   func() color.Attribute { return color.FgHiYellow },
	"FgHiBlue":     func() color.Attribute { return color.FgHiBlue },
	"FgHiMagenta":  func() color.Attribute { return color.FgHiMagenta },
	"FgHiCyan":     func() color.Attribute { return color.FgHiCyan },
	"FgHiWhite":    func() color.Attribute { return color.FgHiWhite },
	"BgBlack":      func() color.Attribute { return color.BgBlack },
	"BgRed":        func() color.Attribute { return color.BgRed },
	"BgGreen":      func() color.Attribute { return color.BgGreen },
	"BgYellow":     func() color.Attribute { return color.BgYellow },
	"BgBlue":       func() color.Attribute { return color.BgBlue },
	"BgMagenta":    func() color.Attribute { return color.BgMagenta },
	"BgCyan":       func() color.Attribute { return color.BgCyan },
	"BgWhite":      func() color.Attribute { return color.BgWhite },
	"BgHiBlack":    func() color.Attribute { return color.BgHiBlack },
	"BgHiRed":      func() color.Attribute { return color.BgHiRed },
	"BgHiGreen":    func() color.Attribute { return color.BgHiGreen },
	"BgHiYellow":   func() color.Attribute { return color.BgHiYellow },
	"BgHiBlue":     func() color.Attribute { return color.BgHiBlue },
	"BgHiMagenta":  func() color.Attribute { return color.BgHiMagenta },
	"BgHiCyan":     func() color.Attribute { return color.BgHiCyan },
	"BgHiWhite":    func() color.Attribute { return color.BgHiWhite },
	"color":        func(i ...interface{}) string { return color.Format(i...) },
}

// Functions lists the color functions
func Functions() template.FuncMap {
	return templateColorFunctions
}

// AddToTemplate adds color functions to a template
func AddToTemplate(t *template.Template) {
	t.Funcs(templateColorFunctions)
}
