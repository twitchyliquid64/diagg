package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/alecthomas/chroma/lexers"
	"github.com/gotk3/gotk3/gtk"
	"github.com/twitchyliquid64/diagg/editor"
)

// Win encapsulates the UI state of the window.
type Win struct {
	win        *gtk.Window
	tlBox      *gtk.Box
	editScroll *gtk.ScrolledWindow

	editor *editor.Editor
}

func (w *Win) build() error {
	var err error

	if w.win, err = gtk.WindowNew(gtk.WINDOW_TOPLEVEL); err != nil {
		return err
	}
	w.win.SetTitle("Editor demo")
	w.win.Connect("destroy", func() {
		gtk.MainQuit()
	})
	w.win.SetPosition(gtk.WIN_POS_CENTER)
	//w.win.SetDefaultSize(800, 600)

	if w.tlBox, err = gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 0); err != nil {
		return err
	}
	if w.editScroll, err = gtk.ScrolledWindowNew(nil, nil); err != nil {
		return err
	}
	if w.editor, err = editor.New(lexers.Get("markdown"), nil); err != nil {
		return err
	}

	w.editScroll.Add(w.editor.UI())
	w.tlBox.Add(w.editScroll)
	w.win.Add(w.tlBox)
	return nil
}

func makeWin() (*Win, error) {
	w := &Win{}
	if err := w.build(); err != nil {
		return nil, err
	}
	w.win.SetSizeRequest(600, 400)
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
