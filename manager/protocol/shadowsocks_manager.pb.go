// Code generated by protoc-gen-go.
// source: shadowsocks_manager.proto
// DO NOT EDIT!

/*
Package protocol is a generated protocol buffer package.

It is generated from these files:
	shadowsocks_manager.proto

It has these top-level messages:
	ShadowsocksService
	ServiceList
	Statistics
	AllocateRequest
	AllocateResponse
	FreeRequest
	FreeResponse
	RegisterRequest
	RegisterResponse
*/
package protocol

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"
import google_protobuf "github.com/golang/protobuf/ptypes/empty"

import (
	context "golang.org/x/net/context"
	grpc "google.golang.org/grpc"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion2 // please upgrade the proto package

type ShadowsocksService struct {
	UserId   string `protobuf:"bytes,99,opt,name=user_id,json=userId" json:"user_id,omitempty"`
	Port     int32  `protobuf:"varint,1,opt,name=port" json:"port,omitempty"`
	Password string `protobuf:"bytes,2,opt,name=password" json:"password,omitempty"`
}

func (m *ShadowsocksService) Reset()                    { *m = ShadowsocksService{} }
func (m *ShadowsocksService) String() string            { return proto.CompactTextString(m) }
func (*ShadowsocksService) ProtoMessage()               {}
func (*ShadowsocksService) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

func (m *ShadowsocksService) GetUserId() string {
	if m != nil {
		return m.UserId
	}
	return ""
}

func (m *ShadowsocksService) GetPort() int32 {
	if m != nil {
		return m.Port
	}
	return 0
}

func (m *ShadowsocksService) GetPassword() string {
	if m != nil {
		return m.Password
	}
	return ""
}

type ServiceList struct {
	Services []*ShadowsocksService `protobuf:"bytes,1,rep,name=services" json:"services,omitempty"`
}

func (m *ServiceList) Reset()                    { *m = ServiceList{} }
func (m *ServiceList) String() string            { return proto.CompactTextString(m) }
func (*ServiceList) ProtoMessage()               {}
func (*ServiceList) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{1} }

func (m *ServiceList) GetServices() []*ShadowsocksService {
	if m != nil {
		return m.Services
	}
	return nil
}

type Statistics struct {
	Traffics map[int32]int64 `protobuf:"bytes,1,rep,name=traffics" json:"traffics,omitempty" protobuf_key:"varint,1,opt,name=key" protobuf_val:"varint,2,opt,name=value"`
}

func (m *Statistics) Reset()                    { *m = Statistics{} }
func (m *Statistics) String() string            { return proto.CompactTextString(m) }
func (*Statistics) ProtoMessage()               {}
func (*Statistics) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{2} }

func (m *Statistics) GetTraffics() map[int32]int64 {
	if m != nil {
		return m.Traffics
	}
	return nil
}

type AllocateRequest struct {
	ServiceList *ServiceList `protobuf:"bytes,1,opt,name=service_list,json=serviceList" json:"service_list,omitempty"`
}

func (m *AllocateRequest) Reset()                    { *m = AllocateRequest{} }
func (m *AllocateRequest) String() string            { return proto.CompactTextString(m) }
func (*AllocateRequest) ProtoMessage()               {}
func (*AllocateRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{3} }

func (m *AllocateRequest) GetServiceList() *ServiceList {
	if m != nil {
		return m.ServiceList
	}
	return nil
}

type AllocateResponse struct {
	ServiceList *ServiceList `protobuf:"bytes,1,opt,name=service_list,json=serviceList" json:"service_list,omitempty"`
}

func (m *AllocateResponse) Reset()                    { *m = AllocateResponse{} }
func (m *AllocateResponse) String() string            { return proto.CompactTextString(m) }
func (*AllocateResponse) ProtoMessage()               {}
func (*AllocateResponse) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{4} }

func (m *AllocateResponse) GetServiceList() *ServiceList {
	if m != nil {
		return m.ServiceList
	}
	return nil
}

