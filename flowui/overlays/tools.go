package overlays

import (
	"github.com/gotk3/gotk3/cairo"
	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/gtk"
)

const (
	boxSize       = 48
	lineThickness = 1

	// AnchorToRight can be provided as the X anchor to pin the tool overlay
	// to the right of the screen.
	AnchorToRight = 9001
	// AnchorToLeft can be provided as the X anchor to pin the tool overlay
	// to the left of the screen.
	AnchorToLeft = 1337
)

type Tool struct {
	Icon     *gdk.Pixbuf
	Drop     func(x, y float64)
	Selected func()
}

// ToolOverlay implements a tool selection overlay.
type ToolOverlay struct {
	w, h        int
	boxWidth    float64
	boxHeight   float64
	leftBound   float64
	rightBound  float64
	topBound    float64
	bottomBound float64

	showSelection    bool
	Tools            []Tool
	anchorX, anchorY float64

	selection      int
	dragging       bool
	mouseX, mouseY float64
}

// Toolbar contructs a tool overlay from the given tools.
func Toolbar(selectable bool, tools []Tool) *ToolOverlay {
	return &ToolOverlay{
		selection:     -1,
		showSelection: selectable,
		Tools:         tools,
		anchorX:       0.5,
		anchorY:       0,
	}
}

// ToolbarAtAnchor contructs a tool overlay from the given tools, anchored
// at the specified X and Y ratios. If yAnchor is zero, the toolbar will
// be pinned to the top of the flowchart.
func ToolbarAtAnchor(selectable bool, tools []Tool, xAnchor, yAnchor float64) *ToolOverlay {
	return &ToolOverlay{
		selection:     -1,
		showSelection: selectable,
		Tools:         tools,
		anchorX:       xAnchor,
		anchorY:       yAnchor,
	}
}

// HandleMotionEvent implements flowui.Overlay.
func (o *ToolOverlay) HandleMotionEvent(evt *gdk.EventMotion) bool {
	if o.dragging {
		o.mouseX, o.mouseY = evt.MotionVal()
	}
	return false
}

// HandlePressEvent implements flowui.Overlay.
func (o *ToolOverlay) HandlePressEvent(event *gdk.Event, press bool) bool {
	evt := gdk.EventButtonNewFromEvent(event)
	x, y := evt.MotionVal()
	if !press {
		if o.dragging && o.mouseInCanvas() && o.selection >= 0 {
			if t := o.Tools[o.selection]; t.Drop != nil {
				t.Drop(x, y)
			}
		}
		o.dragging = false
	}

	if x < o.leftBound || x > o.rightBound || y > o.bottomBound || y < o.topBound {
		return false
	}

	if press {
		o.selection = int((x - o.leftBound) / (boxSize + lineThickness))
		o.dragging = true
		o.mouseX, o.mouseY = x, y
		return true
	}
	return false
}

