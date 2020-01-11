package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/gotk3/gotk3/gtk"
	"github.com/twitchyliquid64/diagg/flow"
	"github.com/twitchyliquid64/diagg/ui"
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

	l := flow.NewLayout(flow.NewSNode("test node", ""))
	fcv, fcvRoot, err := ui.NewFlowchartView(l)
	if err != nil {
		return err
	}
	w.fcv = fcv

	if w.status, err = gtk.LabelNew("X: 0, Y: 0"); err != nil {
		return err
	}

	w.tlBox.Add(w.status)
	w.tlBox.Add(fcvRoot)
	w.win.Add(w.tlBox)
	return nil
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
