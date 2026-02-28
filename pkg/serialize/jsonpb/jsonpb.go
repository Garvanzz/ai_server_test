package jsonpb

import (
	"bytes"
	"errors"
	"fmt"
	"reflect"

	"github.com/gogo/protobuf/jsonpb"
	"github.com/gogo/protobuf/proto"
)

var ErrWrongValueType = errors.New("jsonpb: convert on wrong type value")

type JsonMessage struct {
	TypeName string
	Json     string
}

type Serializer struct {
	jsonpb.Marshaler
	jsonpb.Unmarshaler
}

func NewSerializer() *Serializer {
	return &Serializer{
		Marshaler: jsonpb.Marshaler{},
		Unmarshaler: jsonpb.Unmarshaler{
			AllowUnknownFields: true,
		},
	}
}

func (j *Serializer) Marshal(v interface{}) ([]byte, error) {
	if message, ok := v.(*JsonMessage); ok {
		return []byte(message.Json), nil
	} else if message, ok := v.(proto.Message); ok {
		str, err := j.Marshaler.MarshalToString(message)
		if err != nil {
			return nil, err
		}

		return []byte(str), nil
	}
	return nil, fmt.Errorf("msg must be proto.Message")
}

func (j *Serializer) Unmarshal(data []byte, v interface{}) error {
	if msg, ok := v.(*JsonMessage); ok {
		msg.Json = string(data)
		return nil
	}
	pb, ok := v.(proto.Message)
	if !ok {
		return ErrWrongValueType
	}

	r := bytes.NewReader(data)
	j.Unmarshaler.Unmarshal(r, pb)
	return nil
}

func (j *Serializer) MarshalType(v interface{}) (string, []byte, error) {
	if message, ok := v.(*JsonMessage); ok {
		return message.TypeName, []byte(message.Json), nil
	} else if message, ok := v.(proto.Message); ok {
		str, err := j.Marshaler.MarshalToString(message)
		if err != nil {
			return "", nil, err
		}

		return proto.MessageName(message), []byte(str), nil
	}
	return "", nil, fmt.Errorf("msg must be proto.Message")
}

func (j *Serializer) UnmarshalType(name string, data []byte) (interface{}, error) {
	protoType := proto.MessageType(name)
	if protoType == nil {
		m := &JsonMessage{
			TypeName: name,
			Json:     string(data),
		}
		return m, nil
	}
	t := protoType.Elem()
	ptr := reflect.New(t)
	pb, ok := ptr.Interface().(proto.Message)
	if !ok {
		return nil, ErrWrongValueType
	}
	r := bytes.NewReader(data)
	j.Unmarshaler.Unmarshal(r, pb)
	return pb, nil
}

func (j *Serializer) Name(v interface{}) (string, error) {
	msg, ok := v.(*JsonMessage)
	if ok {
		return msg.TypeName, nil
	}
	pb, ok := v.(proto.Message)
	if !ok {
		return "", ErrWrongValueType
	}
	return proto.MessageName(pb), nil
}
