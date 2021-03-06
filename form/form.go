// Package form generates forms based on structs.
package form

import (
	"errors"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"

	"github.com/gotk3/gotk3/gtk"
)

// Dialog represents a form within a window.
type Dialog struct {
	Form *Form
	win  *gtk.Dialog
}

// Run blocks until the dialog is closed.
func (w *Dialog) Run() gtk.ResponseType {
	for {
		result := w.win.Run()
		switch result {
		case gtk.RESPONSE_CANCEL, gtk.RESPONSE_REJECT, gtk.RESPONSE_CLOSE, gtk.RESPONSE_DELETE_EVENT:
			w.Form.noSave = true
		default:
			if !w.Form.Valid() {
				continue
			}
			w.Form.noSave = false
		}
		w.win.Destroy()
		return result
	}
}

func (w *Dialog) closePressed() bool {
	w.win.Destroy()
	return false
}

// Popup creates a dialog containing a form.
func Popup(title string, s interface{}) (*Dialog, error) {
	f, err := Build(s)
	if err != nil {
		return nil, err
	}

	w := Dialog{
		Form: f,
	}

	if w.win, err = gtk.DialogNew(); err != nil {
		return nil, err
	}
	w.win.SetTitle(title)
	b, _ := w.win.AddButton("Save", gtk.RESPONSE_ACCEPT)
	w.win.AddButton("Cancel", gtk.RESPONSE_CANCEL)
	w.win.Connect("destroy", w.closePressed)
	w.win.SetPosition(gtk.WIN_POS_CENTER)

	ca, err := w.win.GetContentArea()
	if err != nil {
		return nil, err
	}
	for _, row := range f.fields {
		if entry, isEntry := row.widget.(*gtk.Entry); isEntry {
			entry.Connect("activate", func() {
				b.Activate()
			})
		}
	}
	ca.Add(f.UI())
	w.win.ShowAll()
	return &w, nil
}

// Form represents a generated set of UI inputs.
type Form struct {
	noSave bool
	box    *gtk.Box
	spec   *formDef
	fields []*formRow
}

func (f *Form) Valid() bool {
	for _, r := range f.fields {
		if r.validationFailed {
			return false
		}
	}
	return true
}

func (f *Form) UI() *gtk.Box {
	return f.box
}

func (f *Form) onDestroy() bool {
	if f.noSave {
		return false
	}

	for i, field := range f.fields {
		if err := field.spec.inputType.Apply(field, field.spec.field); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to apply value from field %d: %v\n", i, err)
		}
	}
	return false
}

// Build constructs a form using GTK elements.
func Build(s interface{}) (*Form, error) {
	if err := maybeInitCSS(); err != nil {
		return nil, err
	}

	spec, err := interpretStruct(s)
	if err != nil {
		return nil, err
	}
	out := Form{
		spec:   spec,
		fields: make([]*formRow, 0, len(spec.fields)),
	}
	if out.box, err = gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 2); err != nil {
		return nil, err
	}
	out.box.Connect("destroy", out.onDestroy)

	for i, field := range spec.fields {
		row, err := makeRow(field)
		if err != nil {
			return nil, fmt.Errorf("failed making row %d: %w", i, err)
		}
		out.box.Add(row.box)
		if err := field.inputType.Populate(row, field.field); err != nil {
			return nil, fmt.Errorf("populating initial value for row %d: %w", i, err)
		}
		out.fields = append(out.fields, row)
	}

	return &out, nil
}

// formField describes a single field within a form.
type formField struct {
	tagSpec tagSpec
	label   string
	explain string

	field     reflect.Value
	fieldType reflect.StructField

	inputType       inputType
	customValidator *reflect.Value
	comboFunc       *reflect.Value
}

// formDef describes a form.
type formDef struct {
	fields []*formField
}

