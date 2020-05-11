// Package flowui implements an interactive graph renderer.
package flowui

import (
	"fmt"

	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
	"github.com/twitchyliquid64/diagg/flow"
	"github.com/twitchyliquid64/diagg/flowui/render"
)

var flowSelectionSig, _ = glib.SignalNew("flow-selection")
var createdLinkSig, _ = glib.SignalNew("flow-created-link")

// NewFlowchartView constructs a new flowchart display widget, reading nodes
// and position information from the provided layout.
func NewFlowchartView(l *flow.Layout) (*FlowchartView, *gtk.DrawingArea, error) {
	var err error
	fcv := &FlowchartView{
		zoom: 1,
		model: Model{
			l:         l,
			r:         &render.BasicRenderer{},
			nodeState: map[string]modelNode{},
			drawTime:  averageMetric{Name: "draw time"},
			mkHitTime: averageMetric{Name: "hit build time"},
			hitTime:   averageMetric{Name: "hit test time"},
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
	fcv.da.Connect("leave-notify-event", fcv.onLeftFocus)
	fcv.da.SetEvents(int(gdk.POINTER_MOTION_MASK |
		gdk.BUTTON_PRESS_MASK |
		gdk.BUTTON_RELEASE_MASK |
		gdk.SCROLL_MASK |
		gdk.LEAVE_NOTIFY_MASK)) // GDK_MOTION_NOTIFY

	err = fcv.model.initRenderState()

	// Set initial offsets so the left-top side is in full view.
	fcv.offsetX, fcv.offsetY = -fcv.model.nMin.X+30, -fcv.model.nMin.Y+30
	return fcv, fcv.da, err
}

// AddNode inserts a new node into the layout and view.
func (fcv *FlowchartView) AddNode(n flow.Node, x, y float64) error {
	pos := fcv.drawCoordsToFlow(x, y)
	x, y = quantizeCoords(pos.X, pos.Y)
	fcv.model.l.MoveNode(n, x, y)

	if err := fcv.model.buildDrawList(); err != nil {
		return err
	}
	fcv.model.buildModel()
	fcv.da.QueueDraw()
	return nil
}

// DeleteNode removes a node from the flowchart, breaking all links.
func (fcv *FlowchartView) DeleteNode(n flow.Node) error {
	if mn, ok := fcv.model.nodeState[n.NodeID()]; ok {
		fcv.model.h.Delete(mn)
		delete(fcv.model.nodeState, n.NodeID())
	}
	for _, p := range n.Pads() {
		if mn, ok := fcv.model.nodeState[p.PadID()]; ok {
			fcv.model.h.Delete(mn)
		}
		delete(fcv.model.nodeState, p.PadID())
	}

	fcv.model.l.DeleteNode(n)
	return fcv.Rebuild()
}

// Rebuild discards all internal state, rebuilding the view internals from
// the layout.
func (fcv *FlowchartView) Rebuild() error {
	if err := fcv.model.buildDrawList(); err != nil {
		return err
	}
	fcv.model.buildModel()
	fcv.da.QueueDraw()
	return nil
}

func (fcv *FlowchartView) ClearSelection() {
	fcv.model.SetTargetActive(fcv.lmc.target, false)
	fcv.da.QueueDraw()
}

// GetSelection returns the currently selected node or pad.
func (fcv *FlowchartView) GetSelection() interface{} {
	switch t := fcv.lmc.target.(type) {
	case *rectNode:
		return t.Node()
	case *circPad:
		return t.Pad()
	case nil:
		return nil
	default:
		panic(fmt.Sprintf("cannot handle type: %T", t))
	}
}

// GetAtPosition returns the pad or node at the given position, or nil if
// the provided position was empty space.
func (fcv *FlowchartView) GetAtPosition(x, y float64) interface{} {
	tp := fcv.drawCoordsToFlow(x, y)

	if t := fcv.model.HitTest(tp); t != nil {
		switch m := t.(type) {
		case *rectNode:
			return m.Node()
		case *circPad:
			return m.Pad()
		case nil:
			return nil
		default:
			panic(fmt.Sprintf("cannot handle type: %T", m))
		}
	}
	return nil
}

// AddOverlay installs the provided overlay.
func (fcv *FlowchartView) AddOverlay(o Overlay) {
	fcv.overlays = append(fcv.overlays, o)
}

// SetDoubleClickCallback sets a callback to be invoked when a\
// double-click happens with mouse button one.
func (fcv *FlowchartView) SetDoubleClickCallback(cb func(interface{}, float64, float64)) {
	fcv.doublePressCB = cb
}
