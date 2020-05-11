package form

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/gotk3/gotk3/gtk"
)

type inputType uint8

func setCombo(r *formRow, c *gtk.ComboBox, val string) error {
	if val == "" {
		return nil
	}
	l, ok := r.comboModel.GetIterFirst()
	if !ok {
		return nil
	}

	for i := 0; true; i++ {
		v, err := r.comboModel.GetValue(l, 0)
		if err != nil {
			return err
		}
		s, err := v.GetString()
		if err != nil {
			return err
		}
		if s == val {
			c.SetActive(i)
			return nil
		}
		if !r.comboModel.IterNext(l) {
			break
		}
	}

	return fmt.Errorf("no combo option %q", val)
}

func applyCombo(r *formRow, c *gtk.ComboBox, val reflect.Value) error {
	iter, err := c.GetActiveIter()
	if err != nil {
		return err
	}
	gv, err := r.comboModel.GetValue(iter, 0)
	if err != nil {
		return err
	}
	s, err := gv.GoValue()
	if err != nil {
		return err
	}
	val.SetString(s.(string))
	return nil
}

func (i inputType) Populate(r *formRow, val reflect.Value) error {
	w := r.widget
	switch i {
	case InputText:
		if c, isCombo := w.(*gtk.ComboBox); isCombo {
			return setCombo(r, c, val.String())
		}
		w.(*gtk.Entry).SetText(val.String())
	case InputBool:
		w.(*gtk.Switch).SetActive(val.Bool())
	case InputInt:
		w.(*gtk.Entry).SetText(fmt.Sprint(val.Int()))
	case InputUint:
		w.(*gtk.Entry).SetText(fmt.Sprint(val.Uint()))
	default:
		return fmt.Errorf("unknown input type: %v", i)
	}

	return nil
}

func (i inputType) Apply(r *formRow, val reflect.Value) error {
	w := r.widget
	switch i {
	case InputText:
		if c, isCombo := w.(*gtk.ComboBox); isCombo {
			return applyCombo(r, c, val)
		}
		t, _ := w.(*gtk.Entry).GetText()
		val.SetString(t)
	case InputBool:
		val.SetBool(w.(*gtk.Switch).GetActive())
	case InputInt:
		t, _ := w.(*gtk.Entry).GetText()
		num, _ := strconv.ParseInt(t, 10, 64)
		val.SetInt(num)
	case InputUint:
		t, _ := w.(*gtk.Entry).GetText()
		num, _ := strconv.ParseUint(t, 10, 64)
		val.SetUint(num)
	default:
		return fmt.Errorf("unknown input type: %v", i)
	}

	return nil
}

func (i inputType) Validate(contents string) error {
	switch i {
	case InputText:
		return nil
	case InputBool:
		if c := strings.ToUpper(contents); c != "TRUE" && c != "FALSE" {
			return errors.New("invalid input: expected true or false")
		}
	case InputInt:
		_, err := strconv.ParseInt(contents, 10, 64)
		return err
	case InputUint:
		_, err := strconv.ParseUint(contents, 10, 64)
		return err
	default:
		return fmt.Errorf("unknown input type: %v", i)
	}

	return nil
}

// Valid input types.
const (
	InputText inputType = iota
	InputBool
	InputInt
	InputUint
)
