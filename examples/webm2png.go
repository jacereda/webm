package main

import (
	"code.google.com/p/ebml-go/common"
	"code.google.com/p/ffvp8-go/ffvp8"
	"flag"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"log"
	"os"
	"runtime"
)

var out = flag.String("o", "", "Output prefix")
var format = flag.String("f", "png", "Output format")

func jpegenc(w io.Writer, img image.Image) error {
	return jpeg.Encode(w, img, nil)
}

func pngenc(w io.Writer, img image.Image) error {
	return png.Encode(w, img)
}

func write(ch <-chan *ffvp8.Frame) {
	var enc func(io.Writer, image.Image) error
	switch *format {
	case "jpeg", "jpg":
		enc = jpegenc
	case "png":
		enc = pngenc
	default:
		log.Panic("Unsupported output format")
	}
	i := 0
	for img := range ch {
		if *out != "" {
			path := fmt.Sprint(*out, i, ".", *format)
			f, err := os.Create(path)
			if err != nil {
				log.Panic("unable to open file " + *out)
			}
			enc(f, img)
			f.Close()
			runtime.GC()
		}
		i++
	}
}

func main() {
	flag.Parse()
	common.Main(write, nil)
}