// Draw implements flowui.Overlay.
func (o *ToolOverlay) Draw(da *gtk.DrawingArea, cr *cairo.Context) {
	cr.SetSourceRGB(0.3, 0.3, 0.3)
	cr.SetLineWidth(lineThickness)

	cr.NewPath()
	cr.MoveTo(o.leftBound, o.topBound)
	cr.LineTo(o.rightBound, o.topBound)
	cr.LineTo(o.rightBound, o.topBound+o.boxHeight)
	cr.LineTo(o.leftBound, o.topBound+o.boxHeight)
	cr.LineTo(o.leftBound, o.topBound)
	cr.ClosePath()
	cr.StrokePreserve()
	cr.SetSourceRGB(0.12, 0.12, 0.12)
	cr.Fill()

	cr.SetSourceRGB(0.3, 0.3, 0.3)
	for i := 1; i < len(o.Tools); i++ {
		x := o.leftBound + float64(i*(lineThickness+boxSize))
		cr.MoveTo(x, o.topBound)
		cr.LineTo(x, o.topBound+o.boxHeight)
	}
	cr.Stroke()

	for i, t := range o.Tools {
		if t.Icon != nil {
			x, y := o.leftBound+float64(i*(lineThickness+boxSize)), float64(o.topBound)
			cr.Translate(x, y)
			gtk.GdkCairoSetSourcePixBuf(cr, t.Icon, 0, 0)
			cr.Paint()
			cr.Translate(-x, -y)
		}
	}

	if s := o.selection % len(o.Tools); o.showSelection && s >= 0 {
		cr.SetSourceRGB(1, 1, 1)
		x := o.leftBound + float64(s*(lineThickness+boxSize))
		cr.NewPath()
		cr.MoveTo(x, o.topBound)
		cr.LineTo(x+boxSize, o.topBound)
		cr.LineTo(x+boxSize, o.topBound+o.boxHeight)
		cr.LineTo(x, o.topBound+o.boxHeight)
		cr.ClosePath()
		cr.Stroke()
	}

	// Draw icon & plus while dragging.
	if o.dragging && o.mouseInCanvas() {
		// cr.SetSourceRGB(1, 0, 0)
		// cr.SetLineWidth(2)
		// cr.MoveTo(o.mouseX+8, o.mouseY-8)
		// cr.LineTo(o.mouseX+18, o.mouseY-8)
		// cr.MoveTo(o.mouseX+13, o.mouseY-13)
		// cr.LineTo(o.mouseX+13, o.mouseY-3)
		// cr.Stroke()

		if o.selection >= 0 {
			if i := o.Tools[o.selection].Icon; i != nil {
				px, py := o.mouseX-float64(i.GetWidth())/2, o.mouseY-float64(i.GetHeight())/2
				cr.Translate(px, py)
				gtk.GdkCairoSetSourcePixBuf(cr, i, 0, 0)
				cr.Paint()
				cr.Translate(-px, -py)
			}
		}
	}
}

func (o *ToolOverlay) mouseInCanvas() bool {
	return o.mouseX > o.rightBound || o.mouseX < o.leftBound || o.mouseY > o.bottomBound || o.mouseY < o.topBound
}

// Configure implements flowui.Overlay.
func (o *ToolOverlay) Configure(w, h int) {
	o.w, o.h = w, h
	o.boxWidth = float64(len(o.Tools)*boxSize + len(o.Tools)*lineThickness)
	o.boxHeight = boxSize + lineThickness*2

	if o.anchorX == AnchorToRight {
		o.leftBound = float64(o.w) - o.boxWidth - 3
		o.rightBound = o.leftBound + o.boxWidth
	} else if o.anchorX == AnchorToLeft {
		o.leftBound = 3
		o.rightBound = o.leftBound + o.boxWidth
	} else {
		o.leftBound = (float64(o.w) - o.boxWidth) * o.anchorX
		o.rightBound = o.leftBound + o.boxWidth
	}

	if o.anchorY > 0 {
		o.topBound = float64(o.h)*o.anchorY - float64(o.boxHeight)/2
		o.bottomBound = float64(o.h)*o.anchorY + float64(o.boxHeight)/2
	} else {
		o.topBound = 3
		o.bottomBound = 3 + o.boxHeight
	}
}

// GetBounds returns the bounding box of the toolbar.
func (o *ToolOverlay) GetBounds() (x1, y1, x2, y2 float64) {
	return o.leftBound, o.topBound, o.rightBound, o.bottomBound
}

func (o *ToolOverlay) HandleScrollEvent(evt *gdk.EventScroll) bool {
	return false
}

// HandleKeypress implements tab & numbering shortcuts for keypress events. The
// caller should invoke this function for key presses.
func (o *ToolOverlay) HandleKeypress(keyEvent *gdk.EventKey) bool {
	if o.showSelection {
		kv := keyEvent.KeyVal()
		if kv == gdk.KEY_Tab && keyEvent.State()&gdk.GDK_CONTROL_MASK == 0 {
			o.selection++
			o.selection = o.selection % len(o.Tools)
			o.dragging = false
			if o.Tools[o.selection].Selected != nil {
				o.Tools[o.selection].Selected()
			}
			return true
		}
		if kv >= gdk.KEY_0 && kv <= gdk.KEY_9 && keyEvent.State()&gdk.GDK_CONTROL_MASK == 0 {
			o.selection = int(kv-gdk.KEY_0) - 1%len(o.Tools)
			o.dragging = false
			if o.Tools[o.selection].Selected != nil {
				o.Tools[o.selection].Selected()
			}
			return true
		}
		if kv == gdk.KEY_Escape {
			o.selection = -1
			return true
		}
	}
	return false
}
