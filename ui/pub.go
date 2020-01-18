package ui

import (
	"fmt"

	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/gtk"
	"github.com/twitchyliquid64/diagg/flow"
	"github.com/twitchyliquid64/diagg/ui/flowrender"
)

// NewFlowchartView constructs a new flowchart display widget, reading nodes
// and position information from the provided layout.
func NewFlowchartView(l *flow.Layout) (*FlowchartView, *gtk.DrawingArea, error) {
	var err error
	fcv := &FlowchartView{
		zoom: 1,
		model: Model{
			l:         l,
			r:         &flowrender.BasicRenderer{},
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

func (fcv *FlowchartView) AddOrphanedNode(n flow.Node) {
	x, y := (fcv.offsetX+float64(fcv.width/2))/fcv.zoom, (fcv.offsetY+float64(fcv.height/2))/fcv.zoom
	fmt.Println(x, y)
	fcv.model.l.MoveNode(n, x, y)
	fcv.model.orphans = append(fcv.model.orphans, flow.DrawNodeCmd{
		Layout: fcv.model.l.Node(n),
		Node:   n,
	})
	fcv.model.buildHitTester()
	fcv.da.QueueDraw()
}
