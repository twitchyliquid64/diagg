package ui

import (
	"github.com/gotk3/gotk3/cairo"
	"github.com/gotk3/gotk3/gtk"
	"github.com/twitchyliquid64/diagg/flow"
	"github.com/twitchyliquid64/diagg/ui/flowrender"
)

// FlowNodeState describes the layout state of a flowchart node.
type FlowNodeState struct {
	X, Y float64
}

func (fns FlowNodeState) Pos() (float64, float64) {
	return fns.X, fns.Y
}

// FlowPadState describes the layout state of a flowchart pad.
type FlowPadState struct {
	X, Y float64
}

func (fps FlowPadState) Pos() (float64, float64) {
	return fps.X, fps.Y
}

func NewLayout(root flow.Node) *FlowLayout {
	return &FlowLayout{
		root:     root,
		nodes:    map[string]FlowNodeState{},
		pads:     map[string]FlowPadState{},
		renderer: &flowrender.BasicRenderer{},
	}
}

// FlowLayout keeps track of state describing how elements of a flowchart
// should be positioned.
type FlowLayout struct {
	root  flow.Node
	nodes map[string]FlowNodeState
	pads  map[string]FlowPadState

	// TODO: Keeping track of nodes already drawn is expensive. Maybe we should
	// compute the right order ahead of time?

	renderer flowrender.Appearance
}

func (fl *FlowLayout) Draw(da *gtk.DrawingArea, cr *cairo.Context, animStep float64) error {
	return fl.drawNode(da, cr, fl.root, drawState{
		renderedNodes: make(map[string]struct{}, len(fl.nodes)),
		renderedPads:  make(map[string]struct{}, len(fl.pads)),
		animStep:      animStep,
	})
}

type drawState struct {
	renderedNodes map[string]struct{}
	renderedPads  map[string]struct{}
	animStep      float64
}

func (fl *FlowLayout) drawNode(da *gtk.DrawingArea, cr *cairo.Context, n flow.Node, s drawState) error {
	nID := n.NodeID()
	if _, alreadyProcessed := s.renderedNodes[nID]; alreadyProcessed {
		return nil
	}
	s.renderedNodes[nID] = struct{}{}
	fl.renderer.DrawNode(da, cr, s.animStep, n, fl.nodes[nID])

	for _, p := range n.Pads() {
		pID := p.PadID()
		if _, alreadyProcessed := s.renderedPads[pID]; alreadyProcessed {
			continue
		}
		s.renderedPads[pID] = struct{}{}
		fl.renderer.DrawPad(da, cr, s.animStep, p, fl.pads[pID])
	}

	return nil
}
