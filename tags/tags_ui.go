package tags

import (
	"sync"

	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/gtk"
	"github.com/gotk3/gotk3/pango"
)

var tagProvider *gtk.CssProvider
var globalMut sync.Mutex

func maybeInitCSS() error {
	globalMut.Lock()
	defer globalMut.Unlock()
	if tagProvider != nil {
		return nil
	}

	var err error
	tagProvider, err = gtk.CssProviderNew()
	if err != nil {
		return err
	}
	screen, err := gdk.ScreenGetDefault()
	if err != nil {
		return err
	}
	gtk.AddProviderForScreen(screen, tagProvider, gtk.STYLE_PROVIDER_PRIORITY_APPLICATION)
	return tagProvider.LoadFromData(tagStyling)
}

type tagView struct {
	name   string
	box    *gtk.Box
	remove *gtk.Button
	parent *TagsUI
}

func (tv *tagView) onRemoveClicked() {
	tv.parent.Remove(tv.name)
}

func newTagView(name string, parent *TagsUI) (tagView, *gtk.Box) {
	b, _ := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 1)
	b.SetHAlign(gtk.ALIGN_START)
	b.SetVAlign(gtk.ALIGN_START)
	bStyle, _ := b.GetStyleContext()
	bStyle.AddClass("tag-box")

	l, _ := gtk.LabelNew(name)
	l.SetEllipsize(pango.ELLIPSIZE_END)
	l.SetHExpand(true)
	l.SetHAlign(gtk.ALIGN_START)
	l.SetVAlign(gtk.ALIGN_CENTER)
	lStyle, _ := l.GetStyleContext()
	lStyle.AddClass("tag-name")

	r, _ := gtk.ButtonNewFromIconName("window-close", gtk.ICON_SIZE_MENU)
	rStyle, _ := r.GetStyleContext()
	rStyle.AddClass("tag-close-button")
	tv := tagView{
		name:   name,
		box:    b,
		remove: r,
		parent: parent,
	}

	r.Connect("clicked", tv.onRemoveClicked)

	b.Add(l)
	b.Add(r)
	b.SetHExpand(false)
	b.SetVExpand(false)
	return tv, b
}

func (v *TagsUI) Remove(name string) {
	idx := -1
	for i := range v.tags {
		if v.tags[i].name == name {
			p, _ := v.tags[i].box.GetParent()
			v.box.Remove(p)
			idx = i
			break
		}
	}

	if idx >= 0 {
		v.tags = append(v.tags[:idx], v.tags[idx+1:]...)
	}
}

func (v *TagsUI) Add(name string) {
	// Check it doesnt already exist
	for _, t := range v.tags {
		if t.name == name {
			return
		}
	}

	t, b := newTagView(name, v)
	v.tags = append(v.tags, t)
	v.box.Insert(b, -1)
	b.ShowAll()
}

func MakeTagsView() (*TagsUI, error) {
	v := TagsUI{}
	var err error
	if err = maybeInitCSS(); err != nil {
		return nil, err
	}

	if v.box, err = gtk.FlowBoxNew(); err != nil {
		return nil, err
	}
	v.box.SetHExpand(false)
	v.box.SetVExpand(false)
	v.box.SetMaxChildrenPerLine(600)
	v.box.SetSizeRequest(-1, 40)
	v.box.SetSelectionMode(gtk.SELECTION_NONE)
	return &v, nil
}

type TagsUI struct {
	tags []tagView
	box  *gtk.FlowBox
}

func (v *TagsUI) UI() *gtk.FlowBox {
	return v.box
}

const tagStyling = `
.tag-box {
	border-style: solid;
	border-width: 1px;
	border-color: @borders;
	border-radius: 6px;

	margin: 2px;
	padding: 2px;

	background-color: @insensitive_bg_color;
}

.tag-name {
  margin-left: 8px;
}

.tag-close-button {
	border-style: none;
	background-image: none;
	min-height: 15px;
	min-width: 15px;
}
`
