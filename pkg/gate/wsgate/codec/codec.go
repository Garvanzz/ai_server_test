package codec

import (
	"bytes"
	"encoding/binary"
)

const (
	HeadLength    = 4
	ProtoIdLength = 4
	MaxPacketSize = 128 * 1024
)

type Decoder struct {
	buf  *bytes.Buffer
	size int
	typ  uint32
}

func NewDecoder() *Decoder {
	return &Decoder{
		buf:  bytes.NewBuffer(nil),
		size: -1,
	}
}

// Decode decode the network bytes slice to packet.Packet(s)
// TODO(Warning): shared slice
func (c *Decoder) Decode(data []byte) ([]*Packet, error) {
	c.buf.Write(data)

	var (
		packets []*Packet
		err     error
	)

	if c.buf.Len() < HeadLength+ProtoIdLength {
		return nil, err
	}

	if c.size < 0 {
		if err = c.forward(); err != nil {
			return nil, err
		}
	}

	for c.size <= c.buf.Len() {
		p := &Packet{Type: c.typ, Length: c.size, Data: c.buf.Next(c.size)}
		packets = append(packets, p)

		// more packet
		if c.buf.Len() < HeadLength+ProtoIdLength {
			c.size = -1
			break
		}

		if err = c.forward(); err != nil {
			return nil, err
		}
	}
	return packets, err
}

func (c *Decoder) forward() error {
	header := c.buf.Next(HeadLength + ProtoIdLength)

	c.size = int(binary.LittleEndian.Uint32(header[:4]))
	c.typ = binary.LittleEndian.Uint32(header[4:8])
	if c.typ <= 0 {
		return ErrPacketWrongType
	}

	// packet length limitation
	if c.size > MaxPacketSize {
		return ErrPacketSizeExceed
	}
	return nil
}

func Encode(typ uint32, data []byte) ([]byte, error) {
	outBuff := new(bytes.Buffer)
	err := binary.Write(outBuff, binary.LittleEndian, uint32(len(data)))
	if err != nil {
		return nil, err
	}

	err = binary.Write(outBuff, binary.LittleEndian, typ)
	if err != nil {
		return nil, err
	}

	err = binary.Write(outBuff, binary.LittleEndian, data)
	if err != nil {
		return nil, err
	}

	return outBuff.Bytes(), err
}
