package form

import (
	"sync"

	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/gtk"
)

var cssProvider *gtk.CssProvider
var globalMut sync.Mutex

func maybeInitCSS() error {
	globalMut.Lock()
	defer globalMut.Unlock()
	if cssProvider != nil {
		return nil
	}

	var err error
	cssProvider, err = gtk.CssProviderNew()
	if err != nil {
		return err
	}
	screen, err := gdk.ScreenGetDefault()
	if err != nil {
		return err
	}
	gtk.AddProviderForScreen(screen, cssProvider, gtk.STYLE_PROVIDER_PRIORITY_APPLICATION)
	return cssProvider.LoadFromData(formStyling)
}

const formStyling = `
.explain-text {
  opacity: 0.5;
  font-size: 0.8em;
}

.validation-error {
  color: @error_color;
  font-size: 0.8em;
}
`
