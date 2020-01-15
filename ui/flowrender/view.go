package flowrender

import (
	"math"

	"github.com/gotk3/gotk3/cairo"
	"github.com/gotk3/gotk3/gtk"
	"github.com/twitchyliquid64/diagg/flow"
)

// headlineElement types are elements which have text which should be rendered
// on  them.
type headlineElement interface {
	NodeHeadline() string
}

// focusableElement types are elements which can be focused, and are drawn with
// a thicker outline if they are currently focused.
type focusableElement interface {
	Active() bool
}

type Node interface {
	Pos() (float64, float64)
	Node() flow.Node
}

type Pad interface {
	Pos() (float64, float64)
	Pad() flow.Pad
}

// Appearance represents an implementation which can display a flowchart.
type Appearance interface {
	DrawNode(da *gtk.DrawingArea, cr *cairo.Context, animStep float64, n Node)
	DrawPad(da *gtk.DrawingArea, cr *cairo.Context, animStep float64, p Pad)
}

type BasicRenderer struct{}

func (r *BasicRenderer) isFocused(n interface{}) bool {
	if fe, ok := n.(focusableElement); ok {
		return fe.Active()
	}
	return false
}

func (r *BasicRenderer) DrawNode(da *gtk.DrawingArea, cr *cairo.Context, animStep float64, n Node) {
	var (
		node                     = n.Node()
		x, y             float64 = n.Pos()
		w, h             float64 = node.Size()
		hw, hh           float64 = w / 2, h / 2
		sub, borderWidth float64 = 2, 2
	)
	if r.isFocused(n) {
		borderWidth = 6
	}

	cr.SetSourceRGB(1, 1, 1)
	cr.SetLineWidth(borderWidth)
	roundedRect(da, cr, x-hw, y-hh, w-sub, h-sub, 2)
	cr.StrokePreserve()
	cr.SetSourceRGB(0.5, 0.1, 0.1)
	cr.Fill()

	if hln, ok := node.(headlineElement); ok {
		cr.MoveTo(x-hw+7, y-hh+18)
		cr.SetSourceRGB(1, 1, 1)
		cr.SetFontSize(16)
		cr.ShowText(hln.NodeHeadline())
		cr.Fill()
	}
}

func (renderer *BasicRenderer) DrawPad(da *gtk.DrawingArea, cr *cairo.Context, animStep float64, p Pad) {
	var (
		pad             = p.Pad()
		x, y    float64 = p.Pos()
		dia, _  float64 = pad.Size()
		focused         = renderer.isFocused(p)
		r, g, b         = 0.1, 0.55, 0.1
	)

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
		cr.SetDash([]float64{4, 4}, 1)
		cr.NewPath()
		cr.Arc(x, y, dia/2+dia/10, -math.Pi, math.Pi)
		cr.ClosePath()
		cr.Stroke()
		cr.SetDash(nil, 0)
	}
}
