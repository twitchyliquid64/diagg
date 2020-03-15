package editor

import (
	"fmt"
	"time"

	"github.com/alecthomas/chroma"
	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
)

// Editor manages a text editor with syntax highlighting.
type Editor struct {
	editor *gtk.TextView
	buffer *gtk.TextBuffer

	bg    *gdk.RGBA
	css   *gtk.CssProvider
	style *gtk.StyleContext
	tags  *tagSet

	lexer    chroma.Lexer
	changeCB func()
}

// New creates a new, blank editing widge.
func New(lexer chroma.Lexer, changeCB func()) (*Editor, error) {
	return NewWithContent(lexer, changeCB, "")
}

// NewWithContent creates a new editing widget, initialized to the provided content.
func NewWithContent(lexer chroma.Lexer, changeCB func(), content string) (*Editor, error) {
	editor, err := gtk.TextViewNew()
	if err != nil {
		return nil, err
	}
	editor.SetVExpand(true)
	editor.SetHExpand(true)

	buffer, err := editor.GetBuffer()
	if err != nil {
		return nil, err
	}
	buffer.SetText(content)

	style, err := editor.GetStyleContext()
	if err != nil {
		return nil, err
	}

	bg, err := style.GetProperty("background-color", gtk.STATE_FLAG_BACKDROP)
	if err != nil {
		return nil, fmt.Errorf("GetProperty('background-color') failed: %v", err)
	}

	css, tags, err := makeStyling(buffer, bg.(*gdk.RGBA))
	if err != nil {
		return nil, err
	}
	style.AddProvider(css, gtk.STYLE_PROVIDER_PRIORITY_APPLICATION)

	out := &Editor{
		editor:   editor,
		buffer:   buffer,
		css:      css,
		style:    style,
		tags:     tags,
		bg:       bg.(*gdk.RGBA),
		changeCB: changeCB,
		lexer:    lexer,
	}

	if _, err := out.buffer.Connect("insert-text", func(tb *gtk.TextBuffer, loc *gtk.TextIter, ins string, len int, e *Editor) {
		e.onInsert(ins, tb, loc, len)
	}, out); err != nil {
		return nil, err
	}
	if _, err := out.editor.Connect("backspace", out.onBackspace, out); err != nil {
		return nil, err
	}
	if _, err := out.editor.Connect("paste-clipboard", func() {
		go func() {
			time.Sleep(time.Millisecond * 30)
			glib.IdleAdd(func() {
				out.Restyle()
			})
		}()
	}, out); err != nil {
		return nil, err
	}

	glib.IdleAdd(func() {
		out.Restyle()
	})

	return out, nil
}

func (e *Editor) UI() *gtk.TextView {
	return e.editor
}

// Restyle styles the editor from scratch.
func (e *Editor) Restyle() {
	for i := 0; i < e.buffer.GetLineCount(); i++ {
		start, end := e.buffer.GetIterAtLine(i), e.buffer.GetIterAtLine(i+1)
		content := start.GetSlice(end)
		tok, _ := e.lexer.Tokenise(nil, content+"\n")
		e.processLine(tok, content, e.buffer, start, end, i)
	}
}

// SetContent changes the contents of the editor.
func (e *Editor) SetContent(content string) {
	e.buffer.SetText(content)
	e.Restyle()
}

// Focus forces the editor to take focus.
func (e *Editor) Focus() {
	e.editor.GrabFocus()
}

// GetContent returns the contents of the editor.
func (e *Editor) GetContent() string {
	t, _ := e.buffer.GetText(e.buffer.GetStartIter(), e.buffer.GetEndIter(), false)
	return t
}

