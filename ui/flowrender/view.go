package flowrender

import (
	"github.com/gotk3/gotk3/cairo"
	"github.com/gotk3/gotk3/gtk"
	"github.com/twitchyliquid64/diagg/flow"
)

type positionedElement interface {
	Pos() (float64, float64)
}

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

// Appearance represents an implementation which can display a flowchart.
type Appearance interface {
	DrawNode(da *gtk.DrawingArea, cr *cairo.Context, animStep float64, n flow.Node, layout positionedElement)
	DrawPad(da *gtk.DrawingArea, cr *cairo.Context, animStep float64, n flow.Pad, layout positionedElement)
}

type BasicRenderer struct{}

func (r *BasicRenderer) isFocused(n flow.Node) bool {
	if fe, ok := n.(focusableElement); ok {
		return fe.Active()
	}
	return false
}

func (r *BasicRenderer) DrawNode(da *gtk.DrawingArea, cr *cairo.Context, animStep float64, n flow.Node, layout positionedElement) {
	var (
		x, y        float64 = layout.Pos()
		w, h        float64 = n.Size()
		hw, hh      float64 = w / 2, h / 2
		borderWidth float64 = 2
	)
	if r.isFocused(n) {
		borderWidth = 6
	}

	cr.SetSourceRGB(1, 1, 1)
	cr.SetLineWidth(borderWidth)
	roundedRect(da, cr, x-hw, y-hh, w, h, 2)
	cr.StrokePreserve()
	cr.SetSourceRGB(0.5, 0.1, 0.1)
	cr.Fill()

	if hln, ok := n.(headlineElement); ok {
		cr.MoveTo(x-hw+7, y-hh+18)
		cr.SetSourceRGB(1, 1, 1)
		cr.SetFontSize(16)
		cr.ShowText(hln.NodeHeadline())
		cr.Fill()
	}
}

func (r *BasicRenderer) DrawPad(da *gtk.DrawingArea, cr *cairo.Context, animStep float64, n flow.Pad, layout positionedElement) {

}
