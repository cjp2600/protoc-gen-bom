package main

import (
	"github.com/cjp2600/protoc-gen-bom/plugin"
	"github.com/gogo/protobuf/vanity/command"
)

func main() {
	plugin := &plugin.MongoPlugin{}
	response := command.GeneratePlugin(command.Read(), plugin, ".pb.bom.go")
	command.Write(response)
}