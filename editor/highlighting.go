package editor

import (
	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/gtk"
)

type tagSet struct {
	str     *gtk.TextTag
	keyword *gtk.TextTag
	parenth *gtk.TextTag
	op      *gtk.TextTag
	fun     *gtk.TextTag
	pseudo  *gtk.TextTag
	comment *gtk.TextTag
	field   *gtk.TextTag

	heading     *gtk.TextTag
	inlineBlock *gtk.TextTag
	nameTag     *gtk.TextTag
	nameAttr    *gtk.TextTag
}

func makeStyling(buffer *gtk.TextBuffer, bg *gdk.RGBA) (*gtk.CssProvider, *tagSet, error) {
	s, err := gtk.CssProviderNew()
	if err != nil {
		return nil, nil, err
	}
	s.LoadFromData(`
		GtkTextView {
		    font-family: monospace;
		}
		textview {
		    font-family: monospace;
		}
    `)

	var strTag, fun, btk, nt, na *gtk.TextTag
	if f := bg.Floats(); f[0] > 0.75 && f[1] > 0.75 && f[2] > 0.75 { // light background
		strTag = buffer.CreateTag("string", map[string]interface{}{
			"foreground": "#aa00aa",
		})
		fun = buffer.CreateTag("func", map[string]interface{}{
			"foreground": "#211fd4",
		})
		btk = buffer.CreateTag("inlineBlock", map[string]interface{}{
			"foreground": "#589339",
		})
		nt = buffer.CreateTag("nameTag", map[string]interface{}{
			"foreground": "#418fcf",
		})
		na = buffer.CreateTag("nameAttr", map[string]interface{}{
			"foreground": "#3696a2",
		})
	} else { // dark background / theme
		strTag = buffer.CreateTag("string", map[string]interface{}{
			"foreground": "#98c379",
		})
		fun = buffer.CreateTag("func", map[string]interface{}{
			"foreground": "#61afef",
		})
		btk = buffer.CreateTag("inlineBlock", map[string]interface{}{
			"foreground": "#98c379",
		})
		nt = buffer.CreateTag("nameTag", map[string]interface{}{
			"foreground": "#61afef",
		})
		na = buffer.CreateTag("nameAttr", map[string]interface{}{
			"foreground": "#56b6c2",
		})
	}

	// function blue: #61afef
	// type green: #56b6c2
	// cool magenta: #c678dd
	// comment grey: #5c6370
	// heading: #e06c75
	// backtick: #98c379
	// link tag: #61afef
	// link href: #56b6c2

	return s, &tagSet{
		str: strTag,
		keyword: buffer.CreateTag("keyword", map[string]interface{}{
			"foreground": "orange",
		}),
		parenth: buffer.CreateTag("parenth", map[string]interface{}{
			//"foreground": "cyan",
			//"weight": pango.WEIGHT_BOLD,
		}),
		op: buffer.CreateTag("op", map[string]interface{}{
			"foreground": "red",
		}),
		field: buffer.CreateTag("field", map[string]interface{}{
			"foreground": "red",
		}),
		fun: fun,
		pseudo: buffer.CreateTag("pseudo", map[string]interface{}{
			"foreground": "#c678dd",
		}),
		comment: buffer.CreateTag("comment", map[string]interface{}{
			"foreground": "#5c6370",
		}),
		heading: buffer.CreateTag("heading", map[string]interface{}{
			"foreground": "#e06c75",
		}),
		inlineBlock: btk,
		nameTag:     nt,
		nameAttr:    na,
	}, nil
}
