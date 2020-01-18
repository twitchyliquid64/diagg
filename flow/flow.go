// Package flow implements a flow chart
package flow

import "errors"

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

	StartEdges() []Edge
	EndEdges() []Edge

	ConnectTo(Edge) error
	ConnectFrom(Edge) error
}

// Edge describes a link between two pads.
type Edge interface {
	EdgeID() string
	From() Pad
	To() Pad
}

var ErrSelfLink = errors.New("cannot link to self")
