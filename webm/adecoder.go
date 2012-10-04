package webm

import (
	"code.google.com/p/ffvorbis-go/ffvorbis"
	"log"
	"time"
)

type Samples struct {
	Data     []float32
	Timecode time.Duration
}

type AudioDecoder struct {
	Chan     chan Samples
	dec      *ffvorbis.Decoder
	lasttc   time.Duration
	duration time.Duration
	emitted  int
}

func NewAudioDecoder(track *TrackEntry) *AudioDecoder {
	return &AudioDecoder{
		Chan: make(chan Samples, 4),
		dec: ffvorbis.NewDecoder(track.CodecPrivate,
			int(track.Audio.SamplingFrequency),
			int(track.Audio.Channels)),
		duration: track.samplesDuration(1),
	}
}

func (d *AudioDecoder) Decode(pkt *Packet) {
	data := d.dec.Decode(pkt.Data)
	if data != nil {
		smp := Samples{data, pkt.Timecode}
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