type FreeRequest struct {
	ServiceList *ServiceList `protobuf:"bytes,1,opt,name=service_list,json=serviceList" json:"service_list,omitempty"`
}

func (m *FreeRequest) Reset()                    { *m = FreeRequest{} }
func (m *FreeRequest) String() string            { return proto.CompactTextString(m) }
func (*FreeRequest) ProtoMessage()               {}
func (*FreeRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{5} }

func (m *FreeRequest) GetServiceList() *ServiceList {
	if m != nil {
		return m.ServiceList
	}
	return nil
}

type FreeResponse struct {
	ServiceList *ServiceList `protobuf:"bytes,1,opt,name=service_list,json=serviceList" json:"service_list,omitempty"`
}

func (m *FreeResponse) Reset()                    { *m = FreeResponse{} }
func (m *FreeResponse) String() string            { return proto.CompactTextString(m) }
func (*FreeResponse) ProtoMessage()               {}
func (*FreeResponse) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{6} }

func (m *FreeResponse) GetServiceList() *ServiceList {
	if m != nil {
		return m.ServiceList
	}
	return nil
}

type RegisterRequest struct {
	SelfUrl string `protobuf:"bytes,1,opt,name=self_url,json=selfUrl" json:"self_url,omitempty"`
}

func (m *RegisterRequest) Reset()                    { *m = RegisterRequest{} }
func (m *RegisterRequest) String() string            { return proto.CompactTextString(m) }
func (*RegisterRequest) ProtoMessage()               {}
func (*RegisterRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{7} }

func (m *RegisterRequest) GetSelfUrl() string {
	if m != nil {
		return m.SelfUrl
	}
	return ""
}

type RegisterResponse struct {
	Token string `protobuf:"bytes,1,opt,name=token" json:"token,omitempty"`
}

func (m *RegisterResponse) Reset()                    { *m = RegisterResponse{} }
func (m *RegisterResponse) String() string            { return proto.CompactTextString(m) }
func (*RegisterResponse) ProtoMessage()               {}
func (*RegisterResponse) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{8} }

func (m *RegisterResponse) GetToken() string {
	if m != nil {
		return m.Token
	}
	return ""
}

