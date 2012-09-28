package webm

import (
	"code.google.com/p/ffvorbis-go/ffvorbis"
	"log"
)

func adecode(dec *ffvorbis.Decoder,
	in <-chan Packet, out chan<- *ffvorbis.Samples) {
	for pkt := range in {
		buf := dec.Decode(pkt.Data, pkt.Timecode)
		if buf != nil {
			if !pkt.Invisible {
				out <- buf
			} else {
				log.Println("Invisible audio packet")
			}
		}
	}
	close(out)
}

func DecodeAudio(s *Stream) <-chan *ffvorbis.Samples {
	if s == nil {
		return nil
	}
	out := make(chan *ffvorbis.Samples, 4)
	dec := ffvorbis.NewDecoder(s.Track.CodecPrivate)
	go adecode(dec, s.Chan, out)
	return out
}
