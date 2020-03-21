package tags

import (
	"sort"
	"strings"

	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
)

var newTagSig, _ = glib.SignalNew("new-tag")

// MakeNewTagView creates a UI element for creating views.
func MakeNewTagView() (*NewTagView, error) {
	v := NewTagView{}
	var err error
	if v.box, err = gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 0); err != nil {
		return nil, err
	}

	if v.input, err = gtk.EntryNew(); err != nil {
		return nil, err
	}
	v.input.Connect("activate", v.onInputEnter)
	v.input.Connect("changed", v.onInputChange)
	v.input.SetHExpand(true)
	v.input.SetMarginStart(4)
	v.input.SetMarginEnd(4)

	if v.baseStore, err = gtk.ListStoreNew(glib.TYPE_STRING, glib.TYPE_STRING); err != nil {
		return nil, err
	}

	e, _ := gtk.EntryCompletionNew()
	e.SetModel(v.baseStore)
	e.SetTextColumn(0)
	e.Connect("match-selected", v.onSuggestionClicked)
	v.input.SetCompletion(e)

	v.box.Add(v.input)
	return &v, nil
}

// NewTagView represents a new tag UI component.
type NewTagView struct {
	base       []Tag
	baseStore  *gtk.ListStore
	currInput  string
	lastNewTag string

	box   *gtk.Box
	input *gtk.Entry
}

func (v *NewTagView) UI() *gtk.Box {
	return v.box
}

func (v *NewTagView) GrabFocus() {
	v.input.GrabFocus()
}

func (v *NewTagView) GetNewTagName() string {
	return v.lastNewTag
}

func (v *NewTagView) newTag(name string) {
	v.lastNewTag = name
	v.input.SetText("")
	v.box.Emit("new-tag")
}

func (v *NewTagView) onInputEnter() {
	t, _ := v.input.GetText()
	if t == "" {
		return
	}
	v.newTag(t)
}

func (v *NewTagView) onSuggestionClicked(_ *gtk.EntryCompletion, _ *gtk.TreeModel, i *gtk.TreeIter) bool {
	text, _ := v.baseStore.GetValue(i, 0) // col 0 = tag name
	s, _ := text.GetString()
	v.newTag(s)
	return true
}

func (v *NewTagView) onInputChange() {
	t, _ := v.input.GetText()
	v.currInput = strings.ToUpper(t)
}

func (v *NewTagView) SetSuggestions(suggestions []Tag) {
	v.base = suggestions
	sort.Slice(v.base, func(i, j int) bool {
		return v.base[i].Seen < v.base[j].Seen
	})
	v.Refresh()
}

func (v *NewTagView) Refresh() {
	v.baseStore.Clear()
	for _, s := range v.base {
		v.baseStore.Set(v.baseStore.Append(), []int{0, 1}, []interface{}{s.Name, s.SeenCount()})
	}
}
