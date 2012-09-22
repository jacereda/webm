package main

import (
	"code.google.com/p/ebml-go/common"
	"code.google.com/p/ffvp8-go/ffvp8"
	"flag"
	"fmt"
	"image/png"
	"log"
	"os"
	"runtime"
)

var out = flag.String("o", "", "Output prefix")

func write(ch <-chan *ffvp8.Frame) {
	i := 0
	for img := range ch {
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
		i++
	}
}

func main() {
	common.Main(write)
}
