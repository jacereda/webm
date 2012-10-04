package webm

import ()

type Splitter struct {
	streams [16]*Stream
	ch      <-chan Packet
}

func NewSplitter(ch <-chan Packet) *Splitter {
	var s Splitter
	s.ch = ch
	return &s
}

func (s *Splitter) AddStream(stream *Stream) {
	s.streams[stream.Track.TrackNumber] = stream
}

func (s *Splitter) split() {
	for pkt := range s.ch {
		strm := s.streams[pkt.TrackNumber]
		if strm != nil {
			strm.Decoder.Decode(&pkt)
		}
	}
	for _, strm := range s.streams {
		if strm != nil {
			strm.Decoder.Close()
		}
	}
}

func (s *Splitter) Split() {
	go s.split()
}
