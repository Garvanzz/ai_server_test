package pb_packet

import (
	"encoding/binary"
	"errors"
	l4g "github.com/alecthomas/log4go"
	"github.com/golang/protobuf/proto"
	"io"
)

const (
	DataLen      = 4
	MessageIDLen = 4
	MaxPacketLen = 128 * 1024
)

/*

s->c

|--totalDataLen(uint32)--|--msgIDLen(uint32)--|--------------data--------------|
|-------------4----------|---------4---------|---------(totalDataLen-2-1)-----|

*/

type IPacket interface {
	Serialize() []byte
}

type Protocol interface {
	ReadPacket(conn io.Reader) (IPacket, error)
	ReadPacketByByte(conn []byte) (IPacket, error)
}

// Packet 服务端发往客户端的消息
type Packet struct {
	id   uint32
	data []byte
}

func (p *Packet) GetMessageID() uint32 {
	return p.id
}

func (p *Packet) GetData() []byte {
	return p.data
}

func (p *Packet) Serialize() []byte {
	dataLen := len(p.data)
	buff := make([]byte, DataLen+MessageIDLen+dataLen)
	buf := make([]byte, DataLen+MessageIDLen)
	buf[0] = byte(dataLen & 0xFF)
	buf[1] = byte(dataLen >> 8 & 0xFF)
	buf[2] = byte(dataLen >> 16 & 0xFF)
	buf[3] = byte(dataLen >> 32 & 0xFF)
	buf[4] = byte(p.id & 0xFF)
	buf[5] = byte(p.id >> 8 & 0xFF)
	buf[6] = byte(p.id >> 16 & 0xFF)
	buf[7] = byte(p.id >> 24 & 0xFF)
	copy(buff[:DataLen+MessageIDLen], buf)
	copy(buff[DataLen+MessageIDLen:], p.data)
	return buff
}

func (p *Packet) Unmarshal(m interface{}) error {
	return proto.Unmarshal(p.data, m.(proto.Message))
}

func NewPacket(id uint32, msg interface{}) *Packet {

	p := &Packet{
		id: id,
	}

	switch v := msg.(type) {
	case []byte:
		p.data = v
	case proto.Message:
		if mdata, err := proto.Marshal(v); err == nil {
			p.data = mdata
		} else {
			l4g.Error("[NewPacket] proto marshal msg: %d error: %v",
				id, err)
			return nil
		}
	case nil:
	default:
		l4g.Error("[NewPacket] error msg type msg: %d", id)
		return nil
	}

	return p
}

type MsgProtocol struct {
}

func (p *MsgProtocol) ReadPacket(r io.Reader) (IPacket, error) /*Packet*/ {

	buff := make([]byte, 4)
	// data length
	if _, err := io.ReadFull(r, buff); err != nil {
		return nil, err
	}

	dataLen := binary.LittleEndian.Uint32(buff)
	if dataLen > MaxPacketLen {
		return nil, errors.New("data max")
	}

	buff = make([]byte, 4)
	if _, err := io.ReadFull(r, buff); err != nil {
		return nil, err
	}
	dataId := binary.LittleEndian.Uint32(buff)

	// id
	msg := &Packet{
		id: dataId,
	}

	// data
	if dataLen > 0 {
		msg.data = make([]byte, dataLen)
		if _, err := io.ReadFull(r, msg.data); err != nil {
			return nil, err
		}
	}

	return msg, nil
}

func (p *MsgProtocol) ReadPacketByByte(bufs []byte) (IPacket, error) /*Packet*/ {
	dataLen := binary.LittleEndian.Uint32(bufs[:4])
	if dataLen > MaxPacketLen {
		return nil, errors.New("data max")
	}

	dataId := binary.LittleEndian.Uint32(bufs[4:8])
	// id
	msg := &Packet{
		id: dataId,
	}

	// data
	if dataLen > 0 {
		msg.data = bufs[DataLen+MessageIDLen:]
	}

	return msg, nil
}
