package common

import (
	"bufio"
	"code.google.com/p/ebml-go/webm"
	"code.google.com/p/ffvp8-go/ffvp8"
	"flag"
	"log"
	"os"
)

var (
	in = flag.String("i", "", "Input file")
	nf = flag.Int("n", 0x7fffffff, "Number of frames")
)

const chancap = 0

func decode(dchan <-chan webm.Packet, wchan chan<- *ffvp8.Frame) {
	dec := ffvp8.NewDecoder()
	for pkt := range dchan {
		img := dec.Decode(pkt.Data, pkt.Timecode)
		if !pkt.Invisible {
			wchan <- img
		}
	}
	close(wchan)
}

func read(dchan chan<- webm.Packet) {
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
	if track == nil {
		log.Panic("No video track")
	}
	i := 0
	for pkt := range pchan {
		if i >= *nf {
			break
		}
		if pkt.TrackNumber == track.TrackNumber {
			dchan <- pkt
			i++
		}
	}
	close(dchan)
}

func Main(write func(ch <-chan *ffvp8.Frame)) {
	flag.Parse()
	dchan := make(chan webm.Packet, chancap)
	wchan := make(chan *ffvp8.Frame, chancap)
	go read(dchan)
	go decode(dchan, wchan)
	write(wchan)
}
