// Code generated by protoc-gen-go. DO NOT EDIT.
// source: plugin/options/bom.proto

package bom

import (
	fmt "fmt"
	proto "github.com/gogo/protobuf/proto"
	descriptor "github.com/gogo/protobuf/protoc-gen-gogo/descriptor"
	math "math"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.GoGoProtoPackageIsVersion2 // please upgrade the proto package

type BomFileOptions struct {
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *BomFileOptions) Reset()         { *m = BomFileOptions{} }
func (m *BomFileOptions) String() string { return proto.CompactTextString(m) }
func (*BomFileOptions) ProtoMessage()    {}
func (*BomFileOptions) Descriptor() ([]byte, []int) {
	return fileDescriptor_0098a2fdbc082952, []int{0}
}

func (m *BomFileOptions) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_BomFileOptions.Unmarshal(m, b)
}
func (m *BomFileOptions) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_BomFileOptions.Marshal(b, m, deterministic)
}
func (m *BomFileOptions) XXX_Merge(src proto.Message) {
	xxx_messageInfo_BomFileOptions.Merge(m, src)
}
func (m *BomFileOptions) XXX_Size() int {
	return xxx_messageInfo_BomFileOptions.Size(m)
}
func (m *BomFileOptions) XXX_DiscardUnknown() {
	xxx_messageInfo_BomFileOptions.DiscardUnknown(m)
}

var xxx_messageInfo_BomFileOptions proto.InternalMessageInfo

type MongoObject struct {
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *MongoObject) Reset()         { *m = MongoObject{} }
func (m *MongoObject) String() string { return proto.CompactTextString(m) }
func (*MongoObject) ProtoMessage()    {}
func (*MongoObject) Descriptor() ([]byte, []int) {
	return fileDescriptor_0098a2fdbc082952, []int{1}
}

func (m *MongoObject) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_MongoObject.Unmarshal(m, b)
}
func (m *MongoObject) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_MongoObject.Marshal(b, m, deterministic)
}
func (m *MongoObject) XXX_Merge(src proto.Message) {
	xxx_messageInfo_MongoObject.Merge(m, src)
}
func (m *MongoObject) XXX_Size() int {
	return xxx_messageInfo_MongoObject.Size(m)
}
func (m *MongoObject) XXX_DiscardUnknown() {
	xxx_messageInfo_MongoObject.DiscardUnknown(m)
}

var xxx_messageInfo_MongoObject proto.InternalMessageInfo

type BomMessageOptions struct {
	Model                *bool    `protobuf:"varint,1,req,name=model" json:"model,omitempty"`
	Crud                 *bool    `protobuf:"varint,2,opt,name=crud" json:"crud,omitempty"`
	Table                *string  `protobuf:"bytes,3,opt,name=table" json:"table,omitempty"`
	Collection           *string  `protobuf:"bytes,4,opt,name=collection" json:"collection,omitempty"`
	BoundMessage         *string  `protobuf:"bytes,5,opt,name=boundMessage" json:"boundMessage,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *BomMessageOptions) Reset()         { *m = BomMessageOptions{} }
func (m *BomMessageOptions) String() string { return proto.CompactTextString(m) }
func (*BomMessageOptions) ProtoMessage()    {}
func (*BomMessageOptions) Descriptor() ([]byte, []int) {
	return fileDescriptor_0098a2fdbc082952, []int{2}
}

func (m *BomMessageOptions) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_BomMessageOptions.Unmarshal(m, b)
}
func (m *BomMessageOptions) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_BomMessageOptions.Marshal(b, m, deterministic)
}
func (m *BomMessageOptions) XXX_Merge(src proto.Message) {
	xxx_messageInfo_BomMessageOptions.Merge(m, src)
}
func (m *BomMessageOptions) XXX_Size() int {
	return xxx_messageInfo_BomMessageOptions.Size(m)
}
func (m *BomMessageOptions) XXX_DiscardUnknown() {
	xxx_messageInfo_BomMessageOptions.DiscardUnknown(m)
}

var xxx_messageInfo_BomMessageOptions proto.InternalMessageInfo

func (m *BomMessageOptions) GetModel() bool {
	if m != nil && m.Model != nil {
		return *m.Model
	}
	return false
}

func (m *BomMessageOptions) GetCrud() bool {
	if m != nil && m.Crud != nil {
		return *m.Crud
	}
	return false
}

func (m *BomMessageOptions) GetTable() string {
	if m != nil && m.Table != nil {
		return *m.Table
	}
	return ""
}

func (m *BomMessageOptions) GetCollection() string {
	if m != nil && m.Collection != nil {
		return *m.Collection
	}
	return ""
}

func (m *BomMessageOptions) GetBoundMessage() string {
	if m != nil && m.BoundMessage != nil {
		return *m.BoundMessage
	}
	return ""
}

type BomFieldOptions struct {
	Tag                  *BomTag  `protobuf:"bytes,1,opt,name=tag" json:"tag,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *BomFieldOptions) Reset()         { *m = BomFieldOptions{} }
func (m *BomFieldOptions) String() string { return proto.CompactTextString(m) }
func (*BomFieldOptions) ProtoMessage()    {}
func (*BomFieldOptions) Descriptor() ([]byte, []int) {
	return fileDescriptor_0098a2fdbc082952, []int{3}
}

func (m *BomFieldOptions) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_BomFieldOptions.Unmarshal(m, b)
}
func (m *BomFieldOptions) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_BomFieldOptions.Marshal(b, m, deterministic)
}
func (m *BomFieldOptions) XXX_Merge(src proto.Message) {
	xxx_messageInfo_BomFieldOptions.Merge(m, src)
}
func (m *BomFieldOptions) XXX_Size() int {
	return xxx_messageInfo_BomFieldOptions.Size(m)
}
func (m *BomFieldOptions) XXX_DiscardUnknown() {
	xxx_messageInfo_BomFieldOptions.DiscardUnknown(m)
}

