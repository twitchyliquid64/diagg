package flowui

import (
	"errors"
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

// initRenderState populates the display list and builds internal state.
func (m *Model) initRenderState() (err error) {
	if err := m.buildDrawList(); err != nil {
		return err
	}
	m.buildModel()
	return nil
}

func (m *Model) buildDrawList() error {
	var min, max [2]float64
	var err error
	if min, max, m.displayList, err = m.l.DisplayList(); err != nil {
		return err
	}
	m.nMin, m.nMax = hit.Point{X: min[0], Y: min[1]}, hit.Point{X: max[0], Y: max[1]}
	return nil
}

func (m *Model) insertNodeObj(c flow.DrawNodeCmd, area *hit.Area) {
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

func (m *Model) insertPadObj(c flow.DrawPadCmd, area *hit.Area) {
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

func (m *Model) buildModel() {
	started := time.Now()
	m.h = hit.NewArea(m.nMin, m.nMax)
	for _, cmd := range m.displayList {
		switch c := cmd.(type) {
		case flow.DrawNodeCmd:
			m.insertNodeObj(c, m.h)
		case flow.DrawPadCmd:
			m.insertPadObj(c, m.h)
		case flow.DrawEdgeCmd:
			m.nodeState[c.Edge.EdgeID()] = &lineEdge{
				E:    c.Edge,
				From: m.l.Pad(c.Edge.From()),
				To:   m.l.Pad(c.Edge.To()),
			}
		}
	}

	for _, o := range m.orphans {
		switch c := o.(type) {
		case flow.DrawNodeCmd:
			m.insertNodeObj(c, m.h)
			for _, p := range c.Node.Pads() {
				m.insertPadObj(flow.DrawPadCmd{
					Layout: m.l.Pad(p),
					Pad:    p,
				}, m.h)
			}
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
		case flow.DrawEdgeCmd:
			m.r.DrawEdge(da, cr, 0, m.nodeState[c.Edge.EdgeID()].(*lineEdge))
		}
	}
	for _, cmd := range m.orphans {
		switch c := cmd.(type) {
		case flow.DrawNodeCmd:
			m.r.DrawNode(da, cr, 0, m.nodeState[c.Node.NodeID()].(*rectNode))
			for _, p := range c.Node.Pads() {
				m.r.DrawPad(da, cr, 0, m.nodeState[p.PadID()].(*circPad))
			}
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

var ErrNodeNotLinkable = errors.New("node cannot be linked by user")

// UserLinkable describes a node which can have pads linked to another by
// the user performing a drag from one node to another.
type UserLinkable interface {
	flow.Node
	LinkPads(toNode flow.Node, fromPad, toPad flow.Pad) (flow.Edge, error)
}

func (m *Model) OnUserLinksPads(startPad, endPad *circPad) error {
	fromNode, toNode := startPad.P.Parent(), endPad.P.Parent()

	if linkableBaseNode, ok := fromNode.(UserLinkable); ok {
		if _, err := linkableBaseNode.LinkPads(toNode, startPad.P, endPad.P); err != nil {
			return err
		}
		// At this stage the two pads have had an edge allocated and been
		// successfully connected. If either node were previously orphaned,
		// they will need to be removed from the orphan list as they are now
		// connected.
		m.removeNodeOrphan(fromNode)
		m.removeNodeOrphan(toNode)
		// Lastly, we rebuild the display list to account for the new edge
		// and any decendant links.
		if err := m.buildDrawList(); err != nil {
			return err
		}
		m.buildModel()
		return nil
	}

	return ErrNodeNotLinkable
}

// removeNode Orphan removes a node from the orphan list, usually due to the
// node becoming attached to the flowchart via a newly-created edge. True is
// returned if the orphan was found and hence removed.
func (m *Model) removeNodeOrphan(n flow.Node) bool {
	idx := -1
	for i := 0; i < len(m.orphans); i++ {
		if mn, ok := m.orphans[i].(flow.DrawNodeCmd); ok && mn.Node == n {
			idx = i
			break
		}
	}
	if idx >= 0 {
		m.orphans = append(m.orphans[:idx], m.orphans[:idx+1]...)
		return true
	}
	return false
}
