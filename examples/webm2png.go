package main

import (
	"code.google.com/p/ebml-go/common"
	"flag"
	"fmt"
	"image"
	"image/png"
	"log"
	"os"
	"runtime"
)

var out = flag.String("o", "", "Output prefix")

func write(ch <-chan *image.YCbCr) {
	for i, img := 0, <-ch; img != nil; i, img = i+1, <-ch {
		if *out != "" {
			path := fmt.Sprint(*out, i, ".png")
			f, err := os.Create(path)
			if err != nil {
				log.Panic("unable to open file " + *out)
			}
			png.Encode(f, img)
			f.Close()
			runtime.GC()
		}
	}
}

func main() {
	common.Main(write)
}
