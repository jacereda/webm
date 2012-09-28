package common

import (
	"bufio"
	"code.google.com/p/ebml-go/webm"
	"code.google.com/p/ffvorbis-go/ffvorbis"
	"code.google.com/p/ffvp8-go/ffvp8"
	"flag"
	"log"
	"os"
)

var (
	In = flag.String("i", "", "Input file")
	nf = flag.Int("n", 0x7fffffff, "Number of frames")
)

func Main(vpresent func(ch <-chan *ffvp8.Frame),
	apresent func(ch <-chan *ffvorbis.Samples)) {

	var err error
	var wm webm.WebM
	r, err := os.Open(*In)
	defer r.Close()
	if err != nil {
		log.Panic("unable to open file " + *In)
	}
	br := bufio.NewReader(r)
	pchan := webm.Parse(br, &wm)

	vtrack := wm.FindFirstVideoTrack()
	if vpresent == nil {
		vtrack = nil
	}
	vstream := webm.NewStream(vtrack)

	atrack := wm.FindFirstAudioTrack()
	if apresent == nil {
		atrack = nil
	}
	astream := webm.NewStream(atrack)
	webm.Split(pchan, []*webm.Stream{vstream, astream})
	vchan := webm.DecodeVideo(vstream)
	achan := webm.DecodeAudio(astream)
	if apresent != nil && vpresent != nil {
		go apresent(achan)
	}
	if vpresent != nil {
		vpresent(vchan)
	} else {
		apresent(achan)
	}
}
