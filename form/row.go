package form

import (
	"fmt"
	"strconv"

	"github.com/gotk3/gotk3/gtk"
)

// formRow describes a live representation of a form field.
type formRow struct {
	box     *gtk.Box
	tb      *gtk.Box
	errText *gtk.Label

	widget gtk.IWidget
	spec   *formField
}

func (r *formRow) SetValidationText(text string) error {
	if text == "" {
		if r.errText != nil {
			r.tb.Remove(r.errText)
			r.errText = nil
		}
		return nil
	}

	if r.errText != nil {
		r.errText.SetText(text)
	} else {
		var err error
		if r.errText, err = gtk.LabelNew(text); err != nil {
			return err
		}
		r.errText.SetHAlign(gtk.ALIGN_START)
		r.errText.SetMarginStart(4)
		s, _ := r.errText.GetStyleContext()
		s.AddClass("validation-error")
		r.errText.Show()
		r.tb.Add(r.errText)
	}
	return nil
}

func (r *formRow) onEntryChanged() {
	t, _ := r.widget.(*gtk.Entry).GetText()
	err := r.spec.inputType.Validate(t)
	if scErr, ok := err.(*strconv.NumError); ok {
		err = scErr.Err
	}

	if err != nil {
		r.SetValidationText(err.Error())
	} else {
		r.SetValidationText("")
	}
}

func makeRow(field *formField) (*formRow, error) {
	row, err := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 2)
	if err != nil {
		return nil, err
	}
	row.SetMarginTop(4)
	tb, err := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 0)
	if err != nil {
		return nil, err
	}
	tb.SetHExpand(true)
	tb.SetHAlign(gtk.ALIGN_START)
	tb.SetVAlign(gtk.ALIGN_CENTER)

	lab, err := gtk.LabelNew(field.label)
	if err != nil {
		return nil, err
	}
	lab.SetHExpand(true)
	lab.SetHAlign(gtk.ALIGN_START)
	lab.SetXAlign(0)
	lab.SetJustify(gtk.JUSTIFY_LEFT)
	lab.SetMarginStart(4)
	tb.Add(lab)

	if field.explain != "" {
		e, err := gtk.LabelNew(field.explain)
		if err != nil {
			return nil, err
		}
		e.SetHExpand(true)
		e.SetHAlign(gtk.ALIGN_START)
		e.SetXAlign(0)
		e.SetJustify(gtk.JUSTIFY_LEFT)
		e.SetMarginStart(4)
		s, _ := e.GetStyleContext()
		s.AddClass("explain-text")
		tb.Add(e)
	}
	row.Add(tb)

	fr := formRow{
		box:  row,
		spec: field,
		tb:   tb,
	}
	switch fr.spec.inputType {
	case InputText, InputInt, InputUint:
		var w *gtk.Entry
		if w, err = gtk.EntryNew(); err != nil {
			return nil, fmt.Errorf("new entry: %w", err)
		}
		w.SetMarginEnd(4)
		w.Connect("changed", fr.onEntryChanged)
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
