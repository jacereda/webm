package webm

type Stream struct {
	Chan  chan Packet
	Track *TrackEntry
}

func NewStream(track *TrackEntry) *Stream {
	if track == nil {
		return nil
	}
	return &Stream{make(chan Packet, 4), track}
}

func split(pchan <-chan Packet, streams []*Stream) {
	for pkt := range pchan {
		for _, s := range streams {
			if pkt.TrackNumber == s.Track.TrackNumber {
				s.Chan <- pkt
			}
		}
	}
	for _, s := range streams {
		close(s.Chan)
	}
}

func Split(pchan <-chan Packet, streams []*Stream) {
	var fstreams []*Stream
	// XXX Isn't there a filter()?
	for _, s := range streams {
		if s != nil {
			fstreams = append(fstreams, s)
		}
	}
	go split(pchan, fstreams)
}
