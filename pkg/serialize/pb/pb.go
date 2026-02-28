package pb

import (
	"errors"
	"fmt"
	"github.com/gogo/protobuf/proto"
	"reflect"
)

var ErrWrongValueType = errors.New("protobuf: convert on wrong type value")

// Serializer implements the serialize.Serializer interface
type Serializer struct{}

// NewSerializer returns a new Serializer.
func NewSerializer() *Serializer {
	return &Serializer{}
}

// Marshal returns the protobuf encoding of v.
func (s *Serializer) Marshal(v interface{}) ([]byte, error) {
	pb, ok := v.(proto.Message)
	if !ok {
		return nil, ErrWrongValueType
	}
	return proto.Marshal(pb)
}

// Unmarshal parses the protobuf-encoded data and stores the result
// in the value pointed to by v.
func (s *Serializer) Unmarshal(data []byte, v interface{}) error {
	pb, ok := v.(proto.Message)
	if !ok {
		return ErrWrongValueType
	}
	return proto.Unmarshal(data, pb)
}

// MarshalType returns the protobuf encoding of v.
func (s *Serializer) MarshalType(v interface{}) (string, []byte, error) {
	pb, ok := v.(proto.Message)
	if !ok {
		return "", nil, ErrWrongValueType
	}
	data, e := proto.Marshal(pb)
	return proto.MessageName(pb), data, e
}

// UnmarshalType parses the protobuf-encoded data and stores the result
// in the value pointed to by v.
func (s *Serializer) UnmarshalType(name string, data []byte) (interface{}, error) {
	typ := proto.MessageType(name)
	if typ == nil {
		return nil, fmt.Errorf("protobuf: invalid type name %s", name)
	}
	t := typ.Elem()
	ptr := reflect.New(t)
	pb := ptr.Interface().(proto.Message)

	err := s.Unmarshal(data, pb)
	return pb, err
}

// TypeName return the proto message's type name
func (s *Serializer) Name(msg interface{}) (string, error) {
	pb, ok := msg.(proto.Message)
	if !ok {
		return "", ErrWrongValueType
	}
	return proto.MessageName(pb), nil
}
