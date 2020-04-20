package form

import (
	"fmt"

	"github.com/gotk3/gotk3/gtk"
)

// formRow describes a live representation of a form field.
type formRow struct {
	box    *gtk.Box
	widget gtk.IWidget
	spec   *formField
}

func makeRow(field *formField) (*formRow, error) {
	row, err := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 2)
	if err != nil {
		return nil, err
	}
	row.SetMarginTop(4)
	lab, err := gtk.LabelNew(field.label)
	if err != nil {
		return nil, err
	}
	lab.SetHExpand(true)
	lab.SetHAlign(gtk.ALIGN_START)
	lab.SetXAlign(0)
	lab.SetJustify(gtk.JUSTIFY_LEFT)
	lab.SetMarginStart(4)
	row.Add(lab)

	fr := formRow{
		box:  row,
		spec: field,
	}
	switch fr.spec.inputType {
	case InputText, InputInt, InputUint:
		var w *gtk.Entry
		if w, err = gtk.EntryNew(); err != nil {
			return nil, fmt.Errorf("new entry: %w", err)
		}
		w.SetMarginEnd(4)
		fr.widget = w
	case InputBool:
		var w *gtk.Switch
		if w, err = gtk.SwitchNew(); err != nil {
			return nil, fmt.Errorf("new checkbutton: %w", err)
		}
		w.SetMarginEnd(4)
		fr.widget = w
	default:
		return nil, fmt.Errorf("unknown inputType: %v", field.inputType)
	}
	row.Add(fr.widget)

	return &fr, nil
}
