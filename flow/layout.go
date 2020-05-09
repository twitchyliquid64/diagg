package flow

import (
	"errors"
)

// ErrIsRoot is returned if deletion of the root node is attempted.
var ErrIsRoot = errors.New("cannot delete the root node")

// NodeLayout describes the layout state of a flowchart node.
type NodeLayout struct {
	X, Y float64
}

func (fns *NodeLayout) Pos() (float64, float64) {
	if fns == nil {
		return 0, 0
	}
	return fns.X, fns.Y
}

// PadLayout describes the layout state of a flowchart pad.
type PadLayout struct {
	X, Y float64
}

func (fps *PadLayout) Pos() (float64, float64) {
	if fps == nil {
		return 0, 0
	}
	return fps.X, fps.Y
}

// NewLayout constructs a new layout controller, to keep track of the position
// and compute the draw order of flowchart nodes.
func NewLayout() *Layout {
	return &Layout{
		allNodes: map[string]Node{},
		nodes:    map[string]*NodeLayout{},
		pads:     map[string]*PadLayout{},
	}
}

// Layout keeps track of state describing how elements of a flowchart
// should be positioned.
type Layout struct {
	root     Node
	allNodes map[string]Node
	nodes    map[string]*NodeLayout
	pads     map[string]*PadLayout
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
	renderedEdges map[string]struct{}
	bounds        *bounds
}

type DrawObject uint8

// Valid DrawObject types.
const (
	DrawNode DrawObject = iota
	DrawPad
	DrawEdge
)

type DrawNodeCmd struct {
	Node   Node
	Layout *NodeLayout
}

func (c DrawNodeCmd) DrawObject() DrawObject {
	return DrawNode
}

type DrawPadCmd struct {
	Pad    Pad
	Layout *PadLayout
}

func (c DrawPadCmd) DrawObject() DrawObject {
	return DrawPad
}

type DrawEdgeCmd struct {
	From, To             Pad
	FromLayout, ToLayout *PadLayout
	Edge                 Edge
}

func (c DrawEdgeCmd) DrawObject() DrawObject {
	return DrawEdge
}

type DrawCommand interface {
	DrawObject() DrawObject
}

func (fl *Layout) MoveNode(n Node, x, y float64) {
	nID := n.NodeID()
	if nl, ok := fl.nodes[nID]; ok {
		nl.X = x
		nl.Y = y
	} else {
		fl.nodes[nID] = &NodeLayout{X: x, Y: y}
		fl.allNodes[nID] = n
	}

	// As pad position is dependent on node position, force recomputation.
	for _, p := range n.Pads() {
		fl.padPosRecompute(p)
	}
}

// DeleteNode removes a node from the layout, destroying all edges to other
// nodes in the layout.
func (fl *Layout) DeleteNode(n Node) {
	nID := n.NodeID()
	if fl.root != nil && nID == fl.root.NodeID() {
		fl.findNewRoot()
	}

	delete(fl.nodes, nID)
	delete(fl.allNodes, nID)

	for _, p := range n.Pads() {
		p.DisconnectAll()
		delete(fl.pads, p.PadID())
	}
}

func (fl *Layout) findNewRoot() {
	// Try and select a root adjacent to the existing root if there is one.
	if fl.root != nil {
		for _, p := range fl.root.Pads() {
			for _, e := range p.StartEdges() {
				if fl.root.NodeID() != e.To().Parent().NodeID() {
					fl.root = e.To().Parent()
					return
				}
			}
		}
	}
	// Otherwise, anything will do.
	for nID, n := range fl.allNodes {
		if fl.root != nil && nID == fl.root.NodeID() {
			continue
		}
		fl.root = n
		return
	}
	fl.root = nil
}

// Node returns the object storing the position of a node.
func (fl *Layout) Node(n Node) *NodeLayout {
	nID := n.NodeID()
	if nl, ok := fl.nodes[nID]; ok {
		return nl
	}
	nl := &NodeLayout{}
	fl.nodes[nID] = nl
	fl.allNodes[nID] = n
	return nl
}

type NodePosition struct {
	Pos  NodeLayout
	Node Node
}

// Dump returns a list of all nodes and their positions.
func (fl *Layout) Dump() []NodePosition {
	out := make([]NodePosition, 0, len(fl.allNodes))
	for _, n := range fl.allNodes {
		out = append(out, NodePosition{
			Node: n,
			Pos:  *fl.Node(n),
		})
	}
	return out
}

