package serialize

import (
	"xfx/pkg/serialize/json"
	"xfx/pkg/serialize/jsonpb"
	"xfx/pkg/serialize/pb"
)

// Marshaler represents a marshal interface
type Marshaler interface {
	Marshal(interface{}) ([]byte, error)
}

// Unmarshaler represents a unmarshal interface
type Unmarshaler interface {
	Unmarshal([]byte, interface{}) error
}

// TypeMarshaler represents a marshal interface with type name returned
type TypeMarshaler interface {
	MarshalType(interface{}) (string, []byte, error)
}

// TypeUnmarshaler represents a unmarshal interace, with indicated type name and return the object
type TypeUnmarshaler interface {
	UnmarshalType(string, []byte) (interface{}, error)
}

// Serializer is the interface that groups the basic Marshal and Unmarshal methods.
type Serializer interface {
	Marshaler
	Unmarshaler
	TypeMarshaler
	TypeUnmarshaler
}

type jsonSerializer interface {
	Marshaler
	Unmarshaler
}

var (
	defaultSerializer Serializer
	Pb                Serializer
	Jsonpb            Serializer
	Json              jsonSerializer
)

func init() {
	Pb = pb.NewSerializer()
	Jsonpb = jsonpb.NewSerializer()
	defaultSerializer = Pb
	Json = json.NewSerializer()
}

func Marshal(v interface{}) ([]byte, error)             { return defaultSerializer.Marshal(v) }
func UnMarshal(data []byte, v interface{}) error        { return defaultSerializer.Unmarshal(data, v) }
func MarshalType(v interface{}) (string, []byte, error) { return defaultSerializer.MarshalType(v) }
func UnmarshalType(name string, data []byte) (interface{}, error) {
	return defaultSerializer.UnmarshalType(name, data)
}
