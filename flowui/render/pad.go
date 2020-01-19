package render

import (
	"math"

	"github.com/gotk3/gotk3/cairo"
	"github.com/gotk3/gotk3/gtk"
)

// ColoredPad describes a pad with a custom color.
type ColoredPad interface {
	PadColor() (float64, float64, float64)
}

func (renderer *BasicRenderer) DrawPad(da *gtk.DrawingArea, cr *cairo.Context, animStep int64, p Pad) {
	var (
		pad             = p.Pad()
		x, y    float64 = p.Pos()
		dia, _  float64 = pad.Size()
		focused         = renderer.isFocused(p)
		r, g, b float64 = 0.5, 0.5, 0.5
	)

	if cp, hasColor := pad.(ColoredPad); hasColor {
		r, g, b = cp.PadColor()
	}

	if focused {
		r *= 1.3
		g *= 1.3
		b *= 1.3
	}

	cr.SetSourceRGB(r, g, b)
	cr.NewPath()
	cr.Arc(x, y, dia/2-1, -math.Pi, math.Pi)
	cr.ClosePath()
	cr.SetLineWidth(2)
	cr.MoveTo(x, y)
	cr.Arc(x, y, dia/4-1, -math.Pi, math.Pi)
	cr.ClosePath()
	cr.Fill()

	if focused {
		cr.SetLineWidth(2)
		cr.SetDash([]float64{4, 4}, float64(-(animStep >> 15)))
		cr.NewPath()
		cr.Arc(x, y, dia/2+dia/10, -math.Pi, math.Pi)
		cr.ClosePath()
		cr.Stroke()
		cr.SetDash(nil, 0)
	}
}
