package webm

import (
	"code.google.com/p/ebml-go/ebml"
	"log"
	"time"
)

const BadTC = time.Duration(-1000000000000000)

type Packet struct {
	Data        []byte
	Timecode    time.Duration
	TrackNumber uint
	Invisible   bool
	Keyframe    bool
	Discardable bool
}

const (
	cmdNone   = 0
	cmdPause  = 1
	cmdResume = 2
	cmdStep   = 3
)

type Reader struct {
	Chan  chan Packet
	seek  chan time.Duration
	steps uint
}

func (r *Reader) send(p *Packet) {
	r.handleCommands()
	r.Chan <- *p
}

func remaining(x int8) (rem int) {
	for x > 0 {
		rem++
		x += x
	}
	return
}

func laceSize(v []byte) (val int, rem int) {
	val = int(v[0])
	rem = remaining(int8(val))
	for i, l := 1, rem+1; i < l; i++ {
		val <<= 8
		val += int(v[i])
	}
	val &= ^(128 << uint(rem*8-rem))
	return
}

func laceDelta(v []byte) (val int, rem int) {
	val, rem = laceSize(v)
	val -= (1 << (uint(7*(rem+1) - 1))) - 1
	return
}

func (r *Reader) sendLaces(p *Packet, d []byte, sz []int) {
	var curr int
	for i, l := 0, len(sz); i < l; i++ {
		if sz[i] != 0 {
			p.Data = d[curr : curr+sz[i]]
			r.send(p)
			curr += sz[i]
			p.Timecode = BadTC
		}
	}
	p.Data = d[curr:]
	r.send(p)
}

func parseXiphSizes(d []byte) (sz []int, curr int) {
	laces := int(uint(d[4]))
	sz = make([]int, laces)
	curr = 5
	for i := 0; i < laces; i++ {
		for d[curr] == 255 {
			sz[i] += 255
			curr++
		}
		sz[i] += int(uint(d[curr]))
		curr++
	}
	return
}

func parseFixedSizes(d []byte) (sz []int, curr int) {
	laces := int(uint(d[4]))
	curr = 5
	fsz := len(d[curr:]) / (laces + 1)
	sz = make([]int, laces)
	for i := 0; i < laces; i++ {
		sz[i] = fsz
	}
	return
}

func parseEBMLSizes(d []byte) (sz []int, curr int) {
	laces := int(uint(d[4]))
	sz = make([]int, laces)
	curr = 5
	var rem int
	sz[0], rem = laceSize(d[curr:])
	for i := 1; i < laces; i++ {
		curr += rem + 1
		var dsz int
		dsz, rem = laceDelta(d[curr:])
		sz[i] = sz[i-1] + dsz
	}
	curr += rem + 1
	return
}

func (r *Reader) sendBlock(hdr []byte, tbase time.Duration) {
	var p Packet
	p.TrackNumber = uint(hdr[0]) & 0x7f
	p.Timecode = tbase + time.Millisecond*time.Duration(
		uint(hdr[1])<<8+uint(hdr[2]))
	p.Invisible = (hdr[3] & 8) != 0
	p.Keyframe = (hdr[3] & 0x80) != 0
	p.Discardable = (hdr[3] & 1) != 0
	if p.Discardable {
		log.Println("Discardable packet")
	}
	lacing := (hdr[3] >> 1) & 3
	switch lacing {
	case 0:
		p.Data = hdr[4:]
		r.send(&p)
	case 1:
		sz, curr := parseXiphSizes(hdr)
		r.sendLaces(&p, hdr[curr:], sz)
	case 2:
		sz, curr := parseFixedSizes(hdr)
		r.sendLaces(&p, hdr[curr:], sz)
	case 3:
		sz, curr := parseEBMLSizes(hdr)
		r.sendLaces(&p, hdr[curr:], sz)
	}
}

func (r *Reader) sendPackets(e *ebml.Element, rest *ebml.Element, tbase time.Duration) {
	var err error
	curr := 0
	for err == nil {
		var hdr []byte
		hdr, err = e.ReadData()
		if err != nil {
			log.Println(err)
		}
		if err == nil && e.Id == 163 && len(hdr) > 4 {
			r.sendBlock(hdr, tbase)
		} else {
			log.Println("Unexpected packet")
		}
		e, err = rest.Next()
		curr++
	}
}

func (r *Reader) parseClusters(e *ebml.Element, rest *ebml.Element) {
	var err error
	for err == nil && e != nil {
		var c Cluster
		err = e.Unmarshal(&c)
		if err != nil && err.Error() == "Reached payload" {
			r.sendPackets(err.(ebml.ReachedPayloadError).First,
				err.(ebml.ReachedPayloadError).Rest,
				time.Millisecond*time.Duration(c.Timecode))
		}
		e, err = rest.Next()
	}
	close(r.Chan)
}

func newReader(e, rest *ebml.Element) *Reader {
	r := &Reader{make(chan Packet, 2),
		make(chan int, 0),
		0}
	go r.parseClusters(e, rest)
	return r
}

func (r *Reader) Seek(t time.Duration) {
	r.seek <- t
}
