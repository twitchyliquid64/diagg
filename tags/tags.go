package tags

import (
	"fmt"
	"sort"

	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
)

// NewNewTagView creates a UI element for creating views.
func NewNewTagView() (*NewTagView, error) {
	v := NewTagView{}
	var err error
	if v.box, err = gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 0); err != nil {
		return nil, err
	}
	if v.input, err = gtk.EntryNew(); err != nil {
		return nil, err
	}
	v.input.Connect("activate", v.onInputEnter)

	if v.listView, err = gtk.TreeViewNew(); err != nil {
		return nil, err
	}
	if v.listStore, err = gtk.ListStoreNew(glib.TYPE_STRING); err != nil {
		return nil, err
	}
	v.listView.SetModel(v.listStore)
	renderer, _ := gtk.CellRendererTextNew()
	// wire attribute 0 in each entry to the text property.
	column, _ := gtk.TreeViewColumnNewWithAttribute("", renderer, "text", 0)
	eb, _ := gtk.FixedNew()
	column.SetWidget(eb) // Set header widget to empty box
	v.listView.AppendColumn(column)

	v.listStore.SetValue(v.listStore.Append(), 0, "yeet")

	v.box.Add(v.input)
	v.box.Add(v.listView)
	return &v, nil
}

// NewTagView represents a new tag UI component.
type NewTagView struct {
	suggestions []candidateTag
	listStore   *gtk.ListStore

	box      *gtk.Box
	input    *gtk.Entry
	listView *gtk.TreeView
}

type Suggestion interface {
	Name() string
	ID() string
	Order() int
}

func (v *NewTagView) UI() *gtk.Box {
	return v.box
}

func (v *NewTagView) onInputEnter() {
	t, _ := v.input.GetText()
	fmt.Printf("New tag: %q\n", t)
	v.input.SetText("")
}

func (v *NewTagView) UpdateSuggestions(suggestions []Suggestion) {
	sort.Slice(suggestions, func(i, j int) bool {
		return suggestions[i].Order() < suggestions[j].Order()
	})
	out := make([]candidateTag, len(suggestions))
	for i := range suggestions {
		out[i].suggestion = suggestions[i]
	}
}

type candidateTag struct {
	suggestion Suggestion
}
