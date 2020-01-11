// Package flow implements a flow chart
package flow

// Node describes a symbol in a flowchart.
type Node interface {
	NodeID() string
	Pads() []Pad
	Size() (float64, float64)
}

// Pad describes a connection point on a node.
type Pad interface {
	PadID() string

	// ConnectType restricts the pads which this pad can be connected to,
	// if it returns a non-nil value.
	ConnectType() interface{}
}
