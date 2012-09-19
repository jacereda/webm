package main

import (
	"bufio"
	"code.google.com/p/ebml-go/webm"
	"code.google.com/p/ffvp8-go/ffvp8"
	"flag"
	"fmt"
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
	dec := ffvp8.NewDecoder()
	i := 0
	var track webm.TrackEntry
	for _,track = range wm.Segment.Tracks.TrackEntry {
		if track.IsVideo() {
			break
		}
	}

	for err == nil {
		t := make([]byte, 4)
		io.ReadFull(e.R, t)
		if uint(t[0]) & 0x7f == track.TrackNumber {
			data := make([]byte, e.Size())
			io.ReadFull(e.R, data)
			img := dec.Decode(data)
			path := fmt.Sprint(*out, i, ".png")
			i++
			f, err := os.Create(path)
			if err != nil {
				log.Panic("unable to open file " + *in)
			}
			png.Encode(f, img)
			f.Close()
		}
		_, err = e.ReadData()
		e, err = rest.Next()
	}
}
