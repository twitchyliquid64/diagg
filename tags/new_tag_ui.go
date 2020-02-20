package tags

import (
	"fmt"
	"sort"
	"strings"

	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
)

var newTagSig, _ = glib.SignalNew("new-tag")

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
	v.input.Connect("changed", v.onInputChange)
	v.input.SetSizeRequest(100, 20)

	if err := setupRecentTagsUI(&v); err != nil {
		return nil, err
	}

	v.box.Add(v.input)
	v.box.Add(v.suggestionsBox)
	return &v, nil
}

func setupRecentTagsUI(v *NewTagView) error {
	var err error
	if v.suggestionsBox, err = gtk.ScrolledWindowNew(nil, nil); err != nil {
		return err
	}
	if v.listView, err = gtk.TreeViewNew(); err != nil {
		return err
	}
	renderer1, _ := gtk.CellRendererTextNew()
	renderer2, _ := gtk.CellRendererTextNew()
	// wire attribute 0 in each entry to the text property.
	tagCol, _ := gtk.TreeViewColumnNewWithAttribute("Recently-used tags", renderer1, "text", 0)
	tagCol.SetExpand(true)
	usesCol, _ := gtk.TreeViewColumnNewWithAttribute("Uses", renderer2, "text", 1)
	renderer2.Set("xalign", 1)
	v.listView.AppendColumn(tagCol)
	v.listView.AppendColumn(usesCol)

	if v.baseStore, err = gtk.ListStoreNew(glib.TYPE_STRING, glib.TYPE_STRING); err != nil {
		return err
	}
	if v.filterStore, err = v.baseStore.FilterNew(nil); err != nil {
		return err
	}
	v.filterStore.SetVisibleFunc(v.filterSuggestion)
	v.listView.SetModel(v.filterStore)
	v.listView.Connect("row-activated", v.onSuggestionClicked)

	v.suggestionsBox.Add(v.listView)
	v.suggestionsBox.SetVExpand(true)
	v.Refresh()
	return nil
}

// NewTagView represents a new tag UI component.
type NewTagView struct {
	base        []Tag
	baseStore   *gtk.ListStore
	filterStore *gtk.TreeModelFilter // View of baseStore where items are filtered by name.
	currInput   string

	box            *gtk.Box
	suggestionsBox *gtk.ScrolledWindow
	input          *gtk.Entry
	listView       *gtk.TreeView
}

func (v *NewTagView) filterSuggestion(model *gtk.TreeModelFilter, iter *gtk.TreeIter, _ interface{}) bool {
	if v.currInput == "" {
		return true
	}
	text, _ := model.GetValue(iter, 0) // col 0
	s, _ := text.GetString()
	return strings.Contains(strings.ToUpper(s), v.currInput)
}

func (v *NewTagView) UI() *gtk.Box {
	return v.box
}

func (v *NewTagView) onInputEnter() {
	t, _ := v.input.GetText()
	if t == "" {
		return
	}
	fmt.Printf("New tag: %q\n", t)
	v.input.SetText("")
	v.box.Emit("new-tag", t)
}

func (v *NewTagView) onSuggestionClicked(_ *gtk.TreeView, row *gtk.TreePath) {
	row = v.filterStore.ConvertPathToChildPath(row)
	i, _ := v.baseStore.GetIter(row)
	if i != nil {
		text, _ := v.baseStore.GetValue(i, 0) // col 0 = tag name
		s, _ := text.GetString()
		fmt.Printf("New tag: %q\n", s)
		v.input.SetText("")
		v.box.Emit("new-tag", s)
	}
}

func (v *NewTagView) onInputChange() {
	t, _ := v.input.GetText()
	v.currInput = strings.ToUpper(t)
	v.filterStore.Refilter()
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
