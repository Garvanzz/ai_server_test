protoc -I=. -I=$GOPATH/src --gogoslick_out=\
Mgoogle/protobuf/any.proto=github.com/gogo/protobuf/types,plugins=grpc,\
Mgoogle/protobuf/duration.proto=github.com/gogo/protobuf/types,\
Mgoogle/protobuf/timestamp.proto=github.com/gogo/protobuf/types:. protos.proto