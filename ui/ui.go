// Package UI implements an interactive graph renderer.
package ui

import (
	"fmt"

	"github.com/gotk3/gotk3/cairo"
	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/gtk"
	"github.com/twitchyliquid64/diagg/flow"
	"github.com/twitchyliquid64/diagg/hit"
	"github.com/twitchyliquid64/diagg/ui/flowrender"
)

type dragState struct {
	StartX, StartY float64
	ObjX, ObjY     float64
	dragging       bool
	target         hit.TestableObj // only valid for left-mouse-click.
}

type FlowchartView struct {
	da *gtk.DrawingArea

	lmc dragState // left mouse click, like moving a node.
	pan dragState

	// State of the viewport.
	offsetX float64
	offsetY float64
	zoom    float64

	// width/height of the drawing area.
	width, height int

	// State related to the loaded flowchart.
	nMin, nMax  hit.Point
	l           *flow.Layout
	r           flowrender.Appearance
	h           *hit.Area
	displayList []flow.DrawCommand
}

func NewFlowchartView(l *flow.Layout) (*FlowchartView, *gtk.DrawingArea, error) {
	var err error
	fcv := &FlowchartView{
		offsetX: 30,
		offsetY: 30,
		zoom:    1,
		l:       l,
		r:       &flowrender.BasicRenderer{},
	}

	if fcv.da, err = gtk.DrawingAreaNew(); err != nil {
		return nil, nil, err
	}
	fcv.da.SetHAlign(gtk.ALIGN_FILL)
	fcv.da.SetVAlign(gtk.ALIGN_FILL)
	fcv.da.SetHExpand(true)
	fcv.da.SetVExpand(true)
	fcv.da.Connect("draw", fcv.onCanvasDrawEvent)
	fcv.da.Connect("configure-event", fcv.onCanvasConfigureEvent)

	fcv.da.Connect("motion-notify-event", fcv.onMotionEvent)
	fcv.da.Connect("button-press-event", fcv.onPressEvent)
	fcv.da.Connect("button-release-event", fcv.onReleaseEvent)
	fcv.da.Connect("scroll-event", fcv.onScrollEvent)
	fcv.da.SetEvents(int(gdk.POINTER_MOTION_MASK |
		gdk.BUTTON_PRESS_MASK |
		gdk.BUTTON_RELEASE_MASK |
		gdk.SCROLL_MASK)) // GDK_MOTION_NOTIFY

	err = fcv.initRenderState()

	// Set initial offsets so the left-top side is in full view.
	fcv.offsetX, fcv.offsetY = -fcv.nMin.X+30, -fcv.nMin.Y+30
	return fcv, fcv.da, err
}

func (fcv *FlowchartView) initRenderState() (err error) {
	var min, max [2]float64
	if min, max, fcv.displayList, err = fcv.l.DisplayList(); err != nil {
		return err
	}
	fcv.nMin, fcv.nMax = hit.Point{X: min[0], Y: min[1]}, hit.Point{X: max[0], Y: max[1]}
	fcv.buildHitTester()
	return nil
}

// rectHitChecker returns true as rectangles should be completely represented
// by their min/max points tracked by the hit tester.
type rectHitChecker struct{ flow.Node }

func (rectHitChecker) HitTest(p hit.Point) bool {
	return true
}

func (fcv *FlowchartView) buildHitTester() {
	fcv.h = hit.NewArea(fcv.nMin, fcv.nMax)
	for _, cmd := range fcv.displayList {
		switch c := cmd.(type) {
		case flow.DrawNodeCmd:
			x, y := c.Layout.Pos()
			w, h := c.Node.Size()
			min, max := hit.Point{X: x - w/2, Y: y - h/2}, hit.Point{X: x + w/2, Y: y + h/2}
			fcv.h.Add(min, max, rectHitChecker{c.Node})
		case flow.DrawPadCmd:
			panic("not implemented")
		}
	}
}

func (fcv *FlowchartView) onCanvasConfigureEvent(da *gtk.DrawingArea, event *gdk.Event) bool {
	ce := gdk.EventConfigureNewFromEvent(event)
	fcv.width = ce.Width()
	fcv.height = ce.Height()
	return false
}

func (fcv *FlowchartView) onCanvasDrawEvent(da *gtk.DrawingArea, cr *cairo.Context) bool {
	cr.SetSourceRGB(0.12, 0.12, 0.12)
	cr.Paint()
	cr.SetLineWidth(5)

	cr.Save()
	cr.Translate(fcv.offsetX, fcv.offsetY)
	if fcv.zoom > 0 {
		cr.Scale(fcv.zoom, fcv.zoom)
	}
	fcv.draw(da, cr)
	cr.Restore()

	fcv.writeDebugStr(da, cr, fmt.Sprintf("Zoom: %.2f", fcv.zoom), 1)
	fcv.writeDebugStr(da, cr, fmt.Sprintf("Pos: %3.2f, %3.2f", fcv.offsetX, fcv.offsetY), 0)
	return false
}

