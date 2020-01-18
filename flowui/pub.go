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

// AddOrphanedNode inserts a new node into the layout and view, unconnected to
// any other nodes.
func (fcv *FlowchartView) AddOrphanedNode(n flow.Node) {
	w, h := n.Size()
	x, y := (fcv.offsetX-float64(fcv.width)+w/2)/fcv.zoom, (fcv.offsetY-float64(fcv.height)+h/2)/fcv.zoom
	fcv.model.l.MoveNode(n, x, y)

	fcv.model.orphans = append(fcv.model.orphans, flow.DrawNodeCmd{
		Layout: fcv.model.l.Node(n),
		Node:   n,
	})
	fcv.model.buildHitTester()
	fcv.da.QueueDraw()
}

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
