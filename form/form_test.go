package form

import (
	"reflect"
	"testing"
)

func TestInterpretStruct(t *testing.T) {
	tcs := []struct {
		name string
		in   interface{}
		want formDef
	}{
		{
			"empty",
			&struct{}{},
			formDef{fields: make([]*formField, 0)},
		},
		{
			"basic",
			struct {
				Name    string `form:"yeet label=moose"`
				Skipped string `form:"-"`
				Alt     string `form:"lazy label"`
				Quoted  int    `form:"label='b'"`
			}{},
			formDef{
				fields: []*formField{
					{label: "moose", tagSpec: tagSpec{label: "moose", terms: []string{"yeet"}}},
					{label: "lazy label", tagSpec: tagSpec{terms: []string{"lazy", "label"}}},
					{label: "b", tagSpec: tagSpec{label: "b"}, inputType: InputInt},
				},
			},
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			got, err := interpretStruct(tc.in)
			if err != nil {
				t.Fatalf("interpretStruct() failed: %v", err)
			}

			// We can't really check the 'field' so zero it out.
			for i := range got.fields {
				got.fields[i].field = reflect.Value{}
				got.fields[i].fieldType = reflect.StructField{}
			}

			if !reflect.DeepEqual(got, &tc.want) {
				t.Errorf("interpretStruct() = %v, want %v", got, tc.want)
			}
		})
	}
}
