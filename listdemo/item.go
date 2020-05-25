package main

import (
	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
)

// item encapsulates the view state for a row in the list.
type item struct {
	box   *gtk.Box
	label *gtk.Label
	entry *gtk.Entry
	save  *gtk.Button
	del   *gtk.Button

	saveHandler *glib.SignalHandle
	delHandler  *glib.SignalHandle
}

func (i *item) unwire() {
	if i.saveHandler != nil {
		i.save.HandlerDisconnect(*i.saveHandler)
		i.saveHandler = nil
	}
	if i.delHandler != nil {
		i.del.HandlerDisconnect(*i.delHandler)
		i.delHandler = nil
	}
}

func (i *item) onDeletePressed(cb func()) {
	hnd, _ := i.del.Connect("clicked", cb)
	i.delHandler = &hnd
}

func (i *item) onSavePressed(cb func()) {
	hnd, _ := i.save.Connect("clicked", cb)
	i.saveHandler = &hnd
}

func mintItem() (*item, error) {
	var (
		err error
		out item
	)
	if out.box, err = gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 2); err != nil {
		return nil, err
	}
	if out.label, err = gtk.LabelNew("::"); err != nil {
		return nil, err
	}
	if out.entry, err = gtk.EntryNew(); err != nil {
		return nil, err
	}
	if out.save, err = gtk.ButtonNewFromIconName("document-save", gtk.ICON_SIZE_MENU); err != nil {
		return nil, err
	}
	if out.del, err = gtk.ButtonNewFromIconName("edit-delete", gtk.ICON_SIZE_MENU); err != nil {
		return nil, err
	}
	out.box.PackStart(out.label, false, false, 2)
	out.box.PackStart(out.entry, true, true, 2)
	out.box.PackStart(out.save, false, false, 2)
	out.box.PackStart(out.del, false, false, 2)
	out.box.ShowAll()
	return &out, nil
}
