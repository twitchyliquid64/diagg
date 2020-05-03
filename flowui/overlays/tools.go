package overlays

import (
	"github.com/gotk3/gotk3/cairo"
	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/gtk"
)

const (
	boxSize       = 48
	lineThickness = 1
	topOffset     = 3
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
	leftBound   float64
	rightBound  float64
	bottomBound float64

	showSelection bool
	Tools         []Tool

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

	if x < o.leftBound || x > o.rightBound || y > o.bottomBound || y < topOffset {
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
	cr.MoveTo(o.leftBound, topOffset)
	cr.LineTo(float64(o.w)-o.leftBound, topOffset)
	cr.LineTo(float64(o.w)-o.leftBound, topOffset+boxSize+lineThickness*2)
	cr.LineTo(o.leftBound, topOffset+boxSize+lineThickness*2)
	cr.LineTo(o.leftBound, topOffset)
	cr.ClosePath()
	cr.StrokePreserve()
	cr.SetSourceRGB(0.12, 0.12, 0.12)
	cr.Fill()

	cr.SetSourceRGB(0.3, 0.3, 0.3)
	for i := 1; i < len(o.Tools); i++ {
		x := o.leftBound + float64(i*(lineThickness+boxSize))
		cr.MoveTo(x, topOffset)
		cr.LineTo(x, topOffset+boxSize+lineThickness*2)
	}
	cr.Stroke()

	for i, t := range o.Tools {
		if t.Icon != nil {
			x, y := o.leftBound+float64(i*(lineThickness+boxSize)), float64(topOffset)
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
		cr.MoveTo(x, topOffset)
		cr.LineTo(x+boxSize, topOffset)
		cr.LineTo(x+boxSize, topOffset+boxSize+lineThickness*2)
		cr.LineTo(x, topOffset+boxSize+lineThickness*2)
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
	return o.mouseX > o.rightBound || o.mouseX < o.leftBound || o.mouseY > o.bottomBound
}

// Configure implements flowui.Overlay.
func (o *ToolOverlay) Configure(w, h int) {
	o.w, o.h = w, h
	o.boxWidth = float64(len(o.Tools)*boxSize + len(o.Tools)*lineThickness)
	o.leftBound = (float64(o.w) - o.boxWidth) / 2
	o.rightBound = o.leftBound + o.boxWidth
	o.bottomBound = topOffset + boxSize + lineThickness*2
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