func init() {
	proto.RegisterType((*ShadowsocksService)(nil), "protocol.ShadowsocksService")
	proto.RegisterType((*ServiceList)(nil), "protocol.ServiceList")
	proto.RegisterType((*Statistics)(nil), "protocol.Statistics")
	proto.RegisterType((*AllocateRequest)(nil), "protocol.AllocateRequest")
	proto.RegisterType((*AllocateResponse)(nil), "protocol.AllocateResponse")
	proto.RegisterType((*FreeRequest)(nil), "protocol.FreeRequest")
	proto.RegisterType((*FreeResponse)(nil), "protocol.FreeResponse")
	proto.RegisterType((*RegisterRequest)(nil), "protocol.RegisterRequest")
	proto.RegisterType((*RegisterResponse)(nil), "protocol.RegisterResponse")
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// Client API for ShadowsocksManagerSlave service

type ShadowsocksManagerSlaveClient interface {
	Allocate(ctx context.Context, in *AllocateRequest, opts ...grpc.CallOption) (*AllocateResponse, error)
	Free(ctx context.Context, in *FreeRequest, opts ...grpc.CallOption) (*FreeResponse, error)
	ListServices(ctx context.Context, in *google_protobuf.Empty, opts ...grpc.CallOption) (*ServiceList, error)
	GetStats(ctx context.Context, in *google_protobuf.Empty, opts ...grpc.CallOption) (*Statistics, error)
	GetStatsStream(ctx context.Context, in *google_protobuf.Empty, opts ...grpc.CallOption) (ShadowsocksManagerSlave_GetStatsStreamClient, error)
	SetStat(ctx context.Context, in *Statistics, opts ...grpc.CallOption) (*google_protobuf.Empty, error)
}

type shadowsocksManagerSlaveClient struct {
	cc *grpc.ClientConn
}

func NewShadowsocksManagerSlaveClient(cc *grpc.ClientConn) ShadowsocksManagerSlaveClient {
	return &shadowsocksManagerSlaveClient{cc}
}

func (c *shadowsocksManagerSlaveClient) Allocate(ctx context.Context, in *AllocateRequest, opts ...grpc.CallOption) (*AllocateResponse, error) {
	out := new(AllocateResponse)
	err := grpc.Invoke(ctx, "/protocol.ShadowsocksManagerSlave/Allocate", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *shadowsocksManagerSlaveClient) Free(ctx context.Context, in *FreeRequest, opts ...grpc.CallOption) (*FreeResponse, error) {
	out := new(FreeResponse)
	err := grpc.Invoke(ctx, "/protocol.ShadowsocksManagerSlave/Free", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *shadowsocksManagerSlaveClient) ListServices(ctx context.Context, in *google_protobuf.Empty, opts ...grpc.CallOption) (*ServiceList, error) {
	out := new(ServiceList)
	err := grpc.Invoke(ctx, "/protocol.ShadowsocksManagerSlave/ListServices", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *shadowsocksManagerSlaveClient) GetStats(ctx context.Context, in *google_protobuf.Empty, opts ...grpc.CallOption) (*Statistics, error) {
	out := new(Statistics)
	err := grpc.Invoke(ctx, "/protocol.ShadowsocksManagerSlave/GetStats", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *shadowsocksManagerSlaveClient) GetStatsStream(ctx context.Context, in *google_protobuf.Empty, opts ...grpc.CallOption) (ShadowsocksManagerSlave_GetStatsStreamClient, error) {
	stream, err := grpc.NewClientStream(ctx, &_ShadowsocksManagerSlave_serviceDesc.Streams[0], c.cc, "/protocol.ShadowsocksManagerSlave/GetStatsStream", opts...)
	if err != nil {
		return nil, err
	}
	x := &shadowsocksManagerSlaveGetStatsStreamClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type ShadowsocksManagerSlave_GetStatsStreamClient interface {
	Recv() (*Statistics, error)
	grpc.ClientStream
}

type shadowsocksManagerSlaveGetStatsStreamClient struct {
	grpc.ClientStream
}

func (x *shadowsocksManagerSlaveGetStatsStreamClient) Recv() (*Statistics, error) {
	m := new(Statistics)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *shadowsocksManagerSlaveClient) SetStat(ctx context.Context, in *Statistics, opts ...grpc.CallOption) (*google_protobuf.Empty, error) {
	out := new(google_protobuf.Empty)
	err := grpc.Invoke(ctx, "/protocol.ShadowsocksManagerSlave/SetStat", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Server API for ShadowsocksManagerSlave service

type ShadowsocksManagerSlaveServer interface {
	Allocate(context.Context, *AllocateRequest) (*AllocateResponse, error)
	Free(context.Context, *FreeRequest) (*FreeResponse, error)
	ListServices(context.Context, *google_protobuf.Empty) (*ServiceList, error)
	GetStats(context.Context, *google_protobuf.Empty) (*Statistics, error)
	GetStatsStream(*google_protobuf.Empty, ShadowsocksManagerSlave_GetStatsStreamServer) error
	SetStat(context.Context, *Statistics) (*google_protobuf.Empty, error)
}

func RegisterShadowsocksManagerSlaveServer(s *grpc.Server, srv ShadowsocksManagerSlaveServer) {
	s.RegisterService(&_ShadowsocksManagerSlave_serviceDesc, srv)
}

func _ShadowsocksManagerSlave_Allocate_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(AllocateRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ShadowsocksManagerSlaveServer).Allocate(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/protocol.ShadowsocksManagerSlave/Allocate",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ShadowsocksManagerSlaveServer).Allocate(ctx, req.(*AllocateRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ShadowsocksManagerSlave_Free_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(FreeRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ShadowsocksManagerSlaveServer).Free(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/protocol.ShadowsocksManagerSlave/Free",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ShadowsocksManagerSlaveServer).Free(ctx, req.(*FreeRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ShadowsocksManagerSlave_ListServices_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(google_protobuf.Empty)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ShadowsocksManagerSlaveServer).ListServices(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/protocol.ShadowsocksManagerSlave/ListServices",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ShadowsocksManagerSlaveServer).ListServices(ctx, req.(*google_protobuf.Empty))
	}
	return interceptor(ctx, in, info, handler)
}

func _ShadowsocksManagerSlave_GetStats_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(google_protobuf.Empty)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ShadowsocksManagerSlaveServer).GetStats(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/protocol.ShadowsocksManagerSlave/GetStats",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ShadowsocksManagerSlaveServer).GetStats(ctx, req.(*google_protobuf.Empty))
	}
	return interceptor(ctx, in, info, handler)
}

func _ShadowsocksManagerSlave_GetStatsStream_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(google_protobuf.Empty)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(ShadowsocksManagerSlaveServer).GetStatsStream(m, &shadowsocksManagerSlaveGetStatsStreamServer{stream})
}

type ShadowsocksManagerSlave_GetStatsStreamServer interface {
	Send(*Statistics) error
	grpc.ServerStream
}

type shadowsocksManagerSlaveGetStatsStreamServer struct {
	grpc.ServerStream
}

func (x *shadowsocksManagerSlaveGetStatsStreamServer) Send(m *Statistics) error {
	return x.ServerStream.SendMsg(m)
}

func _ShadowsocksManagerSlave_SetStat_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Statistics)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ShadowsocksManagerSlaveServer).SetStat(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/protocol.ShadowsocksManagerSlave/SetStat",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ShadowsocksManagerSlaveServer).SetStat(ctx, req.(*Statistics))
	}
	return interceptor(ctx, in, info, handler)
}

var _ShadowsocksManagerSlave_serviceDesc = grpc.ServiceDesc{
	ServiceName: "protocol.ShadowsocksManagerSlave",
	HandlerType: (*ShadowsocksManagerSlaveServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Allocate",
			Handler:    _ShadowsocksManagerSlave_Allocate_Handler,
		},
		{
			MethodName: "Free",
			Handler:    _ShadowsocksManagerSlave_Free_Handler,
		},
		{
			MethodName: "ListServices",
			Handler:    _ShadowsocksManagerSlave_ListServices_Handler,
		},
		{
			MethodName: "GetStats",
			Handler:    _ShadowsocksManagerSlave_GetStats_Handler,
		},
		{
			MethodName: "SetStat",
			Handler:    _ShadowsocksManagerSlave_SetStat_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "GetStatsStream",
			Handler:       _ShadowsocksManagerSlave_GetStatsStream_Handler,
			ServerStreams: true,
		},
	},
	Metadata: "shadowsocks_manager.proto",
}

// Client API for ShadowsocksManagerMaster service

type ShadowsocksManagerMasterClient interface {
	// Register registes self on the master node.
	Register(ctx context.Context, in *RegisterRequest, opts ...grpc.CallOption) (*RegisterResponse, error)
}

type shadowsocksManagerMasterClient struct {
	cc *grpc.ClientConn
}

func NewShadowsocksManagerMasterClient(cc *grpc.ClientConn) ShadowsocksManagerMasterClient {
	return &shadowsocksManagerMasterClient{cc}
}

func (c *shadowsocksManagerMasterClient) Register(ctx context.Context, in *RegisterRequest, opts ...grpc.CallOption) (*RegisterResponse, error) {
	out := new(RegisterResponse)
	err := grpc.Invoke(ctx, "/protocol.ShadowsocksManagerMaster/Register", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Server API for ShadowsocksManagerMaster service

type ShadowsocksManagerMasterServer interface {
	// Register registes self on the master node.
	Register(context.Context, *RegisterRequest) (*RegisterResponse, error)
}

func RegisterShadowsocksManagerMasterServer(s *grpc.Server, srv ShadowsocksManagerMasterServer) {
	s.RegisterService(&_ShadowsocksManagerMaster_serviceDesc, srv)
}

func _ShadowsocksManagerMaster_Register_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(RegisterRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ShadowsocksManagerMasterServer).Register(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/protocol.ShadowsocksManagerMaster/Register",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ShadowsocksManagerMasterServer).Register(ctx, req.(*RegisterRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _ShadowsocksManagerMaster_serviceDesc = grpc.ServiceDesc{
	ServiceName: "protocol.ShadowsocksManagerMaster",
	HandlerType: (*ShadowsocksManagerMasterServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Register",
			Handler:    _ShadowsocksManagerMaster_Register_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "shadowsocks_manager.proto",
}

func init() { proto.RegisterFile("shadowsocks_manager.proto", fileDescriptor0) }

var fileDescriptor0 = []byte{
	// 518 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x09, 0x6e, 0x88, 0x02, 0xff, 0xac, 0x53, 0xcf, 0x6f, 0xd3, 0x30,
	0x14, 0x6e, 0xd6, 0x75, 0xcd, 0x5e, 0x0b, 0xab, 0xac, 0xb2, 0xa5, 0x81, 0x43, 0xe5, 0x53, 0x0f,
	0x28, 0x43, 0xe5, 0xc0, 0x18, 0x12, 0x08, 0xd0, 0x28, 0x88, 0xed, 0x92, 0xc0, 0x11, 0x45, 0x5e,
	0xea, 0x96, 0xa8, 0x6e, 0x1d, 0x6c, 0xa7, 0x53, 0xff, 0x04, 0xce, 0xfc, 0xc3, 0xc8, 0x71, 0x7e,
	0xd1, 0x75, 0x07, 0xb4, 0x9d, 0x92, 0x17, 0x7f, 0xef, 0xcb, 0xf7, 0xde, 0xf7, 0x19, 0x06, 0xf2,
	0x27, 0x99, 0xf2, 0x1b, 0xc9, 0xa3, 0x85, 0x0c, 0x97, 0x64, 0x45, 0xe6, 0x54, 0x78, 0x89, 0xe0,
	0x8a, 0x23, 0x3b, 0x7b, 0x44, 0x9c, 0xb9, 0x4f, 0xe7, 0x9c, 0xcf, 0x19, 0x3d, 0xcd, 0x3e, 0x5c,
	0xa7, 0xb3, 0x53, 0xba, 0x4c, 0xd4, 0xc6, 0xc0, 0xf0, 0x0f, 0x40, 0x41, 0xc5, 0x11, 0x50, 0xb1,
	0x8e, 0x23, 0x8a, 0x4e, 0xa0, 0x9d, 0x4a, 0x2a, 0xc2, 0x78, 0xea, 0x44, 0x43, 0x6b, 0x74, 0xe8,
	0x1f, 0xe8, 0xf2, 0xcb, 0x14, 0x21, 0xd8, 0x4f, 0xb8, 0x50, 0x8e, 0x35, 0xb4, 0x46, 0x2d, 0x3f,
	0x7b, 0x47, 0x2e, 0xd8, 0x09, 0x91, 0xf2, 0x86, 0x8b, 0xa9, 0xb3, 0x97, 0xa1, 0xcb, 0x1a, 0x4f,
	0xa0, 0x93, 0x73, 0x5e, 0xc6, 0x52, 0xa1, 0x33, 0xb0, 0xa5, 0x29, 0xa5, 0x63, 0x0d, 0x9b, 0xa3,
	0xce, 0xf8, 0x99, 0x57, 0xe8, 0xf4, 0x6e, 0xeb, 0xf0, 0x4b, 0x34, 0xfe, 0x6d, 0x01, 0x04, 0x8a,
	0xa8, 0x58, 0xaa, 0x38, 0x92, 0xe8, 0x2d, 0xd8, 0x4a, 0x90, 0xd9, 0x2c, 0x8e, 0x0a, 0x22, 0x5c,
	0x23, 0x2a, 0x71, 0xde, 0xb7, 0x1c, 0x74, 0xb1, 0x52, 0x62, 0xe3, 0x97, 0x3d, 0xee, 0x1b, 0x78,
	0xf4, 0xcf, 0x11, 0xea, 0x41, 0x73, 0x41, 0x37, 0xf9, 0x5c, 0xfa, 0x15, 0xf5, 0xa1, 0xb5, 0x26,
	0x2c, 0xa5, 0xd9, 0x4c, 0x4d, 0xdf, 0x14, 0xe7, 0x7b, 0x67, 0x16, 0xfe, 0x0a, 0x47, 0xef, 0x19,
	0xe3, 0x11, 0x51, 0xd4, 0xa7, 0xbf, 0x52, 0x9a, 0x0d, 0xd6, 0xcd, 0xa5, 0x86, 0x2c, 0x96, 0x66,
	0x3f, 0x9d, 0xf1, 0x93, 0x9a, 0xa6, 0x6a, 0x0b, 0x7e, 0x47, 0x56, 0x05, 0xbe, 0x84, 0x5e, 0x45,
	0x26, 0x13, 0xbe, 0x92, 0xf4, 0x1e, 0x6c, 0x13, 0xe8, 0x7c, 0x12, 0xf4, 0x01, 0x64, 0x7d, 0x86,
	0xae, 0x21, 0xba, 0xb7, 0xa4, 0xe7, 0x70, 0xe4, 0xd3, 0x79, 0x2c, 0x15, 0x15, 0x85, 0xac, 0x81,
	0x8e, 0x01, 0x9b, 0x85, 0xa9, 0x60, 0x19, 0xd1, 0xa1, 0xdf, 0xd6, 0xf5, 0x77, 0xc1, 0xf0, 0x08,
	0x7a, 0x15, 0x3a, 0xff, 0x77, 0x1f, 0x5a, 0x8a, 0x2f, 0xe8, 0x2a, 0xc7, 0x9a, 0x62, 0xfc, 0xa7,
	0x09, 0x27, 0xb5, 0xc8, 0x5c, 0x99, 0xf4, 0x07, 0x8c, 0xac, 0x29, 0xfa, 0x08, 0x76, 0xb1, 0x54,
	0x34, 0xa8, 0x34, 0x6e, 0xb9, 0xe6, 0xba, 0xbb, 0x8e, 0xcc, 0x4f, 0x71, 0x03, 0xbd, 0x82, 0x7d,
	0xbd, 0x02, 0x54, 0x1b, 0xb2, 0xb6, 0x5b, 0xf7, 0x78, 0xfb, 0x73, 0xd9, 0xf8, 0x0e, 0xba, 0x7a,
	0xf2, 0x7c, 0x23, 0x12, 0x1d, 0x7b, 0xe6, 0x06, 0x7a, 0xc5, 0x0d, 0xf4, 0x2e, 0xf4, 0x0d, 0x74,
	0x77, 0x6f, 0x0f, 0x37, 0xd0, 0x39, 0xd8, 0x13, 0xaa, 0x74, 0x8c, 0xef, 0x6e, 0xee, 0xef, 0xca,
	0x3b, 0x6e, 0xa0, 0x0f, 0xf0, 0xb8, 0xe8, 0x0d, 0x94, 0xa0, 0x64, 0xf9, 0xbf, 0x0c, 0x2f, 0x2c,
	0xf4, 0x1a, 0xda, 0x81, 0xe1, 0x40, 0x3b, 0x41, 0xee, 0x1d, 0x94, 0xb8, 0x31, 0x0e, 0xc1, 0xb9,
	0x6d, 0xca, 0x15, 0xd1, 0x7e, 0x6a, 0x57, 0x0a, 0x6f, 0xeb, 0xae, 0x6c, 0xa5, 0xa3, 0xee, 0xca,
	0x76, 0x14, 0x70, 0xe3, 0xfa, 0x20, 0x3b, 0x7c, 0xf9, 0x37, 0x00, 0x00, 0xff, 0xff, 0x22, 0x22,
	0xc1, 0x4c, 0xfb, 0x04, 0x00, 0x00,
}
