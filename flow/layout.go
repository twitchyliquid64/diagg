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

type positionable interface {
	Pos() (float64, float64)
}

type sizeable interface {
	Size() (float64, float64)
}

type bounds struct {
	minX, minY float64
	maxX, maxY float64
}

func (b *bounds) update(pos positionable, size sizeable) {
	pX, pY := pos.Pos()
	sX, sY := size.Size()

	lowerX, lowerY := pX-sX/2, pY-sY/2
	if lowerX < b.minX {
		b.minX = lowerX
	}
	if lowerY < b.minY {
		b.minY = lowerY
	}

	higherX, higherY := pX+sX/2, pY+sY/2
	if higherX > b.maxX {
		b.maxX = higherX
	}
	if higherY > b.maxY {
		b.maxY = higherY
	}
}

type dlState struct {
	renderedNodes map[string]struct{}
	renderedPads  map[string]struct{}
	bounds        *bounds
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

func (fl *Layout) DisplayList() (min, max [2]float64, dl []DrawCommand, err error) {
	b := &bounds{}
	b.update(fl.nodes[fl.root.NodeID()], fl.root)

	dl, err = fl.populateDrawList(make([]DrawCommand, 0, 256), fl.root, dlState{
		renderedNodes: make(map[string]struct{}, len(fl.nodes)),
		renderedPads:  make(map[string]struct{}, len(fl.pads)),
		bounds:        b,
	})
	return [2]float64{b.minX, b.minY}, [2]float64{b.maxY, b.maxY}, dl, err
}

func (fl *Layout) populateDrawList(outList []DrawCommand, n Node, s dlState) ([]DrawCommand, error) {
	nID := n.NodeID()
	if _, alreadyProcessed := s.renderedNodes[nID]; alreadyProcessed {
		return outList, nil
	}
	s.renderedNodes[nID] = struct{}{}
	nl := fl.nodes[nID]
	outList = append(outList, DrawNodeCmd{Node: n, Layout: nl})
	s.bounds.update(nl, n)

	for _, p := range n.Pads() {
		pID := p.PadID()
		if _, alreadyProcessed := s.renderedPads[pID]; alreadyProcessed {
			continue
		}
		s.renderedPads[pID] = struct{}{}
		pl := fl.pads[pID]
		outList = append(outList, DrawPadCmd{Pad: p, Layout: pl})
		s.bounds.update(pl, p)
	}
	return outList, nil
}
