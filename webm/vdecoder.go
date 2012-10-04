package webm

import (
	"code.google.com/p/ffvp8-go/ffvp8"
	"log"
	"time"
)

type VideoDecoder struct {
	Chan     chan *ffvp8.Frame
	dec      *ffvp8.Decoder
	duration time.Duration
	emitted  uint
	lasttc   time.Duration
}

func NewVideoDecoder(track *TrackEntry) *VideoDecoder {
	var d VideoDecoder
	d.Chan = make(chan *ffvp8.Frame, 4)
	d.dec = ffvp8.NewDecoder()
	d.duration = time.Duration(track.DefaultDuration)
	return &d
}

func (d *VideoDecoder) Decode(pkt *Packet) {
	img := d.dec.Decode(pkt.Data, pkt.Timecode)
	if img != nil {
		if img.Timecode == BadTC {
			d.emitted++
			img.Timecode = d.lasttc + time.Duration(d.emitted)*d.duration
		} else {
			d.lasttc = img.Timecode
			d.emitted = 0
		}
		if !pkt.Invisible {
			d.Chan <- img
		} else {
			log.Println("Invisible video packet")
		}
	}
}

func (d *VideoDecoder) Close() {
	close(d.Chan)
}