func (fl *Layout) padPosRecompute(p Pad) *PadLayout {
	var (
		side, sideAmt = p.Positioning()
		parentLayout  = fl.Node(p.Parent())
		w, h          = p.Parent().Size()
		pID           = p.PadID()
		pl            = fl.pads[pID]
	)
	if pl == nil {
		pl = &PadLayout{}
	}

	switch side {
	case SideRight:
		pl.X, pl.Y = parentLayout.X+w/2, parentLayout.Y+h*sideAmt/2
	case SideLeft:
		pl.X, pl.Y = parentLayout.X-w/2, parentLayout.Y+h*sideAmt/2
	case SideBottom:
		pl.X, pl.Y = parentLayout.X+w*sideAmt/2, parentLayout.Y+h/2
	case SideTop:
		pl.X, pl.Y = parentLayout.X+w*sideAmt/2, parentLayout.Y-h/2
	}

	fl.pads[pID] = pl
	return pl
}

func (fl *Layout) Pad(p Pad) *PadLayout {
	pID := p.PadID()
	if pl, ok := fl.pads[pID]; ok {
		return pl
	}
	return fl.padPosRecompute(p)
}

// DisplayList returns a list of drawing commands for rendering the layout.
// The bounds of all objects in the layout are also returned.
func (fl *Layout) DisplayList() (min, max [2]float64, dl []DrawCommand, err error) {
	if fl.root == nil {
		fl.findNewRoot()
	}
	if fl.root == nil {
		// No nodes to render.
		return [2]float64{}, [2]float64{}, nil, nil
	}

	b := &bounds{}
	b.update(fl.nodes[fl.root.NodeID()], fl.root)

	state := dlState{
		renderedNodes: make(map[string]struct{}, 4+len(fl.nodes)),
		renderedPads:  make(map[string]struct{}, 12+len(fl.pads)),
		renderedEdges: make(map[string]struct{}, 32),
		bounds:        b,
	}

	dl, err = fl.populateDrawListNode(make([]DrawCommand, 0, 256), fl.root, state)
	if err != nil {
		return [2]float64{}, [2]float64{}, nil, err
	}

	// Run through the full list again to make sure we have any orphaned nodes.
	for _, n := range fl.allNodes {
		dl, err = fl.populateDrawListNode(dl, n, state)
	}

	return [2]float64{b.minX, b.minY}, [2]float64{b.maxY, b.maxY}, dl, err
}

func (fl *Layout) populateDrawListNode(outList []DrawCommand, n Node, s dlState) ([]DrawCommand, error) {
	nID := n.NodeID()
	if _, alreadyProcessed := s.renderedNodes[nID]; alreadyProcessed {
		return outList, nil
	}
	s.renderedNodes[nID] = struct{}{}
	nl := fl.Node(n)
	outList = append(outList, DrawNodeCmd{Node: n, Layout: nl})
	s.bounds.update(nl, n)

	for _, p := range n.Pads() {
		var err error
		if outList, err = fl.populateDrawListPad(outList, p, n, s); err != nil {
			return nil, err
		}
	}
	return outList, nil
}

func (fl *Layout) populateDrawListPad(outList []DrawCommand, p Pad, parent Node, s dlState) ([]DrawCommand, error) {
	pID := p.PadID()
	if _, alreadyProcessed := s.renderedPads[pID]; alreadyProcessed {
		return outList, nil
	}

	s.renderedPads[pID] = struct{}{}
	pl := fl.Pad(p)
	outList = append(outList, DrawPadCmd{Pad: p, Layout: pl})
	s.bounds.update(pl, p)

	for _, se := range p.StartEdges() {
		var err error
		if outList, err = fl.populateDrawListEdge(outList, se, p, s); err != nil {
			return nil, err
		}
	}
	for _, ee := range p.EndEdges() {
		var err error
		if outList, err = fl.populateDrawListEdge(outList, ee, p, s); err != nil {
			return nil, err
		}
	}
	return outList, nil
}

func (fl *Layout) populateDrawListEdge(outList []DrawCommand, e Edge, parent Pad, s dlState) ([]DrawCommand, error) {
	eID := e.EdgeID()
	if _, alreadyProcessed := s.renderedEdges[eID]; alreadyProcessed {
		return outList, nil
	}
	s.renderedEdges[eID] = struct{}{}

	// In case the referenced pad has not been rendered, render it before the
	// edge so the edge appears on top.
	var (
		err            error
		toPad, fromPad = e.To(), e.From()
		toPl, fromPl   = fl.Pad(toPad), fl.Pad(fromPad)
	)
	if outList, err = fl.populateDrawListNode(outList, toPad.Parent(), s); err != nil {
		return nil, err
	}
	if outList, err = fl.populateDrawListNode(outList, fromPad.Parent(), s); err != nil {
		return nil, err
	}

	return append(outList, DrawEdgeCmd{
		From:       fromPad,
		To:         toPad,
		FromLayout: fromPl,
		ToLayout:   toPl,
		Edge:       e,
	}), nil
}
