package main

import (
	"flag"

	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
	"github.com/twitchyliquid64/diagg/list"
)

// Win encapsulates the UI state of the window.
type Win struct {
	win    *gtk.Window
	box    *gtk.Box
	l      *list.List
	addBtn *gtk.Button

	model dataModel
}

func convertLabel(w *gtk.Widget) *gtk.Label {
	return gtk.WrapMap["GtkLabel"].(func(obj *glib.Object) *gtk.Label)(w.InitiallyUnowned.Object)
}

func convertEntry(w *gtk.Widget) *gtk.Entry {
	return gtk.WrapMap["GtkEntry"].(func(obj *glib.Object) *gtk.Entry)(w.InitiallyUnowned.Object)
}

func convertButton(w *gtk.Widget) *gtk.Button {
	return gtk.WrapMap["GtkButton"].(func(obj *glib.Object) *gtk.Button)(w.InitiallyUnowned.Object)
}

// ListItem implements list.ListItemBuilder, wiring a row of data to
// a set of widgets for a row.
func (w *Win) ListItem(data interface{}, v interface{}) (gtk.IWidget, interface{}, error) {
	var (
		row  = data.(rowModel)
		view *item
	)

	if v != nil {
		view = v.(*item)
	} else {
		var err error
		if view, err = mintItem(); err != nil {
			return nil, nil, err
		}
	}

	view.unwire()

	view.entry.SetText(row.name)

	view.onDeletePressed(func() {
		w.model.delete(row.uuid)
		w.l.Update()
	})
	view.onSavePressed(func() {
		name, _ := view.entry.GetText()
		w.model.updateName(row.uuid, name)
		w.l.Update()
	})

	return view.box, view, nil
}

func (w *Win) build() error {
	var err error

	if w.win, err = gtk.WindowNew(gtk.WINDOW_TOPLEVEL); err != nil {
		return err
	}
	w.win.SetTitle("List demo")
	w.win.Connect("destroy", func() {
		gtk.MainQuit()
	})
	w.win.SetDefaultSize(400, 300)
	w.win.SetPosition(gtk.WIN_POS_CENTER)

	if w.box, err = gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 2); err != nil {
		return err
	}

	if w.l, err = list.New(w, 0, -1, -1); err != nil {
		return err
	}

	w.box.Add(w.l.UI())

	if w.addBtn, err = gtk.ButtonNewFromIconName("list-add", gtk.ICON_SIZE_MENU); err != nil {
		return err
	}
	w.addBtn.Connect("clicked", w.add)
	w.box.Add(w.addBtn)

	w.win.Add(w.box)
	return w.l.SetData(&w.model)
}

func (w *Win) add() {
	w.model.add()
	w.l.UpdateIdx(w.model.Len())
}

func main() {
	gtk.Init(nil)
	flag.Parse()

	w := &Win{
		model: dataModel{rows: []rowModel{
			{"port A", "1234"},
			{"port B", "1111"},
		}},
	}
	if err := w.build(); err != nil {
		panic(err)
	}
	w.win.ShowAll()
	gtk.Main()
}