func (fcv *FlowchartView) writeDebugStr(da *gtk.DrawingArea, cr *cairo.Context, msg string, row int) {
	cr.SetSourceRGB(1, 1, 1)
	cr.MoveTo(float64(fcv.width)-cr.TextExtents(msg).Width-4, float64(fcv.height-5-(20*row)))
	cr.ShowText(msg)
}

func (fcv *FlowchartView) draw(da *gtk.DrawingArea, cr *cairo.Context) {
	for _, cmd := range fcv.displayList {
		switch c := cmd.(type) {
		case flow.DrawNodeCmd:
			fcv.r.DrawNode(da, cr, 0, c.Node, c.Layout)
		case flow.DrawPadCmd:
			fcv.r.DrawPad(da, cr, 0, c.Pad, c.Layout)
		}
	}
}

func (fcv *FlowchartView) drawCoordsToFlow(x, y float64) hit.Point {
	return hit.Point{X: (x - fcv.offsetX) / fcv.zoom, Y: (y - fcv.offsetY) / fcv.zoom}
}

func (fcv *FlowchartView) onMotionEvent(area *gtk.DrawingArea, event *gdk.Event) {
	evt := gdk.EventMotionNewFromEvent(event)
	x, y := evt.MotionVal()
	rebuildHits := false

	if fcv.pan.dragging {
		fcv.offsetX = -(fcv.pan.StartX - x)
		fcv.offsetY = -(fcv.pan.StartY - y)
	}
	if fcv.lmc.dragging && fcv.lmc.target != nil {
		x, y := fcv.lmc.ObjX-(fcv.lmc.StartX-x)/fcv.zoom, fcv.lmc.ObjY-(fcv.lmc.StartY-y)/fcv.zoom
		fcv.l.MoveNode(fcv.lmc.target.(rectHitChecker).Node.(flow.Node), x, y)
		rebuildHits = true
		// TODO: Instead of rebuilding completely, implement scanning the hit tester
		// to update the single value being moved.
	}

	if rebuildHits {
		fcv.buildHitTester()
	}
	fcv.da.QueueDraw()
}

func (fcv *FlowchartView) onPressEvent(area *gtk.DrawingArea, event *gdk.Event) {
	evt := gdk.EventButtonNewFromEvent(event)
	switch evt.Button() {
	case 1: // left mouse button.
		tryClearActive(fcv.lmc.target)
		fcv.lmc.dragging = true
		x, y := gdk.EventMotionNewFromEvent(event).MotionVal()
		fcv.lmc.StartX, fcv.lmc.StartY = x, y
		tp := fcv.drawCoordsToFlow(x, y)
		fcv.lmc.target = fcv.h.Test(tp)
		if fcv.lmc.target != nil {
			fcv.lmc.ObjX, fcv.lmc.ObjY = fcv.l.Node(fcv.lmc.target.(rectHitChecker).Node.(flow.Node)).Pos()
			trySetActive(fcv.lmc.target)
		}
		fcv.da.QueueDraw()

	case 2, 3: // middle,right button
		fcv.pan.dragging = true
		fcv.pan.StartX, fcv.pan.StartY = gdk.EventMotionNewFromEvent(event).MotionVal()
		fcv.pan.StartX -= fcv.offsetX
		fcv.pan.StartY -= fcv.offsetY
	}
}

// activeStateTracker is implemented by any part of the flowchart which can
// track whether it is selected or not.
type activeStateTracker interface {
	SetActive(bool)
}

func trySetActive(target hit.TestableObj) {
	switch t := target.(type) {
	case rectHitChecker:
		if t, ok := t.Node.(activeStateTracker); ok {
			t.SetActive(true)
		}
	default:
		if t, ok := target.(activeStateTracker); ok {
			t.SetActive(true)
		}
	}
}

func tryClearActive(target hit.TestableObj) {
	switch t := target.(type) {
	case rectHitChecker:
		if t, ok := t.Node.(activeStateTracker); ok {
			t.SetActive(false)
		}
	default:
		if t, ok := target.(activeStateTracker); ok {
			t.SetActive(false)
		}
	}
}

func (fcv *FlowchartView) onReleaseEvent(area *gtk.DrawingArea, event *gdk.Event) {
	evt := gdk.EventButtonNewFromEvent(event)
	switch evt.Button() {
	case 1:
		fcv.lmc.dragging = false

	case 2, 3: // middle,right button
		fcv.pan.dragging = false
	}
}

func (fcv *FlowchartView) onScrollEvent(area *gtk.DrawingArea, event *gdk.Event) {
	evt := gdk.EventScrollNewFromEvent(event)
	amt := evt.DeltaY() / 20
	if amt == 0 {
		amt = 0.05
	}

	switch evt.Direction() {
	case gdk.SCROLL_DOWN:
		amt *= -1
	}

	fcv.zoom += amt
	if fcv.zoom <= 0 {
		fcv.zoom = 0.05
	}
	fcv.da.QueueDraw()
}
