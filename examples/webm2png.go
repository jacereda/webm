package main

import (
	"bufio"
	"code.google.com/p/ebml-go/webm"
	"code.google.com/p/ffvp8-go/ffvp8"
	"flag"
	"fmt"
	"image"
	"image/png"
	"log"
	"os"
)

var (
	in = flag.String("i", "", "Input file")
	out = flag.String("o", "", "Output prefix")
	nf = flag.Int("n", 0x7fffffff, "Number of frames")
)

const chancap = 0

func decode(dchan chan webm.Packet, wchan chan *image.YCbCr) {
	dec := ffvp8.NewDecoder()
	for pkt := <- dchan; !pkt.IsLast(); pkt = <- dchan {
		img := dec.Decode(pkt.Data)
		if !pkt.Invisible {
			wchan <- img
		}
	}
	wchan <- nil
}

func write(ch chan *image.YCbCr) {
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
}

func read(dchan chan webm.Packet) {
	var err error
	var wm webm.WebM
	r, err := os.Open(*in)
	defer r.Close()
	if err != nil {
		log.Panic("unable to open file " + *in)
	}
	br := bufio.NewReader(r)
	pchan := webm.Parse(br, &wm)
	track := wm.FindFirstVideoTrack()
	for i := 0; err == nil && i < *nf; {
		pkt := <- pchan
		if pkt.IsLast() {
			break
		}
		if pkt.TrackNumber == track.TrackNumber {
			dchan <- pkt
			i++
		}
	}
	dchan <- webm.Last()
}

func main() {
	flag.Parse()
	dchan := make(chan webm.Packet, chancap)
	wchan := make(chan *image.YCbCr, chancap)
	go read(dchan)
	go decode(dchan, wchan)
	write(wchan)
}
