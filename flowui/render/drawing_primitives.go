// +build cgo

package render

import (
	"math"

	"github.com/gotk3/gotk3/cairo"
	"github.com/gotk3/gotk3/gtk"
)

func roundedRect(da *gtk.DrawingArea, cr *cairo.Context, x, y, w, h, r float64) {
	cr.NewPath()
	cr.Arc(x+w-r, y+r, r, -math.Pi/2, 0)
	cr.Arc(x+w-r, y+h-r, r, 0, math.Pi/2)
	cr.Arc(x+r, y+h-r, r, math.Pi/2, math.Pi)
	cr.Arc(x+r, y+r, r, -math.Pi, -math.Pi/2)
	cr.ClosePath()
}
