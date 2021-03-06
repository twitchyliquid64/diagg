package main

import (
	"github.com/gotk3/gotk3/gdk"
	"github.com/twitchyliquid64/diagg/flow"
	"github.com/twitchyliquid64/diagg/flowui/render"
)

type AddDec struct {
	img *gdk.Pixbuf
}

func (d *AddDec) NodeIcon() *gdk.Pixbuf {
	return d.img
}

func (d *AddDec) NodeColor() (float64, float64, float64) {
	return 0.12, 0.22, 0.12
}

func (d *AddDec) NodeOverlayDraw() render.DrawFunc {
	return nil
}

func MakeAdder() *AddNode {
	rawImg := AddButtonImg(55, 55)
	bb := binaryImage(rawImg)

	an := &AddNode{id: flow.AllocNodeID("adder"), img: bb}
	an.inL = flow.NewSPad("add-input-lhs", an, flow.SideLeft, -0.5)
	an.inR = flow.NewSPad("add-input-rhs", an, flow.SideLeft, 0.5)
	an.out = flow.NewSPad("add-output", an, flow.SideRight, 0)
	an.inL.SetPadColor(0.1, 0.6, 0.1)
	an.inR.SetPadColor(0.1, 0.6, 0.1)
	an.out.SetPadColor(0.1, 0.6, 0.1)
	return an
}

type AddNode struct {
	id  string
	img *gdk.Pixbuf
	inL *flow.SPad
	inR *flow.SPad
	out *flow.SPad
}

func (n *AddNode) NodeDecorator() render.NodeDecorator {
	return &AddDec{img: n.img}
}

func (n *AddNode) NodeID() string {
	return n.id
}
func (n *AddNode) Pads() []flow.Pad {
	return []flow.Pad{n.inL, n.inR, n.out}
}
func (n *AddNode) Size() (float64, float64) {
	return 150, 100
}
func (n *AddNode) NodeIcon() *gdk.Pixbuf {
	return n.img
}

// LinkPads implements flowui.UserLinkable.
func (n *AddNode) LinkPads(toNode flow.Node, fromPad, toPad flow.Pad) (flow.Edge, error) {
	for _, e := range append(fromPad.StartEdges(), fromPad.EndEdges()...) {
		switch {
		case e.From() == fromPad && e.To() == toPad:
			return nil, flow.ErrAlreadyLinked
		case e.From() == toPad && e.To() == fromPad:
			return nil, flow.ErrAlreadyLinked
		}
	}
	edge := flow.NewSEdge("", fromPad, toPad)
	if err := fromPad.ConnectTo(edge); err != nil {
		return nil, err
	}
	if err := toPad.ConnectFrom(edge); err != nil {
		return nil, err
	}
	return edge, nil
}
