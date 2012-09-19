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
	e, rest, err := Parse(br, &w)
	t.Log("Duration: ", w.Segment.GetDuration())
	t.Logf("%+v\n%v %v\n", w, err, e)
	for err == nil {
		t.Log("Packet: ", e.Id, e.Size())
		_, err = e.ReadData()
		e, err = rest.Next()
	}
}
