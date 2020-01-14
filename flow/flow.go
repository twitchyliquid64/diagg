// Package flow implements a flow chart
package flow

type NodeSide uint8

// Valid NodeSide values.
const (
	SideRight NodeSide = iota
	SideLeft
	SideTop
	SideBottom
)

// Node describes a symbol in a flowchart.
type Node interface {
	NodeID() string
	Pads() []Pad
	Size() (float64, float64)
}

// Pad describes a connection point on a node.
type Pad interface {
	PadID() string
	Size() (float64, float64)
	Parent() Node

	// Positioning returns the side the pad should be positioned on its node,
	// as well as how far down the axis from the leftmost part of the side.
	Positioning() (NodeSide, float64)

	// ConnectType restricts the pads which this pad can be connected to,
	// if it returns a non-nil value.
	ConnectType() interface{}
}
