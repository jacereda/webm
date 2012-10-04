package main

import (
	"bufio"
	"code.google.com/p/ebml-go/common"
	"code.google.com/p/ebml-go/webm"
	"flag"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"log"
	"os"
	"runtime"
)

type encjob struct {
	path string
	img  image.Image
}

var (
	out      = flag.String("o", "", "Output prefix")
	format   = flag.String("f", "png", "Output format")
	encoders = flag.Int("j", 4, "Number of parallel encoders")
	chans    []chan encjob
	endchans []chan int
)

func jpegenc(w io.Writer, img image.Image) error {
	return jpeg.Encode(w, img, nil)
}

func pngenc(w io.Writer, img image.Image) error {
	return png.Encode(w, img)
}

func encoder(ch <-chan encjob, ech chan<- int) {
	var enc func(io.Writer, image.Image) error
	switch *format {
	case "jpeg", "jpg":
		enc = jpegenc
	case "png":
		enc = pngenc
	default:
		log.Panic("Unsupported output format")
	}
	for job := range ch {
		f, err := os.Create(job.path)
		bf := bufio.NewWriter(f)
		if err != nil {
			log.Panic("unable to open file " + *out)
		}
		enc(bf, job.img)
		f.Close()
	}
	close(ech)
}

func write(ch <-chan webm.Frame) {
	i := 0
	for frame := range ch {
		if *out != "" {
			job := encjob{fmt.Sprint(*out, i, ".", *format), frame}
			chans[i%*encoders] <- job
			runtime.GC()
		}
		i++
	}
	for _, ch := range chans {
		close(ch)
	}
	for _, ech := range endchans {
		<-ech
	}
}

func main() {
	flag.Parse()
	for i := 0; i < *encoders; i++ {
		ch := make(chan encjob, 0)
		chans = append(chans, ch)
		ech := make(chan int, 0)
		endchans = append(endchans, ech)
		go encoder(ch, ech)
	}
	common.Main(write, nil)

}
