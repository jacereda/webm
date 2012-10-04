package webm

import (
	"code.google.com/p/ffvp8-go/ffvp8"
	"image"
	"log"
	"time"
)

type Frame struct {
	*image.YCbCr
	Timecode time.Duration
}

type VideoDecoder struct {
	Chan     chan Frame
	dec      *ffvp8.Decoder
	duration time.Duration
	emitted  uint
	lasttc   time.Duration
}

func NewVideoDecoder(track *TrackEntry) *VideoDecoder {
	var d VideoDecoder
	d.Chan = make(chan Frame, 4)
	d.dec = ffvp8.NewDecoder()
	d.duration = time.Duration(track.DefaultDuration)
	return &d
}

func (d *VideoDecoder) Decode(pkt *Packet) {
	img := d.dec.Decode(pkt.Data)
	if img != nil {
		frame := Frame{img, pkt.Timecode}
		if frame.Timecode == BadTC {
			d.emitted++
			frame.Timecode = d.lasttc + time.Duration(d.emitted)*d.duration
		} else {
			d.lasttc = frame.Timecode
			d.emitted = 0
		}
		if !pkt.Invisible {
			d.Chan <- frame
		} else {
			log.Println("Invisible video packet")
		}
	}
}

func (d *VideoDecoder) Close() {
	close(d.Chan)
}
