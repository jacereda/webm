package webm

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
