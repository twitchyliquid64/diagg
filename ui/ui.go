// Package UI implements an interactive graph renderer.
package ui

import (
	"github.com/gotk3/gotk3/cairo"
	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/gtk"
)

type FlowchartView struct {
	da *gtk.DrawingArea

	dragStartX, dragStartY float64
	dragging               bool

	offsetX float64
	offsetY float64
	zoom    float64

	width, height int
}

func NewFlowchartView() (*FlowchartView, *gtk.DrawingArea, error) {
	var err error
	fcv := &FlowchartView{zoom: 5}

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

	return fcv, fcv.da, nil
}

func (fcv *FlowchartView) onCanvasConfigureEvent(da *gtk.DrawingArea, event *gdk.Event) bool {
	ce := gdk.EventConfigureNewFromEvent(event)
	fcv.width = ce.Width()
	fcv.height = ce.Height()
	return false
}

func (fcv *FlowchartView) onCanvasDrawEvent(da *gtk.DrawingArea, cr *cairo.Context) bool {
	cr.SetSourceRGB(0, 0, 0)
	cr.Paint()
	cr.SetLineWidth(2)

	cr.Translate(float64(30)+fcv.offsetX, float64(30)+fcv.offsetY)
	if fcv.zoom > 0 {
		cr.Scale(fcv.zoom, fcv.zoom)
	}

	cr.SetSourceRGB(1, 1, 1)
	roundedRect(da, cr, 0, 0, 20, 20, 2)
	cr.StrokePreserve()
	cr.SetSourceRGB(0.5, 0.1, 0.1)
	cr.Fill()

	return false
}

func (fcv *FlowchartView) onMotionEvent(area *gtk.DrawingArea, event *gdk.Event) {
	evt := gdk.EventMotionNewFromEvent(event)
	if fcv.dragging {
		x, y := evt.MotionVal()
		fcv.offsetX = -(fcv.dragStartX - x)
		fcv.offsetY = -(fcv.dragStartY - y)
	}
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
	amt := evt.DeltaY()
	if amt == 0 {
		amt = 1
	}

	switch evt.Direction() {
	case gdk.SCROLL_DOWN:
		amt *= -1
	}

	fcv.zoom += amt
	fcv.da.QueueDraw()
}
