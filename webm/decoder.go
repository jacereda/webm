package webm

type Decoder interface {
	Decode(*Packet)
	Close()
}
