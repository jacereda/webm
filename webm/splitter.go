package webm

import ()

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

func split(pchan <-chan Packet, streams []*Stream) {
	for pkt := range pchan {
		for _, s := range streams {
			if pkt.TrackNumber == s.Track.TrackNumber {
				s.Decoder.Decode(&pkt)
			}
		}
	}
	for _, s := range streams {
		s.Decoder.Close()
	}
}

func Split(pchan <-chan Packet, streams []*Stream) {
	go split(pchan, streams)
}
