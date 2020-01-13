package flow

type SPad struct {
	id string
}

func (sp *SPad) PadID() string {
	return sp.id
}
func (sp *SPad) Size() (float64, float64) {
	return 35, 35
}

func NewSPad(t string) *SPad {
	return &SPad{
		id: AllocPadID(t),
	}
}
