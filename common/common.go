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

	var streams []*webm.Stream

	var vtrack *webm.TrackEntry
	var vstream *webm.VideoStream
	if vpresent != nil {
		vtrack = wm.FindFirstVideoTrack()
	}
	if vtrack != nil {
		vstream = webm.NewVideoStream(vtrack)
	}
	if vstream != nil {
		streams = append(streams, &vstream.Stream)
	}

	var astream *webm.AudioStream
	var atrack *webm.TrackEntry
	if apresent != nil {
		atrack = wm.FindFirstAudioTrack()
	}
	if atrack != nil {
		astream = webm.NewAudioStream(atrack)
	}
	if astream != nil {
		streams = append(streams, &astream.Stream)
	}

	webm.Split(pchan, streams)
	switch {
	case astream != nil && vstream != nil:
		go apresent(astream.Decode(), &atrack.Audio)
		vpresent(vstream.Decode())
	case vstream != nil:
		vpresent(vstream.Decode())
	case astream != nil:
		apresent(astream.Decode(), &atrack.Audio)
	}
}
