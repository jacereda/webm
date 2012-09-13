package webm

import (
	"bufio"
	"os"
	"testing"
)

func TestReadStruct(t *testing.T) {
	path := "/Users/jacereda/Downloads/big-buck-bunny_trailer.webm"
	r, err := os.Open(path)
	if err != nil {
		t.Fatal("unable to open file " + path)
	}
	var w WebM
	br := bufio.NewReader(r)
	var id uint
	id, err = Parse(br, &w)
	t.Logf("%+v\n%v\n", w, err)
	for false && err == nil {
		var d []byte
		d,err = Next(br, id)
		t.Log("SLICE: ", d)
	}
}
