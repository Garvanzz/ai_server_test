package codec

import (
	"errors"
	"fmt"
)

var (
	ErrWrongCheckFbFlag = errors.New("codec: packet fb error")
	ErrPacketSizeExceed = errors.New("codec: packet size exceed")
	ErrPacketWrongType  = errors.New("codec: invalid packet type")
)

type Packet struct {
	Type   uint32
	Length int
	Data   []byte
}

// New create a Packet instance.
func New() *Packet {
	return &Packet{}
}

// String represents the Packet's in text mode.
func (p *Packet) String() string {
	return fmt.Sprintf("Type: %d, Length: %d, Data: %s", p.Type, p.Length, string(p.Data))
}
