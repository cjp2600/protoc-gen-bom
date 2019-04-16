### Example

```$xslt
syntax = "proto3";

import "github.com/cjp2600/protoc-gen-bom/plugin/options/bom.proto";
import "google/protobuf/timestamp.proto";
import "github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis/google/api/annotations.proto";

package main;

// user types existing in the system
enum UserTypes {
    pupil = 0; // the main user type is set by default, type is issued for customers of the service who perform training
    teacher = 1; // type of teacher
    admin = 4;
}

message Role {
    option (bom.opts) = {
         model: true
         crud: true
     };

    string id = 1 [(bom.field).tag = {mongoObjectId:true isID:true}];
    string name = 2;
    repeated Permission role = 3;
}

message Permission {
    option (bom.opts) = {
         model: true
         crud: true
     };
    string id = 1 [(bom.field).tag = {mongoObjectId:true isID:true}];

    string service = 3;
    bool create = 4;
    bool read = 5;
    bool update = 6;
    bool delete = 7;
}

// user base model
message User {
    option (bom.opts) = {
         model: true
         crud: true
         collection: "user"
     };

    string id = 1 [(bom.field).tag = {mongoObjectId:true isID:true}];
    bool active = 2;
    string firstName = 3;
    string lastName = 4;
    string phone = 6;
    string email = 7;
    Role role = 9;
    bool EmailConfirm = 10;
    UserTypes type = 11;
    Token token = 12;
    google.protobuf.Timestamp createdAt = 13;
    google.protobuf.Timestamp updatedAt = 14;
}

message Token {
    option (bom.opts) = {model: true};
    string accessToken = 3;
    string refreshToken = 4;
}


message ProviderUsers {
    option (bom.opts) = {
         model: true
         crud: true
     };

    string providerId = 1 [(bom.field).tag = {mongoObjectId: true}];
    string userId = 2 [(bom.field).tag = {mongoObjectId: true}];
}

```