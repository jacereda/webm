package webm

import (
	"code.google.com/p/ffvp8-go/ffvp8"
	"image"
	"time"
)

type Frame struct {
	*image.YCbCr
	Timecode time.Duration
	Rebase   bool
	EOS      bool
}

type VideoDecoder struct {
	Chan     chan Frame
	dec      *ffvp8.Decoder
	duration time.Duration
	emitted  uint
	goodtc   time.Duration
	rebase   bool
	flushed  bool
}

func NewVideoDecoder(track *TrackEntry) *VideoDecoder {
	return &VideoDecoder{
		Chan:     make(chan Frame, 4),
		dec:      ffvp8.NewDecoder(),
		duration: time.Duration(track.DefaultDuration),
	}
}

func (d *VideoDecoder) estimate() time.Duration {
	return d.goodtc + time.Duration(d.emitted)*d.duration
}

func (d *VideoDecoder) Decode(pkt *Packet) bool {
	sent := false
	if false {
		if pkt.Rebase {
			d.rebase = true
			d.flushed = false
		}
		if d.rebase && !d.flushed {
			if pkt.Keyframe {
				d.dec.Flush()
				pkt.Rebase = true
				d.flushed = true
			} else {
				return false
			}
		}
	}
	var img *image.YCbCr
	if pkt.Data == nil {
		eos := Frame{nil, BadTC, false, true}
		//		img = d.dec.Decode(nil)
		d.Chan <- eos
	} else {
		img = d.dec.Decode(pkt.Data)
	}
	if img != nil {
		frame := Frame{img, pkt.Timecode, pkt.Rebase, false}
		d.rebase = false
		if frame.Timecode == BadTC {
			frame.Timecode = d.estimate()
		} else {
			d.goodtc = frame.Timecode
			d.emitted = 0
		}
		d.emitted++
		if !pkt.Invisible {
			d.Chan <- frame
			sent = true
		}
	}
	return sent
}

func (d *VideoDecoder) Close() {
	close(d.Chan)
}
