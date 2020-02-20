package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/gotk3/gotk3/gtk"
	"github.com/twitchyliquid64/diagg/tags"
)

var defaultTags = []string{
	"yeet",
	"work",
	"rust",
	"embedded",
	"crypto",
	"TPMs",
}

func makeDefaultTags() []tags.Tag {
	out := make([]tags.Tag, len(defaultTags))
	for i := range defaultTags {
		out[i] = tags.Tag{Name: defaultTags[i]}
	}
	return out
}

// Win encapsulates the UI state of the window.
type Win struct {
	win   *gtk.Window
	tlBox *gtk.Box

	tags *tags.TagsUI
	nte  *tags.NewTagView
}

func (w *Win) build() error {
	var err error

	if w.win, err = gtk.WindowNew(gtk.WINDOW_TOPLEVEL); err != nil {
		return err
	}
	w.win.SetTitle("Tags demo")
	w.win.Connect("destroy", func() {
		gtk.MainQuit()
	})
	w.win.SetPosition(gtk.WIN_POS_CENTER)
	w.win.SetDefaultSize(800, 600)

	if w.tlBox, err = gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 0); err != nil {
		return err
	}
	if w.tags, err = tags.MakeTagsView(); err != nil {
		return err
	}
	if w.nte, err = tags.MakeNewTagView(); err != nil {
		return err
	}
	w.nte.SetSuggestions(makeDefaultTags())
	w.nte.UI().Connect("new-tag", w.onNewTag)

	w.tlBox.Add(w.tags.UI())
	w.tlBox.Add(w.nte.UI())
	w.win.Add(w.tlBox)
	return nil
}

func (w *Win) onNewTag() {
	t := w.nte.GetNewTagName()
	w.tags.Add(t)
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
