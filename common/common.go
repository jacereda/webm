package common

import (
	"bufio"
	"code.google.com/p/ebml-go/webm"
	"code.google.com/p/ffvp8-go/ffvp8"
	"code.google.com/p/ffvorbis-go/ffvorbis"
	"flag"
	"log"
	"os"
)

var (
	In = flag.String("i", "", "Input file")
	nf = flag.Int("n", 0x7fffffff, "Number of frames")
)

const chancap = 4
const achancap = 32

func vdecode(dec *ffvp8.Decoder, 
	vdchan <-chan webm.Packet, vwchan chan<- *ffvp8.Frame) {
	for pkt := range vdchan {
		img := dec.Decode(pkt.Data, pkt.Timecode)
		if !pkt.Invisible {
			vwchan <- img
		} else {
			log.Println("Invisible video packet")
		}
	}
	close(vwchan)
}

func adecode(dec *ffvorbis.Decoder, 
	adchan <-chan webm.Packet, awchan chan<- *ffvorbis.Samples) {
	for pkt := range adchan {
		buf := dec.Decode(pkt.Data, pkt.Timecode)
		if buf != nil {
			if !pkt.Invisible {
				awchan <- buf
			} else {
				log.Println("Invisible audio packet")
			}
		}
	}
	close(awchan)
}

func read(vtrack *webm.TrackEntry, atrack *webm.TrackEntry, 
	pchan <-chan webm.Packet, 
	vdchan chan<- webm.Packet, adchan chan<- webm.Packet) {
	i := 0
	for pkt := range pchan {
		if i >= *nf {
			break
		}
		if vdchan != nil && pkt.TrackNumber == vtrack.TrackNumber {
			vdchan <- pkt
			i++
		}
		if adchan != nil && pkt.TrackNumber == atrack.TrackNumber {
			adchan <- pkt
		}
	}
	if vdchan != nil {
		close(vdchan)
	}
	if adchan != nil {
		close(adchan)
	}
}

func Main(vpresent func(ch <-chan *ffvp8.Frame), 
	apresent func(ch <-chan *ffvorbis.Samples)) {

	var err error
	var wm webm.WebM
	r, err := os.Open(*In)
	defer r.Close()
	if err != nil {
		log.Panic("unable to open file " + *In)
	}
	br := bufio.NewReader(r)
	pchan := webm.Parse(br, &wm)

	vdchan := make(chan webm.Packet, chancap)
	vwchan := make(chan *ffvp8.Frame, chancap)
	adchan := make(chan webm.Packet, achancap)
	awchan := make(chan *ffvorbis.Samples, chancap)

	vtrack := wm.FindFirstVideoTrack()
	if vpresent == nil {
		vtrack = nil
		vdchan = nil
	}
	atrack := wm.FindFirstAudioTrack()
	if apresent == nil {
		atrack = nil
		adchan = nil
	}

	go read(vtrack, atrack, pchan, vdchan, adchan)
	if vtrack != nil {
		go vdecode(ffvp8.NewDecoder(), vdchan, vwchan)
	}
	if atrack != nil {
		go adecode(ffvorbis.NewDecoder(atrack.CodecPrivate), adchan, awchan)
	}
	if apresent != nil && vpresent != nil {
		go apresent(awchan)
	}
	if vpresent != nil {
		vpresent(vwchan)
	} else {
		apresent(awchan)
	}
}
