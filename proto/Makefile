# server:
# 	cd ./innerproto && make build
# 	cd ../../proto && make server
# 	go run ./buildroutes.go -toserver -path "./Proto_Login,./Proto_Player,./Proto_Game,./Proto_Game_Durak" -outpath "./protodefine/server_define.go"

# robot:
# 	cd ./innerproto && make build
# 	cd ../../proto && make server
# 	go run ./buildroutes.go  -torobot -path "./Proto_Login,./Proto_Player,./Proto_Game,./Proto_Game_Durak" -outpath "./protodefine/robot_define.go"

# mac:
# 	cd ./innerproto && make mac
# 	cd ../../proto && make mac
# 	go run ./buildroutes.go -toserver -path "./Proto_Login,./Proto_Player,./Proto_Game,./Proto_Game_Durak" -outpath "./protodefine/server_define.go"

all:
	protoc -I=. \
	--gogoslick_out=Mgoogle/protobuf/any.proto=github.com/gogo/protobuf/types,plugins=grpc:. login.proto 
win:
	protoc -I=../../proto --gogoslick_out=./proto_login ../../proto/login.proto
	protoc -I=../../proto --gogoslick_out=./proto_player ../../proto/player.proto
	protoc -I=../../proto --gogoslick_out=./proto_game ../../proto/game.proto

node:
	protoc --gogoslick_out=Mgoogle/protobuf/any.proto=github.com/gogo/protobuf/types:./ ./proto_node/node.proto
