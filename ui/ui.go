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

	model Model
}

func NewFlowchartView(l *flow.Layout) (*FlowchartView, *gtk.DrawingArea, error) {
	var err error
	fcv := &FlowchartView{
		zoom: 1,
		model: Model{
			l:         l,
			r:         &flowrender.BasicRenderer{},
			nodeState: map[string]modelNode{},
		},
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

	err = fcv.model.initRenderState()

	// Set initial offsets so the left-top side is in full view.
	fcv.offsetX, fcv.offsetY = -fcv.model.nMin.X+30, -fcv.model.nMin.Y+30
	return fcv, fcv.da, err
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
	cr.SetFillRule(cairo.FILL_RULE_EVEN_ODD)

	cr.Save()
	cr.Translate(fcv.offsetX, fcv.offsetY)
	if fcv.zoom > 0 {
		cr.Scale(fcv.zoom, fcv.zoom)
	}
	fcv.model.Draw(da, cr)
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
		fcv.model.MoveTarget(fcv.lmc.target, x, y)
		rebuildHits = true
		// TODO: Instead of rebuilding completely, implement scanning the hit tester
		// to update the single value being moved.
	}

	if rebuildHits {
		fcv.model.buildHitTester()
	}
	fcv.da.QueueDraw()
}

func (fcv *FlowchartView) onPressEvent(area *gtk.DrawingArea, event *gdk.Event) {
	evt := gdk.EventButtonNewFromEvent(event)
	x, y := gdk.EventMotionNewFromEvent(event).MotionVal()
	switch evt.Button() {
	case 1: // left mouse button.
		fcv.model.SetTargetActive(fcv.lmc.target, false)
		fcv.lmc.dragging = true
		fcv.lmc.StartX, fcv.lmc.StartY = x, y
		tp := fcv.drawCoordsToFlow(x, y)

		if fcv.lmc.target = fcv.model.h.Test(tp); fcv.lmc.target != nil {
			fcv.lmc.ObjX, fcv.lmc.ObjY = fcv.model.TargetPos(fcv.lmc.target)
			fcv.model.SetTargetActive(fcv.lmc.target, true)
		}
		fcv.da.QueueDraw()

	case 2, 3: // middle,right button
		fcv.pan.dragging = true
		fcv.pan.StartX, fcv.pan.StartY = x-fcv.offsetX, y-fcv.offsetY
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