func (e *Editor) processLine(tok func() chroma.Token, content string, tb *gtk.TextBuffer, start, end *gtk.TextIter, line int) {
	tb.RemoveAllTags(start, end)

	var (
		lineOffset      int
		lastType        chroma.Token
		lastStartOffset int
	)
	for t := tok(); t != chroma.EOF; t = tok() {
		// fmt.Println(t)

		switch t.Type {
		case chroma.Operator:
			if t.Value != "." {
				tb.ApplyTag(e.tags.op, tb.GetIterAtLineOffset(line, lineOffset), tb.GetIterAtLineOffset(line, lineOffset+len(t.Value)))
			}
		case chroma.String:
			tb.ApplyTag(e.tags.str, tb.GetIterAtLineOffset(line, lineOffset), tb.GetIterAtLineOffset(line, lineOffset+len(t.Value)))
		case chroma.Keyword:
			tb.ApplyTag(e.tags.keyword, tb.GetIterAtLineOffset(line, lineOffset), tb.GetIterAtLineOffset(line, lineOffset+len(t.Value)))
		case chroma.NameFunction:
			tb.ApplyTag(e.tags.fun, tb.GetIterAtLineOffset(line, lineOffset), tb.GetIterAtLineOffset(line, lineOffset+len(t.Value)))
		case chroma.NameBuiltinPseudo:
			tb.ApplyTag(e.tags.pseudo, tb.GetIterAtLineOffset(line, lineOffset), tb.GetIterAtLineOffset(line, lineOffset+len(t.Value)))
		case chroma.CommentSingle:
			tb.ApplyTag(e.tags.comment, tb.GetIterAtLineOffset(line, lineOffset), tb.GetIterAtLineOffset(line, lineOffset+len(t.Value)))
		case chroma.LiteralStringDouble, chroma.LiteralStringAffix:
			tb.ApplyTag(e.tags.str, tb.GetIterAtLineOffset(line, lineOffset), tb.GetIterAtLineOffset(line, lineOffset+len(t.Value)))
		case chroma.GenericHeading, chroma.GenericSubheading:
			tb.ApplyTag(e.tags.heading, tb.GetIterAtLineOffset(line, lineOffset), tb.GetIterAtLineOffset(line, lineOffset+len(t.Value)))
		case chroma.LiteralStringBacktick:
			tb.ApplyTag(e.tags.inlineBlock, tb.GetIterAtLineOffset(line, lineOffset), tb.GetIterAtLineOffset(line, lineOffset+len(t.Value)))
		case chroma.NameTag:
			tb.ApplyTag(e.tags.nameTag, tb.GetIterAtLineOffset(line, lineOffset), tb.GetIterAtLineOffset(line, lineOffset+len(t.Value)))
		case chroma.NameAttribute:
			tb.ApplyTag(e.tags.nameAttr, tb.GetIterAtLineOffset(line, lineOffset), tb.GetIterAtLineOffset(line, lineOffset+len(t.Value)))
		case chroma.Punctuation:
			switch t.Value {
			case "(":
				if lastType.Type == chroma.Name {
					tb.ApplyTag(e.tags.fun, tb.GetIterAtLineOffset(line, lastStartOffset), tb.GetIterAtLineOffset(line, lineOffset))
				}
				fallthrough
			case ")":
				tb.ApplyTag(e.tags.parenth, tb.GetIterAtLineOffset(line, lineOffset), tb.GetIterAtLineOffset(line, lineOffset+len(t.Value)))
			}
		case chroma.Name:
			if lastType.Type == chroma.Operator && lastType.Value == "." {
				tb.ApplyTag(e.tags.field, tb.GetIterAtLineOffset(line, lineOffset), tb.GetIterAtLineOffset(line, lineOffset+len(t.Value)))
			}
		}

		lastStartOffset = lineOffset
		lineOffset += len(t.Value)
		lastType = t
	}
}

func (e *Editor) onAnyChange() {
	if e.changeCB != nil {
		e.changeCB()
	}
}

func (e *Editor) onBackspace() {
	defer e.onAnyChange()
	iter := e.buffer.GetIterAtMark(e.buffer.GetMark("insert"))
	line := iter.GetLine()
	start, end := e.buffer.GetIterAtLine(line), e.buffer.GetIterAtLine(line+1)
	content := start.GetSlice(end)

	// Do backspace
	if len(content) > 0 {
		content = content[:len(content)-1]
	}

	tok, _ := e.lexer.Tokenise(nil, content+"\n")
	e.processLine(tok, content, e.buffer, start, end, line)
}

func (e *Editor) onInsert(text string, tb *gtk.TextBuffer, loc *gtk.TextIter, l int) {
	defer e.onAnyChange()
	//fmt.Printf("Insert: line %d, pos %d, char %q \n", loc.GetLine(), loc.GetCharsInLine(), text)
	line := loc.GetLine()

	// Schedule a re-tag of the current line.
	glib.IdleAdd(func() {
		start, end := tb.GetIterAtLine(line), tb.GetIterAtLine(line+1)
		content := start.GetSlice(end) + text
		tok, _ := e.lexer.Tokenise(nil, content+"\n")
		e.processLine(tok, content, tb, start, end, line)
	})
}
