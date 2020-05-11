package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/gotk3/gotk3/gtk"
	"github.com/twitchyliquid64/diagg/form"
)

// Win encapsulates the UI state of the window.
type Win struct {
	win  *gtk.Window
	form *form.Form
	data *myForm
}

type myForm struct {
	Name    string `form:"width=30,validator=ValidateName"`
	Coolios bool   `form:"label='some input?'"`
	Age     int    `form:"label='age (since birth lol)' validator_func=ValidateAge"`
	ZZZ     uint16 `form:"explain='some nonsense'"`
	Select  string `form:"width=30,kind=combo,o=a,o='thing b'"`
}

func (f *myForm) ValidateName(input string) error {
	if strings.ToUpper(input) == "JOHN" {
		return errors.New("NOPE")
	}
	return nil
}

func (f *myForm) ValidateAge(input string) string {
	if strings.HasPrefix(input, "-") {
		return "Age can't be negative!"
	}
	return ""
}

func (w *Win) build() error {
	var err error

	if w.win, err = gtk.WindowNew(gtk.WINDOW_TOPLEVEL); err != nil {
		return err
	}
	w.win.SetTitle("Tags demo")
	w.win.Connect("destroy", func() {
		gtk.MainQuit()
	})
	w.win.SetPosition(gtk.WIN_POS_CENTER)
	//w.win.SetDefaultSize(800, 600)

	if w.form, err = form.Build(w.data); err != nil {
		return err
	}

	w.win.Add(w.form.UI())
	return nil
}

func makeWin() (*Win, error) {
	w := &Win{
		data: &myForm{
			Name:    "swiggity",
			Coolios: true,
			Age:     23,
			Select:  "thing b",
		},
	}
	if err := w.build(); err != nil {
		return nil, err
	}
	w.win.ShowAll()
	return w, nil
}

func main() {
	gtk.Init(nil)
	flag.Parse()

	w, err := makeWin()
	if err != nil {
		fmt.Fprintf(os.Stderr, "makeWin() failed: %v\n", err)
		os.Exit(1)
	}

	gtk.Main()
	fmt.Printf("Output: %+v\n", w.data)

	// d := &myForm{
	// 	Name:    "swiggity",
	// 	Coolios: true,
	// 	Age:     23,
	// 	Select:  "thing b",
	// }
	//
	// p, err := form.Popup("Other window!!", d)
	// if err != nil {
	// 	fmt.Fprintf(os.Stderr, "FormPopup() failed: %v\n", err)
	// 	os.Exit(1)
	// }
	// p.Run()
	//
	// fmt.Printf("Output: %+v\n", d)
}
