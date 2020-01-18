package flow

type SEdge struct {
	id       string
	from, to Pad
}

func (se *SEdge) EdgeID() string {
	return se.id
}

func (se *SEdge) From() Pad {
	return se.from
}

func (se *SEdge) To() Pad {
	return se.to
}

func NewSEdge(t string, from, to Pad) *SEdge {
	return &SEdge{
		id:   AllocEdgeID(t),
		from: from,
		to:   to,
	}
}
