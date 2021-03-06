package webm

import (
	"time"
)

type Splitter struct {
	streams   [MAXSTREAMS]*Stream
	expecting [MAXSTREAMS]time.Duration
	ch        <-chan Packet
}

func (s *Splitter) expect(t time.Duration) {
	for i := 0; i < MAXSTREAMS; i++ {
		s.expecting[i] = t - time.Millisecond
	}
}

func NewSplitter(ch <-chan Packet) *Splitter {
	var s Splitter
	s.ch = ch
	return &s
}

func (s *Splitter) addStream(stream *Stream) {
	s.streams[stream.Track.TrackNumber] = stream
}

func tabs(t time.Duration) time.Duration {
	if t < 0 {
		return -t
	}
	return t
}

func (s *Splitter) split() {
	for pkt := range s.ch {
		if pkt.Data == nil {
			if pkt.Timecode == BadTC {
				for _, strm := range s.streams {
					if strm != nil {
						strm.Decoder.Decode(&pkt)
					}
				}
			} else {
				s.expect(pkt.Timecode)
			}
		} else {
			strm := s.streams[pkt.TrackNumber]
			expecting := s.expecting[pkt.TrackNumber]
			if expecting != BadTC {
				if pkt.Timecode < expecting {
					pkt.Invisible = true
				} else {
					pkt.Rebase = true
				}
			}
			if strm != nil {
				sent := strm.Decoder.Decode(&pkt)
				if expecting != BadTC && sent {
					s.expecting[pkt.TrackNumber] = BadTC
				}
			}
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
