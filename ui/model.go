package ui

import (
	"github.com/gotk3/gotk3/cairo"
	"github.com/gotk3/gotk3/gtk"
	"github.com/twitchyliquid64/diagg/flow"
	"github.com/twitchyliquid64/diagg/hit"
	"github.com/twitchyliquid64/diagg/ui/flowrender"
)

// Model encapsulates the state of the flowchart.
type Model struct {
	// Bounds of the flowchart in flowchart coordinates.
	nMin, nMax hit.Point
	// Positions of nodes/pads in the flowchart.
	l *flow.Layout
	// Renderer for nodes/pads during draw.
	r flowrender.Appearance
	// Hit testing for mouse events.
	h *hit.Area
	// Latest display list to use for renders.
	displayList []flow.DrawCommand
	// Maps node/pad ID to state.
	// TODO: Abstract value type to interface.
	nodeState map[string]*rectNode
}

// rectNode represents both the flowchart and layout information for
// a rectangular node in the flowchart.
type rectNode struct {
	N      flow.Node
	Layout *flow.NodeLayout
	active bool
}

func (n rectNode) Pos() (float64, float64) { return n.Layout.Pos() }

func (n rectNode) Node() flow.Node { return n.N }

func (n rectNode) Active() bool { return n.active }

// HitTest returns true as rectangles should be completely represented
// by their min/max points tracked by the hit tester.
func (rectNode) HitTest(p hit.Point) bool {
	return true
}

func (m *Model) maybeUpdateMinMax(x, y float64) {
	if x > m.nMax.X {
		m.nMax.X = x
	} else if x < m.nMin.X {
		m.nMin.X = x
	}
	if y > m.nMax.Y {
		m.nMax.Y = y
	} else if y < m.nMin.Y {
		m.nMin.Y = y
	}
}

func (m *Model) MoveTarget(t hit.TestableObj, x, y float64) {
	switch t := t.(type) {
	case *rectNode:
		m.maybeUpdateMinMax(x, y)
		m.l.MoveNode(t.N.(flow.Node), x, y)
	default:
		panic("cannot handle type")
	}
}

func (m *Model) TargetPos(t hit.TestableObj) (x, y float64) {
	switch t := t.(type) {
	case *rectNode:
		return m.l.Node(t.N.(flow.Node)).Pos()
	default:
		panic("cannot handle type")
	}
}

// initRenderState populates the display list and builds the hit tester.
func (m *Model) initRenderState() (err error) {
	var min, max [2]float64
	if min, max, m.displayList, err = m.l.DisplayList(); err != nil {
		return err
	}
	m.nMin, m.nMax = hit.Point{X: min[0], Y: min[1]}, hit.Point{X: max[0], Y: max[1]}
	m.buildHitTester()
	return nil
}

func (m *Model) buildHitTester() {
	m.h = hit.NewArea(m.nMin, m.nMax)
	for _, cmd := range m.displayList {
		switch c := cmd.(type) {
		case flow.DrawNodeCmd:
			x, y := c.Layout.Pos()
			w, h := c.Node.Size()
			min, max := hit.Point{X: x - w/2, Y: y - h/2}, hit.Point{X: x + w/2, Y: y + h/2}

			sn, ok := m.nodeState[c.Node.NodeID()]
			if !ok {
				sn = &rectNode{N: c.Node, Layout: c.Layout}
				m.nodeState[c.Node.NodeID()] = sn
			}
			m.h.Add(min, max, sn)

		case flow.DrawPadCmd:
			panic("not implemented")
		}
	}
}

func (m *Model) Draw(da *gtk.DrawingArea, cr *cairo.Context) {
	for _, cmd := range m.displayList {
		switch c := cmd.(type) {
		case flow.DrawNodeCmd:
			m.r.DrawNode(da, cr, 0, m.nodeState[c.Node.NodeID()])
		case flow.DrawPadCmd:
			m.r.DrawPad(da, cr, 0, m.nodeState[c.Pad.PadID()])
		}
	}
}

func (m *Model) SetTargetActive(target hit.TestableObj, a bool) {
	switch t := target.(type) {
	case nil:
	case *rectNode:
		t.active = a
	default:
		panic("type not handled")
	}
}
