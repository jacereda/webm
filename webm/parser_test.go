package webm

import (
	"bufio"
	"code.google.com/p/ebml-go/ebml"
	"testing"
	"os"
)

func TestReadStruct(t *testing.T) {
	path := "/Users/jacereda/Downloads/big-buck-bunny_trailer.webm"
	r,err := os.Open(path)
	if err != nil {
		t.Fatal("unable to open file " + path)
	}
	br := bufio.NewReader(r)
	var w WebM
	err = ebml.Read(br, &w)
	t.Logf("%v\n%+v\n", err, w)
}
