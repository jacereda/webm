// Copyright 2011 The ebml-go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.


package webm

type WebM struct {
	Header `id:"1a45dfa3"`
	Segment `id:"18538067"`
}

type Header struct {
	EBMLVersion uint `id:"4286"`
	EBMLReadVersion uint `id:"42f7"`
	EBMLMaxIDLength uint `id:"42f2"`
	EBMLMaxSizeLength uint `id:"42f3"`
	DocType string `id:"4282"`
	DocTypeVersion uint `id:"4287"`
	DocTypeReadVersion uint `id:"4285"`
}

type Segment struct {
	SeekHead `id:"114D9B74"`
	SegmentInformation `id:"1549A966"`
	Tracks `id:"1654AE6B"`
	Cluster `id:"1F43B675"`
	Cues `id:"1C53BB6B"`
}

type Tracks struct {
	TrackEntry []TrackEntry `id:"AE"`
}

type TrackEntry struct {
	TrackNumber uint `id:"D7"`
	TrackUID uint `id:"73C5"`
	TrackType uint `id:"83"`
	FlagEnabled uint  `id:"B9"`
	FlagDefault uint `id:"88"`
	FlagForced uint `id:"55AA"`
	DefaultDuration uint `id:"23E383"`
	Name string `id:"536E"`
	Language string `id:"22B59C"`
	CodecID string `id:"86"`
	CodecPrivate []byte `id:"63A2"`
	CodecName string `id:"258688"`
	Video `id:"E0"`
	Audio `id:"E1"`
}

type Video struct {
	FlagInterlaced uint `id:"9A"`
	StereoMode uint `id:"53B8"`
	PixelWidth uint `id:"B0"`
	PixelHeight uint `id:"BA"`
	PixelCropBottom uint `id:"54AA"`
	PixelCropTop uint `id:"54BB"`
	PixelCropLeft uint `id:"54CC"`
	PixelCropRight uint `id:"54DD"`
	DisplayWidth uint `id:"54B0"`
	DisplayHeight uint `id:"54BA"`
	DisplayUnit uint `id:"54B2"`
	AspectRatioType uint `id:"54B3"`
}

type Audio struct {
	SamplingFrequency float32 `id:"B5"`
	OutputSamplingFrequency float32 `id:"78B5"`
	Channels uint `id:"9F"`
	BitDepth uint `id:"6264"`
}

type SeekHead struct {
	Seek []Seek `id:"4DBB"`
}

type Seek struct {
	SeekID []byte `id:"53AB"`
	SeekPosition uint `id:"53AC"`
}

type SegmentInformation struct {
	TimecodeScale uint `id:"2AD7B1"`
	Duration float32 `id:"4489"`
	DateUTC []byte `id:"4461"`
	MuxingApp string `id:"4D80"`
	WritingApp string `id:"5741"`
	
}

type Cluster struct {
	Timecode uint `id:"E7"`
	PrevSize uint `id:"AB"`
	Position uint `id:"A7"`
	BlockGroup []BlockGroup `id:"A0"`
}

type BlockGroup struct {
	BlockDuration uint `id:"9B"`
	ReferenceBlock int `id:"FB"`
	CodecState []byte `id:"A4"`
	Slices []Slices `id:"8E"`
}

type Slices struct {
	TimeSlice []TimeSlice `id:"E8"`
}

type TimeSlice struct {
	LaceNumber uint `id:"CC"`
}

type Cues struct {
	CuePoint []CuePoint `id:"BB"`
}

type CuePoint struct {
	CueTime uint `id:"B3"`
	CueTrackPositions []CueTrackPositions `id:"B7"`
}

type CueTrackPositions struct {
	CueTrack uint `id:"F7"`
	CueClusterPosition uint `id:"F1"`
	CueBlockNumber uint `id:"5378"`
}

