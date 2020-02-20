package tags

import "strconv"

type Tag struct {
	Name string
	Seen int

	num string
}

func (t *Tag) SeenCount() string {
	if t.num == "" {
		t.num = strconv.Itoa(t.Seen)
	}
	return t.num
}
