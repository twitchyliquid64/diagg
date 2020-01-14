package flow

type SPad struct {
	id      string
	side    NodeSide
	sideAmt float64
	parent  Node
}

func (sp *SPad) PadID() string {
	return sp.id
}
func (sp *SPad) Size() (float64, float64) {
	return 25, 25
}

func (sp *SPad) ConnectType() interface{} {
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
