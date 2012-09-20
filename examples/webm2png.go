package main

import (
	"bufio"
	"code.google.com/p/ebml-go/webm"
	"code.google.com/p/ffvp8-go/ffvp8"
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
var nf = flag.Int("n", 0x7fffffff, "Number of frames")

func decode(ch chan []byte, wch chan *image.YCbCr) {
	dec := ffvp8.NewDecoder()
	for data := <-ch; data != nil; data = <-ch {
		wch <- dec.Decode(data)
	}
	wch <- nil
}

func write(ch chan *image.YCbCr, ech chan int) {
	for i, img := 0, <-ch; img != nil; i, img = i+1, <-ch {
		if *out != "" {
			path := fmt.Sprint(*out, i, ".png")
			f, err := os.Create(path)
			if err != nil {
				log.Panic("unable to open file " + *in)
			}
			png.Encode(f, img)
			f.Close()
		}
	}
	ech <- 1
}

func main() {
	flag.Parse()
	r, err := os.Open(*in)
	if err != nil {
		log.Panic("unable to open file " + *in)
	}
	br := bufio.NewReader(r)
	var wm webm.WebM
	e, rest, err := webm.Parse(br, &wm)
	var track webm.TrackEntry
	for _, track = range wm.Segment.Tracks.TrackEntry {
		if track.IsVideo() {
			break
		}
	}
	dchan := make(chan []byte, 16)
	wchan := make(chan *image.YCbCr, 16)
	echan := make(chan int, 1)
	go decode(dchan, wchan)
	go write(wchan, echan)
	fmt.Println("NF", *nf)
	for i := 0; err == nil && i < *nf; {
		t := make([]byte, 4)
		io.ReadFull(e.R, t)
		if uint(t[0])&0x7f == track.TrackNumber {
			data := make([]byte, e.Size())
			io.ReadFull(e.R, data)
			dchan <- data
			i++
		}
		_, err = e.ReadData()
		e, err = rest.Next()
	}
	dchan <- nil
	<-echan
}