func interpretStruct(s interface{}) (*formDef, error) {
	v := reflect.ValueOf(s)
	if v.Type().Kind() == reflect.Ptr && v.IsNil() {
		return nil, errors.New("nil pointer provided")
	}
	if v.Kind() == reflect.Ptr || v.Kind() == reflect.Interface {
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return nil, fmt.Errorf("only structs supported but %v provided", v)
	}

	t := v.Type()
	fields := make([]*formField, 0, t.NumField())
	for i := 0; i < t.NumField(); i++ {
		//f := v.Field(i)
		tags := t.Field(i).Tag.Get("form")
		if tags == "-" {
			continue
		}
		ts := parseTags(tags)

		ff := formField{
			tagSpec:   ts,
			label:     ts.Label(),
			explain:   ts.explain,
			field:     v.Field(i),
			fieldType: t.Field(i),
		}
		if ff.label == "" {
			ff.label = t.Field(i).Name
		}

		switch ff.fieldType.Type.Kind() {
		case reflect.Bool:
			ff.inputType = InputBool
		case reflect.String:
			ff.inputType = InputText
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			ff.inputType = InputInt
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			ff.inputType = InputUint
		}

		if ts.validatorFunc != "" {
			if ff.inputType == InputBool || ff.tagSpec.kind == kindCombo {
				return nil, errors.New("custom validators are not supported on bool types or combo fields")
			}
			mbn := reflect.ValueOf(s).MethodByName(ts.validatorFunc)
			ff.customValidator = &mbn
		}
		if ts.comboFunc != "" {
			if ff.tagSpec.kind != kindCombo {
				return nil, errors.New("combo func defined on non-combo field")
			}
			f := reflect.ValueOf(s).MethodByName(ts.comboFunc)
			ff.comboFunc = &f
		}

		fields = append(fields, &ff)
	}
	return &formDef{
		fields: fields,
	}, nil
}

type nextTermType uint8

const (
	termNormal nextTermType = iota
	termLabel
	termExplain
	termWidth
	termValidatorFunc
	termKind
	termOption
	termComboFunc
)

type fieldKind uint8

const (
	kindDefault fieldKind = iota // entry for string, toggle for bool.
	kindCombo
)

type tagSpec struct {
	label   string
	explain string
	kind    fieldKind

	validatorFunc string
	comboFunc     string

	comboOptions []string
	width        int

	terms []string
}

func (s *tagSpec) Label() string {
	if s.label != "" {
		return s.label
	}
	return strings.Join(s.terms, " ")
}

func (s *tagSpec) push(term string, kind nextTermType) {
	if term == "" {
		return
	}

	switch kind {
	case termLabel:
		s.label = term
	case termNormal:
		s.terms = append(s.terms, term)
	case termExplain:
		s.explain = term
	case termWidth:
		s.width, _ = strconv.Atoi(term)
	case termValidatorFunc:
		s.validatorFunc = term
	case termComboFunc:
		s.comboFunc = term
	case termKind:
		switch term {
		case "combo":
			s.kind = kindCombo
		}
	case termOption:
		s.comboOptions = append(s.comboOptions, term)
	}
}

func parseTags(inp string) tagSpec {
	var (
		nextTerm    nextTermType
		inQuotes    = false
		quoteChar   = '\''
		accumulator string
		out         tagSpec
	)

	for _, c := range inp {
		switch {
		case inQuotes && c == quoteChar: // Terminating quote reached
			out.push(accumulator, nextTerm)
			accumulator = ""
			nextTerm = termNormal
			inQuotes = false

		case inQuotes: // Still in a quoted term
			accumulator = accumulator + string(c)

		case !inQuotes && c == '\'': // New quoted term
			inQuotes = true
			quoteChar = '\''

		case !inQuotes && (c == ',' || c == ' '): // End of term
			out.push(accumulator, nextTerm)
			accumulator = ""
			nextTerm = termNormal

		default:
			accumulator = accumulator + string(c)
			switch accumulator {
			case "label=":
				nextTerm = termLabel
				accumulator = ""
			case "explain=":
				nextTerm = termExplain
				accumulator = ""
			case "width=":
				nextTerm = termWidth
				accumulator = ""
			case "validator=", "validator_func=":
				nextTerm = termValidatorFunc
				accumulator = ""
			case "combo=", "combo_func=":
				nextTerm = termComboFunc
				accumulator = ""
			case "control=", "kind=":
				nextTerm = termKind
				accumulator = ""
			case "o=", "option=":
				nextTerm = termOption
				accumulator = ""
			}
		}
	}
	out.push(accumulator, nextTerm)
	return out
}
