build-options:
	protoc -I/usr/local/include -I. \
	-I$(GOPATH)/src \
	-I$(GOPATH)/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis \
	--go_out=. \
	plugin/options/bom.proto

build:
	go build -o bin/protoc-gen-bom

	protoc -I/usr/local/include -I. \
	-I$(GOPATH)/src \
	-I$(GOPATH)/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis \
	--go_out=. \
	example/example.proto

	protoc -I/usr/local/include -I.  \
	-I$(GOPATH)/src   \
	-I$(GOPATH)/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis   \
	--plugin=protoc-gen-mongo=bin/protoc-gen-bom \
	--mongo_out="generateCrud=true,gateway:." \
 	example/example.proto