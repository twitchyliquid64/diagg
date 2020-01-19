package flowui

import (
	"math"

	"github.com/twitchyliquid64/diagg/flow"
	"github.com/twitchyliquid64/diagg/hit"
)

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

// lineEdge represents flowchart, layout, and UI state information
// for an edge in the flowchart.
type lineEdge struct {
	E        flow.Edge
	From, To *flow.PadLayout
}

func (e lineEdge) FromPos() (float64, float64) { return e.From.Pos() }
func (e lineEdge) ToPos() (float64, float64)   { return e.To.Pos() }

func (e lineEdge) Pos() (float64, float64) {
	sx, sy := e.FromPos()
	ex, ey := e.ToPos()
	return (sx + ex) / 2, (sy + ey) / 2
}

func (e lineEdge) Edge() flow.Edge { return e.E }

func (e lineEdge) Active() bool { return false }

func (e lineEdge) HitTest(tp hit.Point) bool { return false }
