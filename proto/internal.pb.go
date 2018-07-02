// Code generated by protoc-gen-go. DO NOT EDIT.
// source: internal.proto

package server

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

type Type int32

const (
	Type_PUT Type = 0
	Type_GET Type = 1
)

var Type_name = map[int32]string{
	0: "PUT",
	1: "GET",
}
var Type_value = map[string]int32{
	"PUT": 0,
	"GET": 1,
}

func (x Type) String() string {
	return proto.EnumName(Type_name, int32(x))
}
func (Type) EnumDescriptor() ([]byte, []int) { return fileDescriptor1, []int{0} }

type Mutation struct {
	Key          []byte `protobuf:"bytes,1,opt,name=key,proto3" json:"key,omitempty"`
	Family       []byte `protobuf:"bytes,2,opt,name=family,proto3" json:"family,omitempty"`
	Col          []byte `protobuf:"bytes,3,opt,name=col,proto3" json:"col,omitempty"`
	Timestamp    int64  `protobuf:"varint,4,opt,name=timestamp" json:"timestamp,omitempty"`
	MutationType Type   `protobuf:"varint,5,opt,name=mutation_type,json=mutationType,enum=server.Type" json:"mutation_type,omitempty"`
}

func (m *Mutation) Reset()                    { *m = Mutation{} }
func (m *Mutation) String() string            { return proto.CompactTextString(m) }
func (*Mutation) ProtoMessage()               {}
func (*Mutation) Descriptor() ([]byte, []int) { return fileDescriptor1, []int{0} }

func (m *Mutation) GetKey() []byte {
	if m != nil {
		return m.Key
	}
	return nil
}

func (m *Mutation) GetFamily() []byte {
	if m != nil {
		return m.Family
	}
	return nil
}

func (m *Mutation) GetCol() []byte {
	if m != nil {
		return m.Col
	}
	return nil
}

func (m *Mutation) GetTimestamp() int64 {
	if m != nil {
		return m.Timestamp
	}
	return 0
}

func (m *Mutation) GetMutationType() Type {
	if m != nil {
		return m.MutationType
	}
	return Type_PUT
}

type OpRequest struct {
	Table    string    `protobuf:"bytes,1,opt,name=table" json:"table,omitempty"`
	Mutation *Mutation `protobuf:"bytes,2,opt,name=mutation" json:"mutation,omitempty"`
}

func (m *OpRequest) Reset()                    { *m = OpRequest{} }
func (m *OpRequest) String() string            { return proto.CompactTextString(m) }
func (*OpRequest) ProtoMessage()               {}
func (*OpRequest) Descriptor() ([]byte, []int) { return fileDescriptor1, []int{1} }

func (m *OpRequest) GetTable() string {
	if m != nil {
		return m.Table
	}
	return ""
}

func (m *OpRequest) GetMutation() *Mutation {
	if m != nil {
		return m.Mutation
	}
	return nil
}

type OpResponse struct {
	Value []byte `protobuf:"bytes,1,opt,name=value,proto3" json:"value,omitempty"`
	Err   string `protobuf:"bytes,2,opt,name=err" json:"err,omitempty"`
}

func (m *OpResponse) Reset()                    { *m = OpResponse{} }
func (m *OpResponse) String() string            { return proto.CompactTextString(m) }
func (*OpResponse) ProtoMessage()               {}
func (*OpResponse) Descriptor() ([]byte, []int) { return fileDescriptor1, []int{2} }

func (m *OpResponse) GetValue() []byte {
	if m != nil {
		return m.Value
	}
	return nil
}

func (m *OpResponse) GetErr() string {
	if m != nil {
		return m.Err
	}
	return ""
}

func init() {
	proto.RegisterType((*Mutation)(nil), "server.Mutation")
	proto.RegisterType((*OpRequest)(nil), "server.OpRequest")
	proto.RegisterType((*OpResponse)(nil), "server.OpResponse")
	proto.RegisterEnum("server.Type", Type_name, Type_value)
}

func init() { proto.RegisterFile("internal.proto", fileDescriptor1) }

var fileDescriptor1 = []byte{
	// 261 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x44, 0x50, 0xc1, 0x4a, 0xc4, 0x30,
	0x14, 0x34, 0x76, 0xb7, 0xb6, 0xcf, 0xba, 0x94, 0x20, 0x92, 0x83, 0x87, 0xd2, 0x53, 0x11, 0x29,
	0xb8, 0xfa, 0x05, 0x82, 0x78, 0x92, 0x95, 0x50, 0x2f, 0x5e, 0x24, 0xbb, 0x3c, 0xa1, 0x98, 0x36,
	0x31, 0x49, 0x17, 0xfa, 0x25, 0xfe, 0xae, 0x24, 0x6d, 0xf5, 0x36, 0xf3, 0x98, 0x4c, 0x66, 0x06,
	0x36, 0x6d, 0xef, 0xd0, 0xf4, 0x42, 0xd6, 0xda, 0x28, 0xa7, 0x68, 0x6c, 0xd1, 0x1c, 0xd1, 0x94,
	0x3f, 0x04, 0x92, 0x97, 0xc1, 0x09, 0xd7, 0xaa, 0x9e, 0xe6, 0x10, 0x7d, 0xe1, 0xc8, 0x48, 0x41,
	0xaa, 0x8c, 0x7b, 0x48, 0xaf, 0x20, 0xfe, 0x14, 0x5d, 0x2b, 0x47, 0x76, 0x1a, 0x8e, 0x33, 0xf3,
	0xca, 0x83, 0x92, 0x2c, 0x9a, 0x94, 0x07, 0x25, 0xe9, 0x35, 0xa4, 0xae, 0xed, 0xd0, 0x3a, 0xd1,
	0x69, 0xb6, 0x2a, 0x48, 0x15, 0xf1, 0xff, 0x03, 0xbd, 0x83, 0x8b, 0x6e, 0xfe, 0xe5, 0xc3, 0x8d,
	0x1a, 0xd9, 0xba, 0x20, 0xd5, 0x66, 0x9b, 0xd5, 0x53, 0x8c, 0xba, 0x19, 0x35, 0xf2, 0x6c, 0x91,
	0x78, 0x56, 0xee, 0x20, 0xdd, 0x69, 0x8e, 0xdf, 0x03, 0x5a, 0x47, 0x2f, 0x61, 0xed, 0xc4, 0x5e,
	0x62, 0xc8, 0x96, 0xf2, 0x89, 0xd0, 0x5b, 0x48, 0x96, 0x27, 0x21, 0xdf, 0xf9, 0x36, 0x5f, 0x0c,
	0x97, 0x4e, 0xfc, 0x4f, 0x51, 0x3e, 0x00, 0x78, 0x43, 0xab, 0x55, 0x6f, 0xd1, 0x3b, 0x1e, 0x85,
	0x1c, 0x70, 0x6e, 0x3b, 0x11, 0xdf, 0x0b, 0x8d, 0x09, 0x66, 0x29, 0xf7, 0xf0, 0x86, 0xc1, 0xca,
	0xc7, 0xa1, 0x67, 0x10, 0xbd, 0xbe, 0x35, 0xf9, 0x89, 0x07, 0xcf, 0x4f, 0x4d, 0x4e, 0x1e, 0x93,
	0xf7, 0x79, 0xc4, 0x7d, 0x1c, 0x36, 0xbd, 0xff, 0x0d, 0x00, 0x00, 0xff, 0xff, 0xaf, 0xd5, 0x82,
	0xd5, 0x65, 0x01, 0x00, 0x00,
}