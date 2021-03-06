package webm

import (
	"github.com/jacereda/ffvorbis"
	"time"
)

type Samples struct {
	Data     []float32
	Timecode time.Duration
	Rebase   bool
	EOS      bool
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
			int(track.Channels),
			int(track.SamplingFrequency)),
		duration: int(time.Duration(time.Second) /
			time.Duration(track.Audio.SamplingFrequency)),
		chans: int(track.Channels),
	}
}

func (d *AudioDecoder) estimate() time.Duration {
	return d.goodtc + time.Duration(d.duration*d.emitted)
}

func (d *AudioDecoder) Decode(pkt *Packet) bool {
	sent := false
	var data []float32
	if pkt.Data == nil {
		eos := Samples{nil, BadTC, false, true}
		d.Chan <- eos
	} else {
		data = d.dec.Decode(pkt.Data)
	}
	if data != nil {
		smp := Samples{data, pkt.Timecode, pkt.Rebase, false}
		if smp.Timecode == BadTC {
			smp.Timecode = d.estimate()
		} else {
			d.goodtc = smp.Timecode
			d.emitted = 0
		}
		d.emitted += len(smp.Data) / d.chans
		if !pkt.Invisible {
			d.Chan <- smp
			sent = true
		}
	}
	return sent
}

func (d *AudioDecoder) Close() {
	close(d.Chan)
}
