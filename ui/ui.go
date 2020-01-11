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

type FlowchartView struct {
	da *gtk.DrawingArea

	// TODO: refactor into own type?
	dragStartX, dragStartY float64
	dragging               bool

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
	min, max, fcv.displayList, err = fcv.l.DisplayList()
	fcv.nMin, fcv.nMax = hit.Point{X: min[0], Y: min[1]}, hit.Point{X: max[0], Y: max[1]}
	fcv.h = hit.NewArea(fcv.nMin, fcv.nMax)
	return err
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

	cr.SetSourceRGB(1, 1, 1)
	cr.MoveTo(float64(fcv.width-60), float64(fcv.height-22))
	cr.ShowText(fmt.Sprintf("Zoom: %.2f", fcv.zoom))
	ps := fmt.Sprintf("Pos: %3.2f, %3.2f", fcv.offsetX, fcv.offsetY)
	cr.MoveTo(float64(fcv.width)-cr.TextExtents(ps).Width-4, float64(fcv.height-5))
	cr.ShowText(ps)
	return false
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

func (fcv *FlowchartView) onMotionEvent(area *gtk.DrawingArea, event *gdk.Event) {
	evt := gdk.EventMotionNewFromEvent(event)
	x, y := evt.MotionVal()

	if fcv.dragging {
		fcv.offsetX = -(fcv.dragStartX - x)
		fcv.offsetY = -(fcv.dragStartY - y)
	}

	// tp := hit.Point{X: x - fcv.offsetX, Y: y - fcv.offsetY}
	// hit := fcv.h.Test(tp)
	// fmt.Println(tp, hit)
	fcv.da.QueueDraw()
}

func (fcv *FlowchartView) onPressEvent(area *gtk.DrawingArea, event *gdk.Event) {
	evt := gdk.EventButtonNewFromEvent(event)
	switch evt.Button() {
	case 2, 3: // middle,right button
		fcv.dragging = true
		fcv.dragStartX, fcv.dragStartY = gdk.EventMotionNewFromEvent(event).MotionVal()
		fcv.dragStartX -= fcv.offsetX
		fcv.dragStartY -= fcv.offsetY
	}
}

func (fcv *FlowchartView) onReleaseEvent(area *gtk.DrawingArea, event *gdk.Event) {
	evt := gdk.EventButtonNewFromEvent(event)
	switch evt.Button() {
	case 2, 3: // middle,right button
		fcv.dragging = false
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
