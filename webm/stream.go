package webm

import (
	"code.google.com/p/ffvp8-go/ffvp8"
)

type Stream struct {
	Track   *TrackEntry
	Decoder Decoder
}

func NewStream(track *TrackEntry) *Stream {
	var s Stream
	s.Track = track
	if track.IsAudio() {
		s.Decoder = NewAudioDecoder(track)
	}
	if track.IsVideo() {
		s.Decoder = NewVideoDecoder(track)
	}
	return &s
}

func (s *Stream) VideoChannel() <-chan *ffvp8.Frame {
	return s.Decoder.(*VideoDecoder).Chan
}

func (s *Stream) AudioChannel() <-chan Samples {
	return s.Decoder.(*AudioDecoder).Chan
}
