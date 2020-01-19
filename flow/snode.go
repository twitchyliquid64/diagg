package flow

type SNode struct {
	Headline string
	id       string
	pads     []Pad
}

func (sn *SNode) NodeID() string {
	return sn.id
}
func (sn *SNode) Pads() []Pad {
	return sn.pads
}
func (sn *SNode) NodeHeadline() string {
	return sn.Headline
}

func (sn *SNode) Size() (float64, float64) {
	return 200, 120
}

func (sn *SNode) AppendSPad(t string, side NodeSide, sideAmt float64) {
	sn.pads = append(sn.pads, NewSPad(t, sn, side, sideAmt))
}

func (sn *SNode) AppendPad(pad Pad) {
	sn.pads = append(sn.pads, pad)
}

// LinkPads implements flowui.UserLinkable.
func (sn *SNode) LinkPads(toNode Node, fromPad, toPad Pad) (Edge, error) {
	for _, e := range append(fromPad.StartEdges(), fromPad.EndEdges()...) {
		switch {
		case e.From() == fromPad && e.To() == toPad:
			return nil, ErrAlreadyLinked
		case e.From() == toPad && e.To() == fromPad:
			return nil, ErrAlreadyLinked
		}
	}
	edge := NewSEdge("", fromPad, toPad)
	if err := fromPad.ConnectTo(edge); err != nil {
		return nil, err
	}
	if err := toPad.ConnectFrom(edge); err != nil {
		return nil, err
	}
	return edge, nil
}

func NewSNode(hl, t string) *SNode {
	return &SNode{
		Headline: hl,
		id:       AllocNodeID(t),
	}
}
