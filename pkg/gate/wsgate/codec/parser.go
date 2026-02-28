package codec

import (
	"fmt"
	"github.com/gogo/protobuf/proto"
	"xfx/pkg/log"
	"xfx/pkg/serialize"
	"xfx/pkg/serialize/pb"
	"xfx/proto"
)

func NewParser() *Parser {
	return &Parser{
		serializer: pb.NewSerializer(),
		decoder:    NewDecoder(),
	}
}

type Parser struct {
	serializer serialize.Serializer
	decoder    *Decoder
}

func (p *Parser) EncodeMsg(msg any) ([]byte, error) {
	data, err := p.serializer.Marshal(msg)
	if err != nil {
		return nil, fmt.Errorf("msgParser marshal msg error: %v", err)
	}

	pb, ok := msg.(proto.Message)
	if !ok {
		return nil, fmt.Errorf("msgParser message type error")
	}

	id, err := proto_id.MessageID(pb)
	if err != nil || id <= 0 {
		return nil, fmt.Errorf("msgParser message id error:%v", id)
	}

	b, err := Encode(id, data)
	if err != nil {
		return nil, err
	}

	return b, nil
}

func (p *Parser) DecodeMsg(data []byte) (any, error) {
	packets, err := p.decoder.Decode(data)
	if err != nil {
		return nil, fmt.Errorf("session: recv %v", err)
	}

	pbs := make([]proto.Message, 0, len(packets))
	for _, packet := range packets {
		log.Debug("* Session: recv packet: %v", packet.String())
		pb, err := proto_id.NewMessage(packet.Type)
		if err != nil {
			log.Error("Session: new message type error %v,message type =%v \n", err, packet.Type)
			continue
		}

		err = p.serializer.Unmarshal(packet.Data, pb)
		if err != nil {
			log.Error("Session: unmarshal message error %v,message type =%v \n", err, packet.Type)
			continue
		}

		pbs = append(pbs, pb)
	}
	return pbs, nil
}
