// Package list implements a list adapter widget.
package list

import (
	"os"
	"sync"

	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/gtk"
)

// EmptyList implements the ListData interface.
type EmptyList struct{}

func (EmptyList) Len() int                                  { return 0 }
func (EmptyList) Equal(interface{}, interface{}) bool       { return false }
func (EmptyList) GetItem(position int) (interface{}, error) { return nil, os.ErrNotExist }

// ListData provides the UI with data to display.
type ListData interface {
	Len() int
	Equal(interface{}, interface{}) bool
	GetItem(position int) (interface{}, error)
}

// ListItemBuilder constructs list items given the data they
// should contain.
type ListItemBuilder interface {
	ListItem(data interface{}, view interface{}) (gtk.IWidget, interface{}, error)
}

type item struct {
	b       *gtk.Box
	e       *gtk.EventBox
	w       gtk.IWidget
	visible bool

	data interface{}
	view interface{}
}

func (i *item) onDelete() {}

func (i *item) onKeyPress(b *gtk.EventBox, ev *gdk.Event) {}

func (i *item) onClicked(b *gtk.EventBox, event *gdk.Event) {}

// List implements a scrollable list of items.
type List struct {
	scroll *gtk.ScrolledWindow
	box    *gtk.Box

	l       sync.Mutex
	data    ListData
	builder ListItemBuilder
	items   []item
}

// UI returns the top-level widget representing the list.
func (l *List) UI() gtk.IWidget {
	return l.scroll
}

// SetData sets the backing data for the list and forces an update.
func (l *List) SetData(d ListData) error {
	l.l.Lock()
	l.data = d
	l.l.Unlock()
	return l.Update()
}

// UpdateIdx updates the UI for just the list item at the provided position.
// This method is intended to be more efficient in some cases.
func (l *List) UpdateIdx(i int) error {
	l.l.Lock()
	defer l.l.Unlock()

	switch {
	// item already exists - update in place
	case i < len(l.items):
		data, err := l.data.GetItem(i)
		if err != nil {
			return err
		}
		return l.updateExistingPosition(i, data)

		// item just off the end - create + append
	case i == len(l.items):
		data, err := l.data.GetItem(i)
		if err != nil {
			return err
		}
		r, err := l.makeItem(data)
		if err != nil {
			return err
		}
		l.items = append(l.items, r)
		return nil
	default:
		return l.update()
	}
}

func (l *List) updateExistingPosition(i int, data interface{}) error {
	var err error
	if !l.data.Equal(l.items[i].data, data) {
		prev := l.items[i].w
		l.items[i].w, l.items[i].view, err = l.builder.ListItem(data, l.items[i].view)
		if err != nil {
			return err
		}
		l.items[i].data = data
		if prev != l.items[i].w {
			l.items[i].b.Remove(prev)
			l.items[i].b.Add(l.items[i].w)
		}
	}
	if !l.items[i].visible {
		l.items[i].visible = true
		l.items[i].b.SetVisible(true)
	}
	return nil
}

// Update forces a refresh of all list data.
func (l *List) Update() error {
	l.l.Lock()
	defer l.l.Unlock()
	return l.update()
}

func (l *List) update() error {
	itemLen := len(l.items)
	listLen := l.data.Len()

	for i := 0; i < listLen; i++ {
		data, err := l.data.GetItem(i)
		if err != nil {
			return err
		}
		if i < itemLen { // an item at this offset exists
			if err := l.updateExistingPosition(i, data); err != nil {
				return err
			}
		} else {
			// No item already exists, make one.
			r, err := l.makeItem(data)
			if err != nil {
				return err
			}
			l.items = append(l.items, r)
		}
	}

	// Hide any items which exist past the end of the list.
	for i := listLen; i < len(l.items); i++ {
		l.items[i].visible = false
		l.items[i].b.SetVisible(false)
	}
	return nil
}

func (l *List) makeItem(data interface{}) (item, error) {
	var (
		r   = item{data: data, visible: true}
		err error
	)
	if r.w, r.view, err = l.builder.ListItem(data, nil); err != nil {
		return item{}, err
	}

	if r.b, err = gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 0); err != nil {
		return item{}, err
	}
	r.b.SetCanFocus(true)
	if r.e, err = gtk.EventBoxNew(); err != nil {
		return item{}, err
	}
	r.e.Connect("button-press-event", r.onClicked)
	r.e.Connect("key-press-event", r.onKeyPress)
	r.e.Connect("delete-event", r.onDelete)

	r.b.Add(r.w)
	r.e.Add(r.b)
	l.box.Add(r.e)
	r.e.ShowAll()
	return r, nil
}

// New creates a new list widget using the provided parameters.
// If you do not wish to indicate a desired size, -1 can be provided for
// wantX and wantY.
func New(builder ListItemBuilder, paddingY, wantX, wantY int) (*List, error) {
	l := List{
		data:    EmptyList{},
		builder: builder,
	}
	var err error

	if l.box, err = gtk.BoxNew(gtk.ORIENTATION_VERTICAL, paddingY); err != nil {
		return nil, err
	}
	l.box.SetHExpand(true)
	l.box.SetVExpand(true)
	l.box.SetSizeRequest(wantX, wantY)

	if l.scroll, err = gtk.ScrolledWindowNew(nil, nil); err != nil {
		return nil, err
	}
	l.scroll.Add(l.box)

	return &l, nil
}
