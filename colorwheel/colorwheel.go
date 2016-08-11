package colorwheel

import "github.com/wedeploy/cli/color"

var TextPalette = [][]color.Attribute{
	[]color.Attribute{color.FgHiRed},
	[]color.Attribute{color.FgHiGreen},
	[]color.Attribute{color.FgHiYellow},
	[]color.Attribute{color.FgHiBlue},
	[]color.Attribute{color.FgHiMagenta},
}

var BlockPalette = [][]color.Attribute{
	[]color.Attribute{color.BgHiRed, color.FgBlack},
	[]color.Attribute{color.BgHiGreen, color.FgBlack},
	[]color.Attribute{color.BgHiYellow, color.FgBlack},
	[]color.Attribute{color.BgHiBlue, color.FgBlack},
	[]color.Attribute{color.BgHiMagenta, color.FgBlack},
}

type Wheel struct {
	palette [][]color.Attribute
	hm      map[string][]color.Attribute
	next    int
}

func New(palette [][]color.Attribute) Wheel {
	return Wheel{
		palette: palette,
	}
}

func (w *Wheel) Get(id string) []color.Attribute {
	if w.hm == nil {
		w.hm = map[string][]color.Attribute{}
	}

	var _, ok = w.hm[id]

	if ok {
		return w.hm[id]
	}

	w.hm[id] = w.palette[w.next]
	w.nextColor()

	return w.hm[id]
}

func (w *Wheel) nextColor() {
	if w.next == len(w.palette)-1 {
		w.next = 0
	} else {
		w.next++
	}
}
