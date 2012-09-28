package webm

import (
	"code.google.com/p/ffvp8-go/ffvp8"
	"log"
)

type VideoStream struct {
	Stream
}

func NewVideoStream(track *TrackEntry) *VideoStream {
	var s VideoStream
	s.Stream.init(track)
	return &s
}

func vdecode(dec *ffvp8.Decoder, in <-chan Packet, out chan<- *ffvp8.Frame) {
	for pkt := range in {
		img := dec.Decode(pkt.Data, pkt.Timecode)
		if !pkt.Invisible {
			out <- img
		} else {
			log.Println("Invisible video packet")
		}
	}
	close(out)
}

func (s *VideoStream) Decode() <-chan *ffvp8.Frame {
	dec := ffvp8.NewDecoder()
	out := make(chan *ffvp8.Frame, 4)
	go vdecode(dec, s.Chan, out)
	return out
}
