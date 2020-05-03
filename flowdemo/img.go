package main

import (
	"bytes"
	"encoding/base64"
	"image"
	"image/color"
	"image/png"

	"github.com/gotk3/gotk3/gdk"
	"golang.org/x/image/draw"
)

// AddButtonImg computes a plus button icon at a specific size.
func AddButtonImg(x, y int) *image.RGBA {
	b, err := base64.StdEncoding.DecodeString("iVBORw0KGgoAAAANSUhEUgAAACQAAAAkCAYAAADhAJiYAAAAT0lEQVR4Ae2WAQYAMAhFd7SO9o/WzbYQBmSExXs8AF4SLQCAdyxUausDFO5UBBFEEEHNWKhCv4I8VKF1Td+lBgQNXhlXRlANQQTx5APAbA5+KXS1P2kTZAAAAABJRU5ErkJggg==")
	if err != nil {
		panic(err)
	}
	return decodeAndFormatImg(b, x, y)
}

// CloseButtonImg computes a close button icon.
func CloseButtonImg(x, y int) *image.RGBA {
	b, err := base64.StdEncoding.DecodeString("iVBORw0KGgoAAAANSUhEUgAAACQAAAAkCAQAAABLCVATAAAAoklEQVR4Ae2UtQEDMQDE5LHCsE26nyBQfenuS1dZM8yscGW5M8h8nJPJ9Ag8ItDFoWBBeqAKJOYMENosNiVxj7TrUUcodx0rblHtWksc4l3VXhNxzlTxLY0MiqoRVSnbPaL3M5ILUMJeZU/CCUxPNOETUXLRH7e210yY+GHj1+8f53cPUga5yj+t//13Y2SC0DnXeLBZ1CJRW3wn/PucksksATXJlBc3KdoMAAAAAElFTkSuQmCC")
	if err != nil {
		panic(err)
	}
	return decodeAndFormatImg(b, x, y)
}

// TabImg computes a tab icon at a specific sizer.
func TabImg(x, y int) *image.RGBA {
	b, err := base64.StdEncoding.DecodeString("iVBORw0KGgoAAAANSUhEUgAAACQAAAAkCAYAAADhAJiYAAAAfklEQVR4Ae3XAQaAQBCF4f9mBaA60QZzgaIbdbMCGBH0srt4PwMGPhI7OPdzA1CAqDQFGHjpAK5Gc/BoTstWs5Ba0+IUPsEpgFZSkRbB90IARY8ggwwahb9SB+kRlUAGGWSQQQYZZFDRH/nyUVBITR2cQROP9oaYre9T2jmhG7X1gipdFrDkAAAAAElFTkSuQmCC")
	if err != nil {
		panic(err)
	}
	return decodeAndFormatImg(b, x, y)
}

func decodeAndFormatImg(buf []byte, x, y int) *image.RGBA {
	i, err := png.Decode(bytes.NewBuffer(buf))
	if err != nil {
		panic(err)
	}

	outSize := image.Rect(0, 0, x, y)
	img := image.NewRGBA(outSize)
	draw.BiLinear.Scale(img, outSize, i, i.Bounds(), draw.Over, nil)
	return img
}

// GDKColorToRGBA converts a GDK color to a Go one.
func GDKColorToRGBA(in *gdk.RGBA) color.RGBA {
	c := in.Floats()
	return color.RGBA{
		R: uint8(c[0] * 255),
		G: uint8(c[1] * 255),
		B: uint8(c[2] * 255),
	}
}

func binaryImage(rawImg *image.RGBA) *gdk.Pixbuf {
	bb, err := gdk.PixbufNew(gdk.COLORSPACE_RGB, true, 8, rawImg.Bounds().Dx(), rawImg.Bounds().Dy())
	if err != nil {
		panic(err)
	}
	buf := bb.GetPixels()
	for i := 0; i < len(rawImg.Pix); i += 4 {
		a := rawImg.Pix[i+3]
		var v byte = 0
		if a > 220 {
			v = 255
		}

		buf[i] = v
		buf[i+1] = v
		buf[i+2] = v
		buf[i+3] = v
	}
	return bb
}
