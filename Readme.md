## protoc-gen-bom


### Example

```protobuf
syntax = "proto3";

import "github.com/cjp2600/protoc-gen-bom/plugin/options/bom.proto";

package main;

enum UserTypes {
    user = 0;
    admin = 1;
}

// describe the user model  
// 
message User {

    // define a message as a model
    option (bom.opts) = {
         model: true
         crud: true // methods are needed to redefine
     };

    // Set the basic id using ObjectID
    string id = 1 [(bom.field).tag = {mongoObjectId:true isID:true}];
    string firstName = 3;
    string lastName = 4;
    string secondName = 5;
    string phone = 6;
    string email = 11;
    UserTypes type = 12;
}

```