var xxx_messageInfo_BomFieldOptions proto.InternalMessageInfo

func (m *BomFieldOptions) GetTag() *BomTag {
	if m != nil {
		return m.Tag
	}
	return nil
}

type BomTag struct {
	IsID                 *bool    `protobuf:"varint,3,opt,name=isID" json:"isID,omitempty"`
	Skip                 *bool    `protobuf:"varint,4,opt,name=skip" json:"skip,omitempty"`
	MongoObjectId        *bool    `protobuf:"varint,5,opt,name=mongoObjectId" json:"mongoObjectId,omitempty"`
	Update               *bool    `protobuf:"varint,6,opt,name=update" json:"update,omitempty"`
	Additional           *bool    `protobuf:"varint,7,opt,name=additional" json:"additional,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *BomTag) Reset()         { *m = BomTag{} }
func (m *BomTag) String() string { return proto.CompactTextString(m) }
func (*BomTag) ProtoMessage()    {}
func (*BomTag) Descriptor() ([]byte, []int) {
	return fileDescriptor_0098a2fdbc082952, []int{4}
}

func (m *BomTag) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_BomTag.Unmarshal(m, b)
}
func (m *BomTag) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_BomTag.Marshal(b, m, deterministic)
}
func (m *BomTag) XXX_Merge(src proto.Message) {
	xxx_messageInfo_BomTag.Merge(m, src)
}
func (m *BomTag) XXX_Size() int {
	return xxx_messageInfo_BomTag.Size(m)
}
func (m *BomTag) XXX_DiscardUnknown() {
	xxx_messageInfo_BomTag.DiscardUnknown(m)
}

var xxx_messageInfo_BomTag proto.InternalMessageInfo

func (m *BomTag) GetIsID() bool {
	if m != nil && m.IsID != nil {
		return *m.IsID
	}
	return false
}

func (m *BomTag) GetSkip() bool {
	if m != nil && m.Skip != nil {
		return *m.Skip
	}
	return false
}

func (m *BomTag) GetMongoObjectId() bool {
	if m != nil && m.MongoObjectId != nil {
		return *m.MongoObjectId
	}
	return false
}

func (m *BomTag) GetUpdate() bool {
	if m != nil && m.Update != nil {
		return *m.Update
	}
	return false
}

func (m *BomTag) GetAdditional() bool {
	if m != nil && m.Additional != nil {
		return *m.Additional
	}
	return false
}

type AutoServerOptions struct {
	Autogen              *bool    `protobuf:"varint,1,opt,name=autogen" json:"autogen,omitempty"`
	TxnMiddleware        *bool    `protobuf:"varint,2,opt,name=txn_middleware,json=txnMiddleware" json:"txn_middleware,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *AutoServerOptions) Reset()         { *m = AutoServerOptions{} }
func (m *AutoServerOptions) String() string { return proto.CompactTextString(m) }
func (*AutoServerOptions) ProtoMessage()    {}
func (*AutoServerOptions) Descriptor() ([]byte, []int) {
	return fileDescriptor_0098a2fdbc082952, []int{5}
}

func (m *AutoServerOptions) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_AutoServerOptions.Unmarshal(m, b)
}
func (m *AutoServerOptions) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_AutoServerOptions.Marshal(b, m, deterministic)
}
func (m *AutoServerOptions) XXX_Merge(src proto.Message) {
	xxx_messageInfo_AutoServerOptions.Merge(m, src)
}
func (m *AutoServerOptions) XXX_Size() int {
	return xxx_messageInfo_AutoServerOptions.Size(m)
}
func (m *AutoServerOptions) XXX_DiscardUnknown() {
	xxx_messageInfo_AutoServerOptions.DiscardUnknown(m)
}

