package webm

import (
	"code.google.com/p/ffvorbis-go/ffvorbis"
	"log"
	"time"
)

type AudioDecoder struct {
	Chan     chan *ffvorbis.Samples
	dec      *ffvorbis.Decoder
	lasttc   time.Duration
	duration time.Duration
	emitted  int
}

func NewAudioDecoder(track *TrackEntry) *AudioDecoder {
	var d AudioDecoder
	d.Chan = make(chan *ffvorbis.Samples, 4)
	d.dec = ffvorbis.NewDecoder(track.CodecPrivate)
	d.duration = track.samplesDuration(1)
	return &d
}

func (d *AudioDecoder) Decode(pkt *Packet) {
	smp := d.dec.Decode(pkt.Data, pkt.Timecode)
	if smp != nil {
		if smp.Timecode == BadTC {
			smp.Timecode = d.lasttc + time.Duration(d.emitted)*d.duration
			d.emitted += len(smp.Data) / 4
		} else {
			d.lasttc = smp.Timecode
			d.emitted = 0
		}
		if !pkt.Invisible {
			d.Chan <- smp
		} else {
			log.Println("Invisible audio packet")
		}
	}
}

func (d *AudioDecoder) Close() {
	close(d.Chan)
}
