package proto_id

import (
	"fmt"
	"github.com/golang/protobuf/proto"
	"reflect"
	"xfx/pkg/log"
)

type MsgRouter struct {
	msgInfo map[uint32]reflect.Type
	msgId   map[reflect.Type]uint32
}

func NewMsgRouter() *MsgRouter {
	return &MsgRouter{
		msgInfo: make(map[uint32]reflect.Type),
		msgId:   make(map[reflect.Type]uint32),
	}
}

func (r *MsgRouter) Register(msg any, msgID uint32) uint32 {
	if _, ok := r.msgInfo[msgID]; ok {
		log.Fatal("message %v is already registered", msgID)
	}

	msgType := reflect.TypeOf(msg)
	if msgType == nil || msgType.Kind() != reflect.Ptr {
		log.Fatal("protobuf message pointer required")
	}

	r.msgInfo[msgID] = msgType
	r.msgId[msgType] = msgID

	return msgID
}

func (r *MsgRouter) NewMessage(id uint32) (proto.Message, error) {
	msgType, ok := r.msgInfo[id]
	if !ok {
		return nil, fmt.Errorf("message %v not registered", id)
	}

	msg := reflect.New(msgType.Elem()).Interface()

	return msg.(proto.Message), nil
}

func (r *MsgRouter) MessageID(msg proto.Message) (uint32, error) {
	msgType := reflect.TypeOf(msg)
	protoId, ok := r.msgId[msgType]
	if !ok {
		return 0, fmt.Errorf("message %s not registered", msgType)
	}
	return protoId, nil
}

func NewMessage(id uint32) (proto.Message, error) {
	msgType, ok := Router.msgInfo[id]
	if !ok {
		return nil, fmt.Errorf("message %v not registered", id)
	}

	msg := reflect.New(msgType.Elem()).Interface()

	return msg.(proto.Message), nil
}

func MessageID(msg proto.Message) (uint32, error) {
	msgType := reflect.TypeOf(msg)
	protoId, ok := Router.msgId[msgType]
	if !ok {
		return 0, fmt.Errorf("message %s not registered", msgType)
	}
	return protoId, nil
}
