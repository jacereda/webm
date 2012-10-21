package webm

import (
	"github.com/petar/GoLLRB/llrb"
	"time"
)

type seekEntry struct {
	t      time.Duration
	offset int64
}

type seekIndex struct {
	t llrb.Tree
}

func newSeekIndex() *seekIndex {
	return &seekIndex{*llrb.New(func(a, b interface{}) bool {
		return a.(seekEntry).t > b.(seekEntry).t
	})}
}

func (si *seekIndex) append(se seekEntry) {
	si.t.ReplaceOrInsert(se)
}

func (si *seekIndex) search(t time.Duration) (val seekEntry) {
	si.t.AscendGreaterOrEqual(seekEntry{t, 0}, func(i llrb.Item) bool {
		val = i.(seekEntry)
		return false
	})
	return
}
