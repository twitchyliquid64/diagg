package flowui

import (
	"fmt"
	"math"

	"github.com/gotk3/gotk3/cairo"
	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/gtk"
	"github.com/twitchyliquid64/diagg/hit"
)

const posQuant = 16

type dragState struct {
	StartX, StartY float64
	DragX, DragY   float64
	dragging       bool

	// only valid for left-mouse-click.
	ObjX, ObjY float64
	target     hit.TestableObj
	sqDist     float64
}

type FlowchartView struct {
	da *gtk.DrawingArea

	lmc         dragState // left mouse click, like moving a node.
	pan         dragState
	hoverTarget *circPad

	animHnd       int
	animStartTime int64
	animTime      int64

	// State of the viewport.
	offsetX float64
	offsetY float64
	zoom    float64

	// width/height of the drawing area.
	width, height int

	model Model
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
	fcv.model.Draw(da, cr, fcv.animTime-fcv.animStartTime)
	if fcv.lmc.dragging {
		if rn, ok := fcv.lmc.target.(*circPad); ok {
			fcv.drawDragLink(da, cr, rn)
		}
	}
	cr.Restore()

	fcv.writeDebugStr(da, cr, fmt.Sprintf("Zoom: %.2f", fcv.zoom), 4)
	fcv.writeDebugStr(da, cr, fmt.Sprintf("Pos: %3.2f, %3.2f", fcv.offsetX, fcv.offsetY), 3)
	fcv.writeDebugStr(da, cr, fmt.Sprintf("%s: %s", fcv.model.drawTime.Metric(), fcv.model.drawTime.Compute()), 2)
	fcv.writeDebugStr(da, cr, fmt.Sprintf("%s: %s", fcv.model.mkHitTime.Metric(), fcv.model.mkHitTime.Compute()), 1)
	fcv.writeDebugStr(da, cr, fmt.Sprintf("%s: %s", fcv.model.hitTime.Metric(), fcv.model.hitTime.Compute()), 0)
	return false
}

func (fcv *FlowchartView) drawDragLink(da *gtk.DrawingArea, cr *cairo.Context, startPad *circPad) {
	x, y := startPad.Pos()
	cr.SetLineWidth(2)
	cr.SetSourceRGB(1, 1, 1)
	cr.MoveTo(x, y)
	cr.LineTo(fcv.lmc.DragX, fcv.lmc.DragY)
	cr.Stroke()
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

	// Handle moving the entire view.
	if fcv.pan.dragging {
		fcv.offsetX = -(fcv.pan.StartX - x)
		fcv.offsetY = -(fcv.pan.StartY - y)
	}

	// Handle moving around nodes & dragging a pad connection.
	if fcv.lmc.dragging && fcv.lmc.target != nil {
		// Update the end position for the drag.
		x, y := fcv.lmc.ObjX-(fcv.lmc.StartX-x)/fcv.zoom, fcv.lmc.ObjY-(fcv.lmc.StartY-y)/fcv.zoom
		fcv.lmc.DragX, fcv.lmc.DragY = x, y

		// If the starting element was a node, we need to handle moving it.
		if _, isNode := fcv.lmc.target.(*rectNode); isNode {
			// Either we stay in the same position, or if the diff is greater than the
			// position quanta, we move the target.
			fcv.lmc.sqDist = math.Pow(fcv.lmc.StartX-x, 2) + math.Pow(fcv.lmc.StartY-y, 2)
			if fcv.lmc.sqDist > (posQuant * posQuant) {
				// Quantize the position.
				x, y = quantizeCoords(x, y)
				fcv.model.MoveTarget(fcv.lmc.target, x, y)
				rebuildHits = true
			}
		}
	}

	// Handle hovering over pads while dragging from another pad.
	if start := fcv.draggingFromPad(); start != nil {
		hoverTarget := fcv.model.HitTest(fcv.drawCoordsToFlow(x, y))
		if fcv.hoverTarget != start && fcv.hoverTarget != nil && fcv.hoverTarget != hoverTarget {
			fcv.clearHoverTarget()
		}
		if endPad, hoversPad := hoverTarget.(*circPad); hoversPad {
			fcv.hoverTarget = endPad
			endPad.active = true
		}
	}

	if rebuildHits {
		// TODO: Instead of rebuilding completely, implement scanning the hit tester
		// to update the single value being moved.
		fcv.model.buildModel()
	}
	fcv.da.QueueDraw()
}

