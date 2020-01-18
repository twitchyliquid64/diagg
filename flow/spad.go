package flow

type SPad struct {
	id      string
	side    NodeSide
	sideAmt float64
	parent  Node

	startEdges []Edge
	endEdges   []Edge
}

func (sp *SPad) PadID() string {
	return sp.id
}
func (sp *SPad) Size() (float64, float64) {
	return 25, 25
}

func (sp *SPad) StartEdges() []Edge {
	return sp.startEdges
}
func (sp *SPad) EndEdges() []Edge {
	return sp.endEdges
}

func (sp *SPad) ConnectTo(e Edge) error {
	if e.To() == sp {
		return ErrSelfLink
	}
	sp.startEdges = append(sp.startEdges, e)
	return nil
}

func (sp *SPad) ConnectFrom(e Edge) error {
	if e.From() == sp {
		return ErrSelfLink
	}
	sp.endEdges = append(sp.endEdges, e)
	return nil
}

func (sp *SPad) Parent() Node {
	return sp.parent
}

func (sp *SPad) Positioning() (NodeSide, float64) {
	return sp.side, sp.sideAmt
}

func NewSPad(t string, parent Node, side NodeSide, sideAmt float64) *SPad {
	return &SPad{
		parent:  parent,
		side:    side,
		sideAmt: sideAmt,
		id:      AllocPadID(t),
	}
}
