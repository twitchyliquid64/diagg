package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/gtk"
	"github.com/twitchyliquid64/diagg/flow"
	ui "github.com/twitchyliquid64/diagg/flowui"
	"github.com/twitchyliquid64/diagg/flowui/overlays"
)

// Win encapsulates the UI state of the window.
type Win struct {
	win   *gtk.Window
	tlBox *gtk.Box
	tools *overlays.ToolOverlay

	selected flow.Node

	fcv    *ui.FlowchartView
	canvas *gtk.DrawingArea
	status *gtk.Label
}

func (w *Win) build() error {
	var err error

	if w.win, err = gtk.WindowNew(gtk.WINDOW_TOPLEVEL); err != nil {
		return err
	}
	w.win.SetTitle("Diagg demo")
	w.win.Connect("destroy", func() {
		gtk.MainQuit()
	})
	w.win.SetPosition(gtk.WIN_POS_CENTER)
	w.win.SetDefaultSize(800, 600)

	if w.tlBox, err = gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 0); err != nil {
		return err
	}

	// root := flow.NewSNode("test node", "")
	// p := flow.NewSPad("test", root, flow.SideRight, 0)
	// p.SetPadColor(0.1, 0.55, 0.1)
	// root.AppendPad(p)
	l := flow.NewLayout()
	fcv, fcvRoot, err := ui.NewFlowchartView(l)
	if err != nil {
		return err
	}
	w.fcv = fcv
	w.fcv.AddOverlay(w.tools)
	w.fcv.SetDoubleClickCallback(func(obj interface{}, x, y float64) {
		fmt.Printf("double-click: %v at (%v,%v)\n", obj, x, y)
	})
	w.canvas = fcvRoot

	if w.status, err = gtk.LabelNew("Nothing selected"); err != nil {
		return err
	}

	fcvRoot.Connect("flow-selection", w.onFlowSelect)

	w.tlBox.Add(w.status)
	w.tlBox.Add(fcvRoot)
	w.win.Add(w.tlBox)
	return w.setupKeyBindings()
}

func (w *Win) setupKeyBindings() error {
	// TODO: Refactor this into some configurable mapping.
	w.win.Connect("key-press-event", func(win *gtk.Window, ev *gdk.Event) {
		keyEvent := &gdk.EventKey{Event: ev}
		if w.tools.HandleKeypress(keyEvent) {
			w.canvas.QueueDraw()
		}
		if keyEvent.KeyVal() == gdk.KEY_Delete && w.selected != nil {
			w.fcv.DeleteNode(w.selected)
			w.selected = nil
			w.status.SetText("Nothing selected")
		}
	})
	return nil
}

func (w *Win) onFlowSelect() {
	sel := w.fcv.GetSelection()
	if sel == nil {
		w.status.SetText("Nothing selected")
	} else {
		w.status.SetText(fmt.Sprintf("Selected %T: %+v", sel, sel))
	}

	if n, ok := sel.(flow.Node); ok {
		w.selected = n
	} else {
		w.selected = nil
	}
}

func makeWin() (*Win, error) {
	w := &Win{}

	// Color the close image specially.
	closeImg := binaryImage(CloseButtonImg(48, 48))
	p := closeImg.GetPixels()
	for i := 0; i < len(p); i += 4 {
		p[i] = 200
		p[i+1] = 15
		p[i+2] = 15
	}

	tools := []overlays.Tool{
		{Icon: binaryImage(AddButtonImg(48, 48)), Drop: func(x, y float64) {
			w.fcv.AddNode(MakeAdder(), x, y)
		}},
		{Icon: binaryImage(TabImg(48, 48)), Drop: func(x, y float64) {
			fmt.Printf("Screen dragged to %v,%v\n", x, y)
		}},
		{},
		{Icon: closeImg, Drop: func(x, y float64) {
			if n, wasNode := w.fcv.GetAtPosition(x, y).(flow.Node); wasNode {
				w.fcv.DeleteNode(n)
			}
		}},
	}

	w.tools = overlays.Toolbar(false, tools)
	if err := w.build(); err != nil {
		return nil, err
	}
	w.win.ShowAll()
	return w, nil
}

func main() {
	gtk.Init(nil)
	flag.Parse()

	if _, err := makeWin(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	gtk.Main()
}
