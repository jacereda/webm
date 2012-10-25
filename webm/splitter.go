package webm

import ()

type Splitter struct {
	streams [MAXSTREAMS]*Stream
	ch      <-chan Packet
}

func NewSplitter(ch <-chan Packet) *Splitter {
	var s Splitter
	s.ch = ch
	return &s
}

func (s *Splitter) addStream(stream *Stream) {
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

func (s *Splitter) Split(strms ...*Stream) {
	for _, strm := range strms {
		if strm != nil {
			s.addStream(strm)
		}
	}
	go s.split()
}
