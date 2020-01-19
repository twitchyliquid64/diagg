package flow

type SPad struct {
	id      string
	side    NodeSide
	sideAmt float64
	parent  Node

	startEdges []Edge
	endEdges   []Edge

	r, g, b float64
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

func (sp *SPad) PadColor() (float64, float64, float64) {
	return sp.r, sp.g, sp.b
}
func (sp *SPad) SetPadColor(r, g, b float64) {
	sp.r, sp.g, sp.b = r, g, b
}

func NewSPad(t string, parent Node, side NodeSide, sideAmt float64) *SPad {
	return &SPad{
		parent:  parent,
		side:    side,
		sideAmt: sideAmt,
		id:      AllocPadID(t),
		r:       0.5,
		g:       0.5,
		b:       0.5,
	}
}