func (fcv *FlowchartView) clearHoverTarget() {
	if fcv.hoverTarget != nil {
		fcv.hoverTarget.active = false
		fcv.hoverTarget = nil
	}
}

func quantizeCoords(x, y float64) (float64, float64) {
	x, y = float64(int(x/posQuant)*posQuant), float64(int(y/posQuant)*posQuant)
	x, y = x+(posQuant/2), y+(posQuant/2)
	return x, y
}

func (fcv *FlowchartView) onPressEvent(area *gtk.DrawingArea, event *gdk.Event) {
	evt := gdk.EventButtonNewFromEvent(event)
	x, y := evt.MotionVal()
	switch evt.Button() {
	case 1: // left mouse button.
		fcv.model.SetTargetActive(fcv.lmc.target, false)
		fcv.lmc.dragging = true
		fcv.lmc.StartX, fcv.lmc.StartY = x, y
		tp := fcv.drawCoordsToFlow(x, y)

		// If we clicked on a node/pad, update the selection state and set the
		// element as active.
		if fcv.lmc.target = fcv.model.HitTest(tp); fcv.lmc.target != nil {
			fcv.lmc.ObjX, fcv.lmc.ObjY = fcv.model.TargetPos(fcv.lmc.target)
			fcv.lmc.DragX, fcv.lmc.DragY = fcv.lmc.ObjX, fcv.lmc.ObjY
			fcv.model.SetTargetActive(fcv.lmc.target, true)

			// If the target is a pad, we should animate the hover circles.
			if _, isPad := fcv.lmc.target.(*circPad); isPad {
				fcv.ensureAnimating()
			}
		}
		fcv.da.Emit("flow-selection")
		fcv.da.QueueDraw()

	case 2, 3: // middle,right button
		fcv.pan.dragging = true
		fcv.pan.StartX, fcv.pan.StartY = x-fcv.offsetX, y-fcv.offsetY
	}
}

// draggingFromPad returns the *circPad of the pad which the user is dragging
// from, or nil if the user is not currently dragging from a pad.
func (fcv *FlowchartView) draggingFromPad() *circPad {
	if !fcv.lmc.dragging {
		return nil
	}
	if startPad, ok := fcv.lmc.target.(*circPad); ok {
		return startPad
	}
	return nil
}

func (fcv *FlowchartView) onReleaseEvent(area *gtk.DrawingArea, event *gdk.Event) {
	evt := gdk.EventButtonNewFromEvent(event)
	x, y := gdk.EventMotionNewFromEvent(event).MotionVal()
	releaseTarget := fcv.model.HitTest(fcv.drawCoordsToFlow(x, y))

	switch evt.Button() {
	case 1:
		// Handle the user dragging from one pad to the other.
		if startPad := fcv.draggingFromPad(); startPad != nil && releaseTarget != nil {
			if endPad, ok := releaseTarget.(*circPad); ok && endPad != startPad {
				if err := fcv.model.OnUserLinksPads(startPad, endPad); err != nil {
					fmt.Printf("failed to link pads: %v\n", err)
				} else {
					fcv.da.Emit("flow-created-link")
				}
			}
		}
		fcv.lmc.dragging = false
		fcv.clearHoverTarget()
		fcv.da.QueueDraw()
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
	fcv.offsetX, fcv.offsetY = fcv.offsetX+amt, fcv.offsetY+amt
	fcv.da.QueueDraw()
}

func (fcv *FlowchartView) ensureAnimating() {
	if fcv.animHnd == 0 && fcv.shouldAnimate() {
		fcv.animHnd = fcv.da.AddTickCallback(fcv.animationTick, 0)
	}
}

func (fcv *FlowchartView) animationTick(widget *gtk.Widget, frameClock *gdk.FrameClock, userData uintptr) bool {
	fcv.animTime = frameClock.GetFrameTime()
	if fcv.animStartTime == 0 {
		fcv.animStartTime = fcv.animTime
	}
	fcv.da.QueueDraw()
	if !fcv.shouldAnimate() {
		fcv.da.RemoveTickCallback(fcv.animHnd)
		fcv.animHnd = 0
		fcv.animStartTime = 0
		return false
	}
	return true
}

// shouldAnimate returns true if draw should be called repeatedly to animate
// the flowchart view.
func (fcv *FlowchartView) shouldAnimate() bool {
	if sp := fcv.draggingFromPad(); sp != nil {
		return true // Animate selected pads.
	}
	return false
}
