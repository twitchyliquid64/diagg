package flowrender

import (
	"github.com/gotk3/gotk3/cairo"
	"github.com/gotk3/gotk3/gtk"
	"github.com/twitchyliquid64/diagg/flow"
)

type positionedElement interface {
	Pos() (float64, float64)
}

type headlineElement interface {
	NodeHeadline() string
}

// Appearance represents an implementation which can display a flowchart.
type Appearance interface {
	DrawNode(da *gtk.DrawingArea, cr *cairo.Context, animStep float64, n flow.Node, layout positionedElement)
	DrawPad(da *gtk.DrawingArea, cr *cairo.Context, animStep float64, n flow.Pad, layout positionedElement)
}

type BasicRenderer struct{}

func (r *BasicRenderer) DrawNode(da *gtk.DrawingArea, cr *cairo.Context, animStep float64, n flow.Node, layout positionedElement) {
	x, y := layout.Pos()
	cr.SetSourceRGB(1, 1, 1)
	roundedRect(da, cr, x, y, 200, 120, 2)
	cr.StrokePreserve()
	cr.SetSourceRGB(0.5, 0.1, 0.1)
	cr.Fill()

	if hln, ok := n.(headlineElement); ok {
		cr.MoveTo(x+7, y+18)
		cr.SetSourceRGB(1, 1, 1)
		cr.SetFontSize(16)
		cr.ShowText(hln.NodeHeadline())
		cr.Fill()
	}
}

func (r *BasicRenderer) DrawPad(da *gtk.DrawingArea, cr *cairo.Context, animStep float64, n flow.Pad, layout positionedElement) {

}
