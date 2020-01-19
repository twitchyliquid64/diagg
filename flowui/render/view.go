package render

import (
	"github.com/gotk3/gotk3/cairo"
	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/gtk"
	"github.com/twitchyliquid64/diagg/flow"
)

// HeadlineElement describes nodes which have text labels which should
// be rendered.
type HeadlineElement interface {
	NodeHeadline() string
}

// FocusableElement types are elements which can be focused, and are drawn with
// a thicker outline if they are currently focused.
type FocusableElement interface {
	Active() bool
}

// IconNode describes nodes which have an icon, and need to be rendered in
// their center.
type IconNode interface {
	NodeIcon() *gdk.Pixbuf
}

type Node interface {
	Pos() (float64, float64)
	Node() flow.Node
}

type Pad interface {
	Pos() (float64, float64)
	Pad() flow.Pad
}

type Edge interface {
	FromPos() (float64, float64)
	ToPos() (float64, float64)
	Edge() flow.Edge
}

// Appearance represents an implementation which can display a flowchart.
type Appearance interface {
	DrawNode(da *gtk.DrawingArea, cr *cairo.Context, animStep int64, n Node)
	DrawPad(da *gtk.DrawingArea, cr *cairo.Context, animStep int64, p Pad)
	DrawEdge(da *gtk.DrawingArea, cr *cairo.Context, animStep int64, e Edge)
}

type BasicRenderer struct{}

func (r *BasicRenderer) isFocused(n interface{}) bool {
	if fe, ok := n.(FocusableElement); ok {
		return fe.Active()
	}
	return false
}

func (r *BasicRenderer) DrawNode(da *gtk.DrawingArea, cr *cairo.Context, animStep int64, n Node) {
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

	if hln, ok := node.(HeadlineElement); ok {
		cr.MoveTo(x-hw+7, y-hh+18)
		cr.SetSourceRGB(1, 1, 1)
		cr.SetFontSize(16)
		cr.ShowText(hln.NodeHeadline())
		cr.Fill()
	}

	if in, ok := node.(IconNode); ok {
		pb := in.NodeIcon()
		px, py := x-float64(pb.GetWidth())/2, y-float64(pb.GetHeight())/2
		cr.Translate(px, py)
		//cr.SetAntialias(cairo.ANTIALIAS_NONE)
		gtk.GdkCairoSetSourcePixBuf(cr, pb, 0, 0)
		cr.Paint()
		cr.Translate(-px, -py)
		//cr.SetAntialias(cairo.ANTIALIAS_DEFAULT)
		cr.SetSourceRGB(1, 1, 1)
	}
}

func (renderer *BasicRenderer) DrawEdge(da *gtk.DrawingArea, cr *cairo.Context, animStep int64, e Edge) {
	var (
		sx, sy  float64 = e.FromPos()
		ex, ey  float64 = e.ToPos()
		r, g, b         = 0.9, 0.9, 0.9
	)

	cr.SetSourceRGB(r, g, b)
	cr.MoveTo(sx, sy)
	cr.LineTo(ex, ey)
	cr.Stroke()
}
