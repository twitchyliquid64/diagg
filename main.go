package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/gotk3/gotk3/gtk"
	"github.com/twitchyliquid64/diagg/flow"
	ui "github.com/twitchyliquid64/diagg/flowui"
)

// Win encapsulates the UI state of the window.
type Win struct {
	win   *gtk.Window
	tlBox *gtk.Box

	fcv    *ui.FlowchartView
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

	root := flow.NewSNode("test node", "")
	p := flow.NewSPad("test", root, flow.SideRight, 0)
	p.SetPadColor(0.1, 0.55, 0.1)
	root.AppendPad(p)
	l := flow.NewLayout(root)
	fcv, fcvRoot, err := ui.NewFlowchartView(l)
	if err != nil {
		return err
	}
	w.fcv = fcv

	on := flow.NewSNode("orphan node", "")
	p = flow.NewSPad("test", on, flow.SideLeft, 0)
	p.SetPadColor(0.7, 0.1, 0.1)
	on.AppendPad(p)
	w.fcv.AddOrphanedNode(on)

	if w.status, err = gtk.LabelNew("Nothing selected"); err != nil {
		return err
	}

	fcvRoot.Connect("flow-selection", w.onFlowSelect)

	w.tlBox.Add(w.status)
	w.tlBox.Add(fcvRoot)
	w.win.Add(w.tlBox)
	return nil
}

func (w *Win) onFlowSelect() {
	sel := w.fcv.GetSelection()
	if sel == nil {
		w.status.SetText("Nothing selected")
	} else {
		w.status.SetText(fmt.Sprintf("Selected %T: %+v", sel, sel))
	}
}

func makeWin() (*Win, error) {
	w := &Win{}
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
