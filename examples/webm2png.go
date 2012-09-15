package main

import (
	"bufio"
	"code.google.com/p/ebml-go/webm"
	"code.google.com/p/vp8-go/vp8"
	"flag"
	"fmt"
	"image"
	"image/png"
	"io"
	"log"
	"os"
)

var in = flag.String("i", "", "Input file")
var out = flag.String("o", "", "Output prefix")

func main() {
	flag.Parse()
	r, err := os.Open(*in)
	if err != nil {
		log.Panic("unable to open file " + *in)
	}
	br := bufio.NewReader(r)
	var wm webm.WebM
	e, rest, err := webm.Parse(br, &wm)
	d := vp8.NewDecoder()
	i := 0
	for err == nil {
		t := make([]byte, 4, 4)
		io.ReadFull(e.R, t)
		d.Init(e.R, int(e.Size()))
		_, err = d.DecodeFrameHeader()
		var img image.Image
		img, err = d.DecodeFrame()
		if err == nil {
			path := fmt.Sprint(*out, i, ".png")
			i++
			var w io.Writer
			w, err = os.Create(path)
			if err != nil {
				log.Panic("unable to open file " + *in)
			}
			png.Encode(w, img)
		}
		_, err = e.ReadData()
		e, err = rest.Next()
	}
}
