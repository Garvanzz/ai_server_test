package json

import (
	"encoding/json"
)

// Serializer implements the serialize.Serializer interface
type Serializer struct{}

// NewSerializer returns a new Serializer.
func NewSerializer() *Serializer {
	return &Serializer{}
}

// Marshal returns the JSON encoding of v.
func (s *Serializer) Marshal(v interface{}) ([]byte, error) { return json.Marshal(v) }

// Unmarshal parses the JSON-encoded data and stores the result
// in the value pointed to by v.
func (s *Serializer) Unmarshal(data []byte, v interface{}) error { return json.Unmarshal(data, v) }

// func (s *Serializer) MarshalType(v interface{}) (string, []byte, error) {
// 	name := reflect.ValueOf(v).Type().Name()
// 	d, e := s.Marshal(v)
// 	return name, d, e
// }

// func (s *Serializer) UnmarshalType(name string, data byte) (interface{}, error) {
// 	reflect.
// }
