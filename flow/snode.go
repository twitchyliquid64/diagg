package flow

type SNode struct {
	Headline string
	id       string
}

func (sn *SNode) NodeID() string {
	return sn.id
}
func (sn *SNode) Pads() []Pad {
	return nil
}
func (sn *SNode) NodeHeadline() string {
	return sn.Headline
}

func (sn *SNode) Size() (float64, float64) {
	return 200, 120
}

func NewSNode(hl, t string) *SNode {
	return &SNode{
		Headline: hl,
		id:       AllocNodeID(t),
	}
}
