package form

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"

	"github.com/gotk3/gotk3/gtk"
)

var errType = reflect.TypeOf((*error)(nil)).Elem()

// formRow describes a live representation of a form field.
type formRow struct {
	box     *gtk.Box
	tb      *gtk.Box
	errText *gtk.Label

	widget           gtk.IWidget
	spec             *formField
	validationFailed bool
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

	err := r.validateText(t)

	if err != nil {
		r.SetValidationText(err.Error())
	} else {
		r.SetValidationText("")
	}
}

func (r *formRow) validateText(t string) error {
	var err error
	if r.spec.customValidator != nil {
		out := r.spec.customValidator.Call([]reflect.Value{reflect.ValueOf(t)})
		if len(out) > 0 {
			switch out[0].Kind() {
			case reflect.Bool:
				if !out[0].Bool() {
					err = errors.New("validation failed")
				}
			case reflect.String:
				if s := out[0].String(); s != "" {
					err = errors.New(s)
				}
			case reflect.Interface:
				if out[0].Type().Implements(errType) && !out[0].IsNil() {
					err = out[0].Interface().(error)
				}
			}
		}
	}

	if err == nil {
		err = r.spec.inputType.Validate(t)
		if scErr, ok := err.(*strconv.NumError); ok {
			err = scErr.Err
		}
	}
	r.validationFailed = err != nil
	return err
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

	fr := &formRow{
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
		if field.tagSpec.width > 0 {
			w.SetWidthChars(field.tagSpec.width)
		}
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

	return fr, nil
}
