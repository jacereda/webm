package webm

import (
	"code.google.com/p/ffvorbis-go/ffvorbis"
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

func (s *Stream) VideoChannel() <-chan Frame {
	return s.Decoder.(*VideoDecoder).Chan
}

func (s *Stream) AudioChannel() <-chan *ffvorbis.Samples {
	return s.Decoder.(*AudioDecoder).Chan
}
