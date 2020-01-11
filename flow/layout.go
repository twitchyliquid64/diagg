package flow

// NodeLayout describes the layout state of a flowchart node.
type NodeLayout struct {
	X, Y float64
}

func (fns NodeLayout) Pos() (float64, float64) {
	return fns.X, fns.Y
}

// PadLayout describes the layout state of a flowchart pad.
type PadLayout struct {
	X, Y float64
}

func (fps PadLayout) Pos() (float64, float64) {
	return fps.X, fps.Y
}

// NewLayout constructs a new layout controller, to keep track of the position
// and compute the draw order of flowchart nodes.
func NewLayout(root Node) *Layout {
	return &Layout{
		root:  root,
		nodes: map[string]NodeLayout{},
		pads:  map[string]PadLayout{},
	}
}

// Layout keeps track of state describing how elements of a flowchart
// should be positioned.
type Layout struct {
	root  Node
	nodes map[string]NodeLayout
	pads  map[string]PadLayout

	// TODO: Keeping track of nodes already drawn is expensive. Maybe we should
	// compute the right order ahead of time?
}

type dlState struct {
	renderedNodes map[string]struct{}
	renderedPads  map[string]struct{}
}

type DrawObject uint8

// Valid DrawObject types.
const (
	DrawNode DrawObject = iota
	DrawPad
)

type DrawNodeCmd struct {
	Node   Node
	Layout NodeLayout
}

func (c DrawNodeCmd) DrawObject() DrawObject {
	return DrawNode
}

type DrawPadCmd struct {
	Pad    Pad
	Layout PadLayout
}

func (c DrawPadCmd) DrawObject() DrawObject {
	return DrawPad
}

type DrawCommand interface {
	DrawObject() DrawObject
}

func (fl *Layout) DisplayList() ([]DrawCommand, error) {
	dl := make([]DrawCommand, 0, 256)
	return fl.populateDrawList(dl, fl.root, dlState{
		renderedNodes: make(map[string]struct{}, len(fl.nodes)),
		renderedPads:  make(map[string]struct{}, len(fl.pads)),
	})
}

func (fl *Layout) populateDrawList(outList []DrawCommand, n Node, s dlState) ([]DrawCommand, error) {
	nID := n.NodeID()
	if _, alreadyProcessed := s.renderedNodes[nID]; alreadyProcessed {
		return outList, nil
	}
	s.renderedNodes[nID] = struct{}{}
	nl := fl.nodes[nID]
	outList = append(outList, DrawNodeCmd{Node: n, Layout: nl})

	for _, p := range n.Pads() {
		pID := p.PadID()
		if _, alreadyProcessed := s.renderedPads[pID]; alreadyProcessed {
			continue
		}
		s.renderedPads[pID] = struct{}{}
		outList = append(outList, DrawPadCmd{Pad: p, Layout: fl.pads[pID]})
	}
	return outList, nil
}
