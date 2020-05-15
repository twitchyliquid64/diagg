package render

import "github.com/twitchyliquid64/diagg/flow"

// ColoredPad describes a pad with a custom color.
type ColoredPad interface {
	PadColor() (float64, float64, float64)
}

// DecoratedNode describes nodes which provide decoration information.
type DecoratedNode interface {
	NodeDecorator() NodeDecorator
}

type Node interface {
	Pos() (float64, float64)
	Node() flow.Node
}

type Pad interface {
	Pos() (float64, float64)
	Pad() flow.Pad
}

type Edge interface {
	FromPos() (float64, float64)
	ToPos() (float64, float64)
	Edge() flow.Edge
}
