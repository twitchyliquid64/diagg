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

func (i inputType) Populate(w gtk.IWidget, val reflect.Value) error {
	switch i {
	case InputText:
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

func (i inputType) Apply(w gtk.IWidget, val reflect.Value) error {
	switch i {
	case InputText:
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
