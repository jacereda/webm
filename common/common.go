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
)

func Main(vpresent func(ch <-chan *ffvp8.Frame),
	apresent func(ch <-chan *ffvorbis.Samples, atrack *webm.Audio)) {

	var err error
	var wm webm.WebM
	r, err := os.Open(*In)
	defer r.Close()
	if err != nil {
		log.Panic("unable to open file " + *In)
	}
	br := bufio.NewReader(r)
	pchan := webm.Parse(br, &wm)

	splitter := webm.NewSplitter(pchan)

	var vtrack *webm.TrackEntry
	var vstream *webm.Stream
	if vpresent != nil {
		vtrack = wm.FindFirstVideoTrack()
	}
	if vtrack != nil {
		vstream = webm.NewStream(vtrack)
	}
	if vstream != nil {
		splitter.AddStream(vstream)
	}

	var astream *webm.Stream
	var atrack *webm.TrackEntry
	if apresent != nil {
		atrack = wm.FindFirstAudioTrack()
	}
	if atrack != nil {
		astream = webm.NewStream(atrack)
	}
	if astream != nil {
		splitter.AddStream(astream)
	}

	splitter.Split()
	switch {
	case astream != nil && vstream != nil:
		go apresent(astream.AudioChannel(), &atrack.Audio)
		vpresent(vstream.VideoChannel())
	case vstream != nil:
		vpresent(vstream.VideoChannel())
	case astream != nil:
		apresent(astream.AudioChannel(), &atrack.Audio)
	}
}
