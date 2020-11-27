## Protoc-gen-bom
A protobuf compiler plugin designed to generate MongoDB models and APIs for simple object persistence tasks

## Example 

Describe the essence of what we want to work with.

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
    bool active = 2;
    string firstName = 3;
    string lastName = 4;
    string secondName = 5;
    string phone = 6;
    string email = 11;
    UserTypes type = 12;
}

```

After generation, we have the following methods available

```go

// create MongoDB Model from protobuf (UserMongo)
type UserMongo struct {
	Id              primitive.ObjectID          `_id, omitempty`
	Active          bool                        `json:"active"`
	FirstName       string                      `json:"firstname"`
	LastName        string                      `json:"lastname"`
	SecondName      string                      `json:"secondname"`
	Phone           string                      `json:"phone"`
	Email           string                      `json:"email"`
	Type            UserTypes                   `json:"type"`
	bom             *bom.Bom
}

```
### ToMongo
method of conversion from protobuf of an object to a mongo object

### ToPB 
method of conversion from mongo of an object to a protobuf object

### InsertOne
```go
	query := item.ToMongo()
	p, err := q.InsertOne()
	if err != nil {
		return nil, err
	}
```

### FindOne
```go
    // func (e *UserMongo) FindOne() (*UserMongo, error) {
	user, err := pb.NewUserMongo().WhereId("5f3a4ea2e97e882308d8f5ac").FindOne()
	if err != nil {
		return
	}
```

### List
```go
	// func (e *UserMongo) List() ([]*UserMongo, error) {
	users, err := pb.NewUserMongo().List()
	if err != nil {
		return
	}
```

### ListWithPagination
```go
	// func (e *UserMongo) ListWithPagination() ([]*UserMongo, *bom.Pagination, error) {
	user, pagination, err := pb.NewUserMongo().ListWithPagination()
	if err != nil {
		return
	}
```
### FindOneById
```go
	// func (e *UserMongo) FindOneById(Id string) (*UserMongo, error) {
	user, err := pb.NewUserMongo().FindOneById("5f3a4ea2e97e882308d8f5ac")
	if err != nil {
		return
	}
```
### FindOneById
```go
	// func (e *UserMongo) FindOneById(Id string) (*UserMongo, error) {
	user, err := pb.NewUserMongo().FindOneById("5f3a4ea2e97e882308d8f5ac")
	if err != nil {
		return
	}
```
### GetBulk
```go
	// func (e *UserMongo) GetBulk(ids []string) ([]*UserMongo, error) {
	user, err := pb.NewUserMongo().GetBulk([]string{"5f3a4ea2e97e882308d8f5ac","1f3a4ea2e97e882308d8f5ac"})
	if err != nil {
		return
	}
```

### ListWithLastID
```go
	// func (e *UserMongo) GetBulk(ids []string) ([]*UserMongo, error) {
	user, err := pb.NewUserMongo().GetBulk([]string{"5f3a4ea2e97e882308d8f5ac","1f3a4ea2e97e882308d8f5ac"})
	if err != nil {
		return
	}
```
