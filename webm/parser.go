// Copyright 2011 The ebml-go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package webm

import (
	"code.google.com/p/ebml-go/ebml"
	"io"
)

type WebM struct {
	Header  `ebml:"1a45dfa3"`
	Segment `ebml:"18538067"`
}

type Header struct {
	EBMLVersion        uint   `ebml:"4286" ebmldef:"1"`
	EBMLReadVersion    uint   `ebml:"42f7" ebmldef:"1"`
	EBMLMaxIDLength    uint   `ebml:"42f2" ebmldef:"4"`
	EBMLMaxSizeLength  uint   `ebml:"42f3" ebmldef:"8"`
	DocType            string `ebml:"4282"`
	DocTypeVersion     uint   `ebml:"4287" ebmldef:"1"`
	DocTypeReadVersion uint   `ebml:"4285" ebmldef:"1"`
	Clavijo            uint   `ebml:"ABCD" ebmldef:"123"`
}

type Segment struct {
	SeekHead           `ebml:"114D9B74"`
	SegmentInformation `ebml:"1549A966"`
	Tracks             `ebml:"1654AE6B"`
	Cluster            `ebml:"1F43B675"`
	Cues               `ebml:"1C53BB6B"`
}

type Tracks struct {
	TrackEntry []TrackEntry `ebml:"AE"`
}

type TrackEntry struct {
	TrackNumber     uint   `ebml:"D7"`
	TrackUID        uint64 `ebml:"73C5"`
	TrackType       uint   `ebml:"83"`
	FlagEnabled     uint   `ebml:"B9" ebmldef:"1"`
	FlagDefault     uint   `ebml:"88" ebmldef:"1"`
	FlagForced      uint   `ebml:"55AA" ebmldef:"0"`
	FlagLacing      uint   `ebml:"9C" ebmldef:"1"`
	DefaultDuration uint64 `ebml:"23E383"`
	Name            string `ebml:"536E"`
	Language        string `ebml:"22B59C" ebmldef:"eng"`
	CodecID         string `ebml:"86"`
	CodecPrivate    []byte `ebml:"63A2"`
	CodecName       string `ebml:"258688"`
	Video           `ebml:"E0"`
	Audio           `ebml:"E1"`
}

type Video struct {
	FlagInterlaced  uint `ebml:"9A" ebmldef:"0"`
	StereoMode      uint `ebml:"53B8" ebmldef:"0"`
	PixelWidth      uint `ebml:"B0"`
	PixelHeight     uint `ebml:"BA"`
	PixelCropBottom uint `ebml:"54AA" ebmldef:"0"`
	PixelCropTop    uint `ebml:"54BB" ebmldef:"0"`
	PixelCropLeft   uint `ebml:"54CC" ebmldef:"0"`
	PixelCropRight  uint `ebml:"54DD" ebmldef:"0"`
	DisplayWidth    uint `ebml:"54B0" ebmldeflink:"PixelWidth"`
	DisplayHeight   uint `ebml:"54BA" ebmldeflink:"PixelHeight"`
	DisplayUnit     uint `ebml:"54B2" ebmldef:"0"`
	AspectRatioType uint `ebml:"54B3" ebmldef:"0"`
}

type Audio struct {
	SamplingFrequency       float64 `ebml:"B5" ebmldef:"8000.0"`
	OutputSamplingFrequency float64 `ebml:"78B5" ebmldeflink:"SamplingFrequency"`
	Channels                uint    `ebml:"9F" ebmldef:"1"`
	BitDepth                uint    `ebml:"6264"`
}

type SeekHead struct {
	Seek []Seek `ebml:"4DBB"`
}

type Seek struct {
	SeekID       []byte `ebml:"53AB"`
	SeekPosition uint64 `ebml:"53AC"`
}

type SegmentInformation struct {
	TimecodeScale uint    `ebml:"2AD7B1" ebmldef:"1000000"`
	Duration      float64 `ebml:"4489"`
	DateUTC       []byte  `ebml:"4461"`
	MuxingApp     string  `ebml:"4D80"`
	WritingApp    string  `ebml:"5741"`
}

func (s *SegmentInformation) GetDuration() float64 {
	return s.Duration * float64(s.TimecodeScale) / 1000000000
}

type Cluster struct {
	simpleBlock []byte       `ebml:"A3" ebmlstop:"1"`
	Timecode    uint         `ebml:"E7"`
	PrevSize    uint         `ebml:"AB"`
	Position    uint         `ebml:"A7"`
	BlockGroup  []BlockGroup `ebml:"A0"`
}

type BlockGroup struct {
	block          []byte   `ebml:"A1" ebmlstop:"1"`
	BlockDuration  uint     `ebml:"9B"`
	ReferenceBlock int      `ebml:"FB"`
	CodecState     []byte   `ebml:"A4"`
	Slices         []Slices `ebml:"8E"`
}

type Slices struct {
	TimeSlice []TimeSlice `ebml:"E8"`
}

type TimeSlice struct {
	LaceNumber uint `ebml:"CC" ebmldef:"0"`
}

type Cues struct {
	CuePoint []CuePoint `ebml:"BB"`
}

type CuePoint struct {
	CueTime           uint                `ebml:"B3"`
	CueTrackPositions []CueTrackPositions `ebml:"B7"`
}

type CueTrackPositions struct {
	CueTrack           uint `ebml:"F7"`
	CueClusterPosition uint `ebml:"F1"`
	CueBlockNumber     uint `ebml:"5378" ebmldef:"1"`
}

func Parse(r io.Reader, m *WebM) (first *ebml.Element, rest *ebml.Element, err error) {
	var e *ebml.Element
	e, err = ebml.RootElement(r)
	if err == nil {
		err = e.Unmarshal(m)
	}
	if err.Error() == "Reached payload" {
		first = err.(ebml.ReachedPayloadError).First
		rest = err.(ebml.ReachedPayloadError).Rest
		err = nil
	}
	return
}