var xxx_messageInfo_AutoServerOptions proto.InternalMessageInfo

func (m *AutoServerOptions) GetAutogen() bool {
	if m != nil && m.Autogen != nil {
		return *m.Autogen
	}
	return false
}

func (m *AutoServerOptions) GetTxnMiddleware() bool {
	if m != nil && m.TxnMiddleware != nil {
		return *m.TxnMiddleware
	}
	return false
}

type MethodOptions struct {
	ObjectType           *string  `protobuf:"bytes,1,opt,name=object_type,json=objectType" json:"object_type,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *MethodOptions) Reset()         { *m = MethodOptions{} }
func (m *MethodOptions) String() string { return proto.CompactTextString(m) }
func (*MethodOptions) ProtoMessage()    {}
func (*MethodOptions) Descriptor() ([]byte, []int) {
	return fileDescriptor_0098a2fdbc082952, []int{6}
}

func (m *MethodOptions) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_MethodOptions.Unmarshal(m, b)
}
func (m *MethodOptions) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_MethodOptions.Marshal(b, m, deterministic)
}
func (m *MethodOptions) XXX_Merge(src proto.Message) {
	xxx_messageInfo_MethodOptions.Merge(m, src)
}
func (m *MethodOptions) XXX_Size() int {
	return xxx_messageInfo_MethodOptions.Size(m)
}
func (m *MethodOptions) XXX_DiscardUnknown() {
	xxx_messageInfo_MethodOptions.DiscardUnknown(m)
}

var xxx_messageInfo_MethodOptions proto.InternalMessageInfo

func (m *MethodOptions) GetObjectType() string {
	if m != nil && m.ObjectType != nil {
		return *m.ObjectType
	}
	return ""
}

var E_FileOpts = &proto.ExtensionDesc{
	ExtendedType:  (*descriptor.FileOptions)(nil),
	ExtensionType: (*BomFileOptions)(nil),
	Field:         99432,
	Name:          "bom.file_opts",
	Tag:           "bytes,99432,opt,name=file_opts",
	Filename:      "plugin/options/bom.proto",
}

var E_Opts = &proto.ExtensionDesc{
	ExtendedType:  (*descriptor.MessageOptions)(nil),
	ExtensionType: (*BomMessageOptions)(nil),
	Field:         99432,
	Name:          "bom.opts",
	Tag:           "bytes,99432,opt,name=opts",
	Filename:      "plugin/options/bom.proto",
}

var E_Field = &proto.ExtensionDesc{
	ExtendedType:  (*descriptor.FieldOptions)(nil),
	ExtensionType: (*BomFieldOptions)(nil),
	Field:         99432,
	Name:          "bom.field",
	Tag:           "bytes,99432,opt,name=field",
	Filename:      "plugin/options/bom.proto",
}

var E_Server = &proto.ExtensionDesc{
	ExtendedType:  (*descriptor.ServiceOptions)(nil),
	ExtensionType: (*AutoServerOptions)(nil),
	Field:         99432,
	Name:          "bom.server",
	Tag:           "bytes,99432,opt,name=server",
	Filename:      "plugin/options/bom.proto",
}

var E_Method = &proto.ExtensionDesc{
	ExtendedType:  (*descriptor.MethodOptions)(nil),
	ExtensionType: (*MethodOptions)(nil),
	Field:         99432,
	Name:          "bom.method",
	Tag:           "bytes,99432,opt,name=method",
	Filename:      "plugin/options/bom.proto",
}

func init() {
	proto.RegisterType((*BomFileOptions)(nil), "bom.BomFileOptions")
	proto.RegisterType((*MongoObject)(nil), "bom.MongoObject")
	proto.RegisterType((*BomMessageOptions)(nil), "bom.BomMessageOptions")
	proto.RegisterType((*BomFieldOptions)(nil), "bom.BomFieldOptions")
	proto.RegisterType((*BomTag)(nil), "bom.BomTag")
	proto.RegisterType((*AutoServerOptions)(nil), "bom.AutoServerOptions")
	proto.RegisterType((*MethodOptions)(nil), "bom.MethodOptions")
	proto.RegisterExtension(E_FileOpts)
	proto.RegisterExtension(E_Opts)
	proto.RegisterExtension(E_Field)
	proto.RegisterExtension(E_Server)
	proto.RegisterExtension(E_Method)
}

func init() { proto.RegisterFile("plugin/options/bom.proto", fileDescriptor_0098a2fdbc082952) }

var fileDescriptor_0098a2fdbc082952 = []byte{
	// 505 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x74, 0x52, 0x4d, 0x6f, 0xd3, 0x40,
	0x10, 0x95, 0xdb, 0xc4, 0x75, 0x26, 0xa4, 0xd0, 0xa5, 0xaa, 0x56, 0x88, 0xb6, 0x91, 0x05, 0x52,
	0x4e, 0x49, 0xc5, 0x31, 0x37, 0x22, 0x84, 0x54, 0x41, 0x54, 0x58, 0x72, 0x8f, 0x6c, 0xef, 0xc6,
	0x2c, 0xac, 0x3d, 0x96, 0xbd, 0x86, 0xf6, 0x1f, 0x70, 0xe2, 0xca, 0x5f, 0xe4, 0x67, 0xa0, 0x1d,
	0xc7, 0xf9, 0x50, 0xc3, 0x6d, 0xe7, 0xbd, 0xd1, 0xdb, 0x37, 0x6f, 0x06, 0x78, 0x61, 0xea, 0x54,
	0xe7, 0x13, 0x2c, 0xac, 0xc6, 0xbc, 0x9a, 0xc4, 0x98, 0x8d, 0x8b, 0x12, 0x2d, 0xb2, 0xe3, 0x18,
	0xb3, 0x17, 0xc3, 0x14, 0x31, 0x35, 0x6a, 0x42, 0x50, 0x5c, 0xaf, 0x26, 0x52, 0x55, 0x49, 0xa9,
	0x0b, 0x8b, 0x65, 0xd3, 0x16, 0x3e, 0x83, 0xd3, 0x19, 0x66, 0xef, 0xb5, 0x51, 0x77, 0x8d, 0x44,
	0x38, 0x80, 0xfe, 0x1c, 0xf3, 0x14, 0xef, 0xe2, 0x6f, 0x2a, 0xb1, 0xe1, 0x1f, 0x0f, 0xce, 0x66,
	0x98, 0xcd, 0x55, 0x55, 0x45, 0x69, 0xdb, 0xc4, 0xce, 0xa1, 0x9b, 0xa1, 0x54, 0x86, 0x7b, 0xc3,
	0xa3, 0x51, 0x20, 0x9a, 0x82, 0x31, 0xe8, 0x24, 0x65, 0x2d, 0xf9, 0xd1, 0xd0, 0x1b, 0x05, 0x82,
	0xde, 0xae, 0xd3, 0x46, 0xb1, 0x51, 0xfc, 0x78, 0xe8, 0x8d, 0x7a, 0xa2, 0x29, 0xd8, 0x15, 0x40,
	0x82, 0xc6, 0xa8, 0xc4, 0xc9, 0xf1, 0x0e, 0x51, 0x3b, 0x08, 0x0b, 0xe1, 0x49, 0x8c, 0x75, 0x2e,
	0xd7, 0xdf, 0xf2, 0x2e, 0x75, 0xec, 0x61, 0xe1, 0x0d, 0x3c, 0x25, 0xeb, 0xca, 0xc8, 0xd6, 0xd6,
	0x25, 0x1c, 0xdb, 0x28, 0xe5, 0xde, 0xd0, 0x1b, 0xf5, 0xdf, 0xf4, 0xc7, 0x2e, 0x8d, 0x19, 0x66,
	0x8b, 0x28, 0x15, 0x0e, 0x0f, 0x7f, 0x7b, 0xe0, 0x37, 0xb5, 0xb3, 0xaa, 0xab, 0xdb, 0x77, 0xe4,
	0x2a, 0x10, 0xf4, 0x76, 0x58, 0xf5, 0x5d, 0x17, 0x64, 0x27, 0x10, 0xf4, 0x66, 0xaf, 0x60, 0x90,
	0x6d, 0xd3, 0xb8, 0x95, 0xe4, 0x24, 0x10, 0xfb, 0x20, 0xbb, 0x00, 0xbf, 0x2e, 0x64, 0x64, 0x15,
	0xf7, 0x89, 0x5e, 0x57, 0x6e, 0xcc, 0x48, 0x4a, 0xed, 0xcc, 0x45, 0x86, 0x9f, 0x10, 0xb7, 0x83,
	0x84, 0x0b, 0x38, 0x7b, 0x5b, 0x5b, 0xfc, 0xa2, 0xca, 0x1f, 0xaa, 0x6c, 0x87, 0xe0, 0x70, 0x12,
	0xd5, 0x16, 0x53, 0x95, 0xd3, 0x20, 0x81, 0x68, 0x4b, 0xf6, 0x1a, 0x4e, 0xed, 0x7d, 0xbe, 0xcc,
	0xb4, 0x94, 0x46, 0xfd, 0x8c, 0x4a, 0xb5, 0x4e, 0x7a, 0x60, 0xef, 0xf3, 0xf9, 0x06, 0x0c, 0x6f,
	0x60, 0x30, 0x57, 0xf6, 0x2b, 0x6e, 0x62, 0xb9, 0x86, 0x3e, 0x92, 0xd5, 0xa5, 0x7d, 0x28, 0x14,
	0xa9, 0xf6, 0x04, 0x34, 0xd0, 0xe2, 0xa1, 0x50, 0xd3, 0x4f, 0xd0, 0x5b, 0x69, 0xa3, 0x96, 0x58,
	0xd8, 0x8a, 0xbd, 0x1c, 0x37, 0x57, 0x33, 0x6e, 0xaf, 0x66, 0xbc, 0x73, 0x1e, 0xfc, 0xef, 0x2f,
	0x9f, 0xd2, 0x7d, 0xde, 0xa6, 0xbb, 0x43, 0x8a, 0x60, 0xd5, 0x14, 0xd5, 0x74, 0x0e, 0x1d, 0x12,
	0xbb, 0x7e, 0x24, 0xb6, 0x7f, 0x49, 0x1b, 0xbd, 0x8b, 0x56, 0x6f, 0x9f, 0x17, 0x24, 0x33, 0xfd,
	0x00, 0xdd, 0x95, 0x5b, 0x34, 0xbb, 0x3c, 0x60, 0x6e, 0x7b, 0x00, 0x1b, 0xb5, 0xf3, 0xad, 0xbb,
	0x2d, 0x2b, 0x1a, 0x8d, 0xe9, 0x67, 0xf0, 0x2b, 0x4a, 0xfc, 0x80, 0x3b, 0xb7, 0x0a, 0x9d, 0xfc,
	0xc7, 0xdd, 0xa3, 0x5d, 0x89, 0xb5, 0xd0, 0xf4, 0x23, 0xf8, 0x19, 0x45, 0xce, 0xae, 0x0e, 0x0c,
	0xbc, 0xb3, 0x8b, 0x8d, 0x22, 0x23, 0xc5, 0x3d, 0x4e, 0xac, 0x35, 0xfe, 0x05, 0x00, 0x00, 0xff,
	0xff, 0xc3, 0xc0, 0x21, 0xf7, 0xd6, 0x03, 0x00, 0x00,
}
