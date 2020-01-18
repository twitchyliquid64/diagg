package flowui

import (
	"math"
	"time"

	"github.com/gotk3/gotk3/cairo"
	"github.com/gotk3/gotk3/gtk"
	"github.com/twitchyliquid64/diagg/flow"
	"github.com/twitchyliquid64/diagg/flowui/render"
	"github.com/twitchyliquid64/diagg/hit"
)

// Model encapsulates the state of the flowchart.
type Model struct {
	// Bounds of the flowchart in flowchart coordinates.
	nMin, nMax hit.Point
	// Positions of nodes/pads in the flowchart.
	l *flow.Layout
	// Renderer for nodes/pads during draw.
	r render.Appearance
	// Hit testing for mouse events.
	h *hit.Area
	// Latest display list to use for renders.
	displayList []flow.DrawCommand
	// Maps node/pad ID to state.
	nodeState map[string]modelNode
	// Orphaned nodes (ie: not connected to the graph, just placed)
	orphans []flow.DrawCommand

	// performance metrics
	drawTime  averageMetric
	mkHitTime averageMetric
}

// modelNode represents an element which is part of the flowchart,
// which has both floatchart, positioning, and UI state.
type modelNode interface {
	Pos() (float64, float64)
	Active() bool
	HitTest(hit.Point) bool
}

// rectNode represents flowchart, layout, and UI state information
// for a rectangular node in the flowchart.
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

// circPad represents flowchart, layout, and UI state information
// for a circular pad in the flowchart.
type circPad struct {
	P      flow.Pad
	Layout *flow.PadLayout
	active bool
}

func (p circPad) Pos() (float64, float64) { return p.Layout.Pos() }

func (p circPad) Pad() flow.Pad { return p.P }

func (p circPad) Active() bool { return p.active }

// HitTest returns true as rectangles should be completely represented
// by their min/max points tracked by the hit tester.
func (p circPad) HitTest(tp hit.Point) bool {
	centerX, centerY := p.Pos()
	distSq := math.Pow(tp.X-centerX, 2) + math.Pow(tp.Y-centerY, 2)
	dia, _ := p.Pad().Size()
	return distSq < math.Pow(dia/2, 2)
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
	case *circPad:
		// Not possible to move a pad.
	default:
		panic("cannot handle type")
	}
}

func (m *Model) TargetPos(t hit.TestableObj) (x, y float64) {
	switch t := t.(type) {
	case *rectNode:
		return m.l.Node(t.N.(flow.Node)).Pos()
	case *circPad:
		return m.l.Pad(t.P.(flow.Pad)).Pos()
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

func (m *Model) insertNodeHitObj(c flow.DrawNodeCmd, area *hit.Area) {
	var (
		x, y     = c.Layout.Pos()
		w, h     = c.Node.Size()
		min, max = hit.Point{X: x - w/2, Y: y - h/2}, hit.Point{X: x + w/2, Y: y + h/2}
		nID      = c.Node.NodeID()
	)

	sn, ok := m.nodeState[nID]
	if !ok {
		sn = &rectNode{N: c.Node, Layout: c.Layout}
		m.nodeState[nID] = sn
	}
	area.Add(min, max, sn)
}

func (m *Model) insertPadHitObj(c flow.DrawPadCmd, area *hit.Area) {
	var (
		x, y     = c.Layout.Pos()
		dia, _   = c.Pad.Size()
		min, max = hit.Point{X: x - dia/2, Y: y - dia/2}, hit.Point{X: x + dia/2, Y: y + dia/2}
		pID      = c.Pad.PadID()
	)

	sn, ok := m.nodeState[pID]
	if !ok {
		sn = &circPad{P: c.Pad, Layout: c.Layout}
		m.nodeState[pID] = sn
	}
	area.Add(min, max, sn)
}

func (m *Model) buildHitTester() {
	started := time.Now()
	m.h = hit.NewArea(m.nMin, m.nMax)
	for _, cmd := range m.displayList {
		switch c := cmd.(type) {
		case flow.DrawNodeCmd:
			m.insertNodeHitObj(c, m.h)
		case flow.DrawPadCmd:
			m.insertPadHitObj(c, m.h)
		}
	}

	for _, o := range m.orphans {
		switch c := o.(type) {
		case flow.DrawNodeCmd:
			m.insertNodeHitObj(c, m.h)
		default:
			panic("cannot handle unexpected orphan command")
		}
	}
	m.mkHitTime.Time(started)
}

func (m *Model) Draw(da *gtk.DrawingArea, cr *cairo.Context) {
	started := time.Now()
	for _, cmd := range m.displayList {
		switch c := cmd.(type) {
		case flow.DrawNodeCmd:
			m.r.DrawNode(da, cr, 0, m.nodeState[c.Node.NodeID()].(*rectNode))
		case flow.DrawPadCmd:
			m.r.DrawPad(da, cr, 0, m.nodeState[c.Pad.PadID()].(*circPad))
		}
	}
	for _, cmd := range m.orphans {
		switch c := cmd.(type) {
		case flow.DrawNodeCmd:
			m.r.DrawNode(da, cr, 0, m.nodeState[c.Node.NodeID()].(*rectNode))
		case flow.DrawPadCmd:
			m.r.DrawPad(da, cr, 0, m.nodeState[c.Pad.PadID()].(*circPad))
		}
	}
	m.drawTime.Time(started)
}

func (m *Model) SetTargetActive(target hit.TestableObj, a bool) {
	switch t := target.(type) {
	case nil:
	case *rectNode:
		t.active = a
	case *circPad:
		t.active = a
	default:
		panic("type not handled")
	}
}
