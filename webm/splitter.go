package webm

import ()

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
