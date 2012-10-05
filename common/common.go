package common

import (
	"bufio"
	"code.google.com/p/ebml-go/webm"
	"flag"
	"log"
	"os"
)

var (
	In = flag.String("i", "", "Input file")
)

func Main(vpresent func(<-chan webm.Frame, *webm.Reader),
	apresent func(<-chan webm.Samples, *webm.Audio)) {
	var err error
	var wm webm.WebM
	r, err := os.Open(*In)
	defer r.Close()
	if err != nil {
		log.Panic("Unable to open file " + *In)
	}
	br := bufio.NewReader(r)
	reader, err := webm.Parse(br, &wm)
	reader.Resume()
	if err != nil {
		log.Panic("Unable to parse file:", err)
	}
	splitter := webm.NewSplitter(reader.Chan)

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
		vpresent(vstream.VideoChannel(), reader)
	case vstream != nil:
		vpresent(vstream.VideoChannel(), reader)
	case astream != nil:
		apresent(astream.AudioChannel(), &atrack.Audio)
	}
}
