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
	goodtc   time.Duration
	duration int
	emitted  int
	chans    int
}

func NewAudioDecoder(track *TrackEntry) *AudioDecoder {
	return &AudioDecoder{
		Chan: make(chan Samples, 4),
		dec: ffvorbis.NewDecoder(track.CodecPrivate,
			int(track.SamplingFrequency),
			int(track.Channels)),
		duration: int(time.Duration(time.Second) /
			time.Duration(track.Audio.SamplingFrequency)),
		chans: int(track.Channels),
	}
}

func (d *AudioDecoder) estimate() time.Duration {
	return d.goodtc + time.Duration(d.duration*d.emitted)
}

func (d *AudioDecoder) Decode(pkt *Packet) {
	data := d.dec.Decode(pkt.Data)
	if data != nil {
		smp := Samples{data, pkt.Timecode}
		if smp.Timecode == BadTC {
			smp.Timecode = d.estimate()
		} else {
			//			log.Println("good tc:", smp.Timecode - d.estimate(), d.duration)
			d.goodtc = smp.Timecode
			d.emitted = 0
		}
		d.emitted += len(smp.Data) / d.chans
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
