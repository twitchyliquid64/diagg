package main

import (
	"bytes"
	"encoding/base64"
	"image"
	"image/color"
	"image/png"

	"github.com/gotk3/gotk3/gdk"
	"github.com/twitchyliquid64/diagg/flow"
	"golang.org/x/image/draw"
)

func MakeAdder() *AddNode {
	rawImg := AddButtonImg(55, 55)
	bb, err := gdk.PixbufNew(gdk.COLORSPACE_RGB, true, 8, rawImg.Bounds().Dx(), rawImg.Bounds().Dy())
	if err != nil {
		panic(err)
	}
	buf := bb.GetPixels()
	for i := 0; i < len(rawImg.Pix); i += 4 {
		a := rawImg.Pix[i+3]
		buf[i] = a
		buf[i+1] = a
		buf[i+2] = a
		buf[i+3] = a
	}

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

// AddButtonImg computes a plus button icon at a specific size and
// with a given primary color.
func AddButtonImg(x, y int) *image.RGBA {
	b, err := base64.StdEncoding.DecodeString("iVBORw0KGgoAAAANSUhEUgAAACQAAAAkCAYAAADhAJiYAAAAT0lEQVR4Ae2WAQYAMAhFd7SO9o/WzbYQBmSExXs8AF4SLQCAdyxUausDFO5UBBFEEEHNWKhCv4I8VKF1Td+lBgQNXhlXRlANQQTx5APAbA5+KXS1P2kTZAAAAABJRU5ErkJggg==")
	if err != nil {
		panic(err)
	}
	return decodeAndFormatImg(b, x, y)
}
func decodeAndFormatImg(buf []byte, x, y int) *image.RGBA {
	i, err := png.Decode(bytes.NewBuffer(buf))
	if err != nil {
		panic(err)
	}

	outSize := image.Rect(0, 0, x, y)
	img := image.NewRGBA(outSize)
	draw.BiLinear.Scale(img, outSize, i, i.Bounds(), draw.Over, nil)
	return img
}

// GDKColorToRGBA converts a GDK color to a Go one.
func GDKColorToRGBA(in *gdk.RGBA) color.RGBA {
	c := in.Floats()
	return color.RGBA{
		R: uint8(c[0] * 255),
		G: uint8(c[1] * 255),
		B: uint8(c[2] * 255),
	}
}
