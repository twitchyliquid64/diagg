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

func (sn *SNode) AppendPad(t string, side NodeSide, sideAmt float64) {
	sn.pads = append(sn.pads, NewSPad(t, sn, side, sideAmt))
}

// LinkPads implements flowui.UserLinkable.
func (sn *SNode) LinkPads(toNode Node, fromPad, toPad Pad) (Edge, error) {
	return NewSEdge("", fromPad, toPad), nil
}

func NewSNode(hl, t string) *SNode {
	return &SNode{
		Headline: hl,
		id:       AllocNodeID(t),
	}
}
