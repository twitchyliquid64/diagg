# diagg

A WIP set of UI widgets for Go + GTK3. Intended to hide most of the complexity of common widgets.

### `flowui` package

`flowui` implements an interactive display for a flowchart of connected nodes. Nodes are linked
by edges, which themselves form a link between any two pads. Pads are the connection points on
a node.

The data model used for the nodes can be any type implementing the requisite interfaces in the
`flow` package. Basic implementations are provided as `flow.SNode`, `flow.SPad`, and `flow.SEdge`.

`flowui` additionally implements:

1. Ability to pan and zoom around the flowchart
1. Ability to have 'toolbars' and overlays on the screen, intended for easy creation of new nodes
1. Ability to select or double-click nodes or pads
1. Ability to create new edges by dragging a line between pads
1. Ability to add new nodes to the flowchart
1. Custom renderers for nodes, pads, and edges (thou the default renderer is pretty good)
1. Hit-testing & display culling to maximize performance

Try `go run flowdemo/*.go` for a demo.

### `form` package

`form` implements automatic generation of form elements/fields/windows based on a struct. This
is intended for rapid development of simple dialogs.

Try `go run formdemo/*.go` for a demo.

### `list` package

`list` implements a list of widgets, with an MVC-based API mirroring that of Android's RecyclerView
and ListAdapter API. Try `go run listdemo/*.go` for a demo.

### `editor` and `tags` packages

Implement a syntax-highlighted editor widget, and a widget for adding/removing tags.
