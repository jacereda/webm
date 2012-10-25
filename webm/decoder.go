package webm

type Decoder interface {
	Decode(*Packet) bool
	Close()
}
