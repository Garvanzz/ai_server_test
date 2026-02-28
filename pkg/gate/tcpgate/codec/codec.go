package codec

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"xfx/pkg/log"
	"xfx/pkg/net/tcp"
	"xfx/pkg/serialize"
	"xfx/pkg/serialize/pb"

	"github.com/gogo/protobuf/proto"
)

const (
	MaxSize     = 128 * 1024
	HeadLenSize = 4 //头部长度大小
	ProtoIdSize = 4 //ProtoID大小
)

var _ tcp.Codec = (*Parser)(nil)

func NewParser(conn io.ReadWriter, router tcp.MsgRouter) (tcp.Codec, error) {
	return &Parser{
		conn:       conn.(net.Conn),
		serializer: pb.NewSerializer(),
		msgRouter:  router,
	}, nil
}

type Parser struct {
	conn       net.Conn
	serializer serialize.Serializer
	//agent      gate.Agent
	msgRouter tcp.MsgRouter
}

func (p *Parser) Send(msg any) error {
	//fmt.Println(msg.(proto.Message).String())
	data, err := p.serializer.Marshal(msg)
	if err != nil {
		return fmt.Errorf("msgParser marshal msg error: %v", err)
	}

	pb, ok := msg.(proto.Message)
	if !ok {
		return fmt.Errorf("msgParser message type error")
	}

	id, err := p.msgRouter.MessageID(pb)
	if err != nil || id <= 0 {
		return fmt.Errorf("msgParser message id error:%v", id)
	}

	b, err := Encode(id, data)
	if err != nil {
		return err
	}

	length, err := p.conn.Write(b)
	if err != nil {
		return err
	}

	if length <= 0 {
		return fmt.Errorf("msgParser send data length error:%v", length)
	}

	return nil
}

func (p *Parser) Receive() (any, error) {
	lenData := make([]byte, HeadLenSize)

	if _, err := io.ReadFull(p.conn, lenData); err != nil {
		//log.Info("read io read full, read lenData err returned, lenData : %v", err)
		return nil, err
	}

	dataLen := binary.LittleEndian.Uint32(lenData)
	if int(dataLen) > MaxSize {
		return nil, errors.New("message too long")
	}

	contentData := make([]byte, dataLen+ProtoIdSize)
	if _, err := io.ReadFull(p.conn, contentData); err != nil {
		//log.Error("read io read full, read contentData err returned")
		return nil, err
	}

	protoId := binary.LittleEndian.Uint32(contentData)

	content := contentData[ProtoIdSize:]

	pb, err := p.msgRouter.NewMessage(protoId)
	if err != nil {
		return nil, fmt.Errorf("proto id new message error:%v", protoId)
	}

	err = p.serializer.Unmarshal(content, pb)
	if err != nil {
		return nil, fmt.Errorf("Session: unmarshal message error %v,message type =%v \n", err, protoId)
	}

	//if _, ok := pb.(*proto_activity.C2SLadderRaceSetLineUp); ok {
	//	return nil, fmt.Errorf("Session: unmarshal message test\n")
	//}

	return pb, nil
}

func Encode(typ uint32, data []byte) ([]byte, error) {
	outBuff := new(bytes.Buffer)
	dataLen := len(data)

	err := binary.Write(outBuff, binary.LittleEndian, uint32(dataLen))
	if err != nil {
		return nil, err
	}

	//写入protobuf id
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

//func (p *Parser) SetAgent(agent gate.Agent) {
//	p.agent = agent
//}

func (p *Parser) Close() (err error) {
	err = p.conn.Close()
	if err != nil {
		log.Error("codec connection close error:%v", err)
	}

	//if p.agent != nil {
	//	err = p.agent.Close()
	//	p.agent = nil
	//}
	return
}
