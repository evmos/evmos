// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: evmos/claim/v1/query.proto

package types

import (
	context "context"
	fmt "fmt"
	github_com_cosmos_cosmos_sdk_types "github.com/cosmos/cosmos-sdk/types"
	types "github.com/cosmos/cosmos-sdk/types"
	_ "github.com/gogo/protobuf/gogoproto"
	grpc1 "github.com/gogo/protobuf/grpc"
	proto "github.com/gogo/protobuf/proto"
	_ "google.golang.org/genproto/googleapis/api/annotations"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
	io "io"
	math "math"
	math_bits "math/bits"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.GoGoProtoPackageIsVersion3 // please upgrade the proto package

// QueryTotalUnclaimedRequest is the request type for the Query/TotalUnclaimed RPC method.
type QueryTotalUnclaimedRequest struct {
}

func (m *QueryTotalUnclaimedRequest) Reset()         { *m = QueryTotalUnclaimedRequest{} }
func (m *QueryTotalUnclaimedRequest) String() string { return proto.CompactTextString(m) }
func (*QueryTotalUnclaimedRequest) ProtoMessage()    {}
func (*QueryTotalUnclaimedRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_75c94980d888f50c, []int{0}
}
func (m *QueryTotalUnclaimedRequest) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *QueryTotalUnclaimedRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_QueryTotalUnclaimedRequest.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *QueryTotalUnclaimedRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_QueryTotalUnclaimedRequest.Merge(m, src)
}
func (m *QueryTotalUnclaimedRequest) XXX_Size() int {
	return m.Size()
}
func (m *QueryTotalUnclaimedRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_QueryTotalUnclaimedRequest.DiscardUnknown(m)
}

var xxx_messageInfo_QueryTotalUnclaimedRequest proto.InternalMessageInfo

// QueryTotalUnclaimedResponse is the response type for the Query/TotalUnclaimed RPC method.
type QueryTotalUnclaimedResponse struct {
	// coins define the unclaimed coins
	Coins github_com_cosmos_cosmos_sdk_types.Coins `protobuf:"bytes,1,rep,name=coins,proto3,castrepeated=github.com/cosmos/cosmos-sdk/types.Coins" json:"coins"`
}

func (m *QueryTotalUnclaimedResponse) Reset()         { *m = QueryTotalUnclaimedResponse{} }
func (m *QueryTotalUnclaimedResponse) String() string { return proto.CompactTextString(m) }
func (*QueryTotalUnclaimedResponse) ProtoMessage()    {}
func (*QueryTotalUnclaimedResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_75c94980d888f50c, []int{1}
}
func (m *QueryTotalUnclaimedResponse) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *QueryTotalUnclaimedResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_QueryTotalUnclaimedResponse.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *QueryTotalUnclaimedResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_QueryTotalUnclaimedResponse.Merge(m, src)
}
func (m *QueryTotalUnclaimedResponse) XXX_Size() int {
	return m.Size()
}
func (m *QueryTotalUnclaimedResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_QueryTotalUnclaimedResponse.DiscardUnknown(m)
}

var xxx_messageInfo_QueryTotalUnclaimedResponse proto.InternalMessageInfo

func (m *QueryTotalUnclaimedResponse) GetCoins() github_com_cosmos_cosmos_sdk_types.Coins {
	if m != nil {
		return m.Coins
	}
	return nil
}

// QueryParamsRequest is the request type for the Query/Params RPC method.
type QueryParamsRequest struct {
}

func (m *QueryParamsRequest) Reset()         { *m = QueryParamsRequest{} }
func (m *QueryParamsRequest) String() string { return proto.CompactTextString(m) }
func (*QueryParamsRequest) ProtoMessage()    {}
func (*QueryParamsRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_75c94980d888f50c, []int{2}
}
func (m *QueryParamsRequest) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *QueryParamsRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_QueryParamsRequest.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *QueryParamsRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_QueryParamsRequest.Merge(m, src)
}
func (m *QueryParamsRequest) XXX_Size() int {
	return m.Size()
}
func (m *QueryParamsRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_QueryParamsRequest.DiscardUnknown(m)
}

var xxx_messageInfo_QueryParamsRequest proto.InternalMessageInfo

// QueryParamsResponse is the response type for the Query/Params RPC method.
type QueryParamsResponse struct {
	// params defines the parameters of the module.
	Params Params `protobuf:"bytes,1,opt,name=params,proto3" json:"params"`
}

func (m *QueryParamsResponse) Reset()         { *m = QueryParamsResponse{} }
func (m *QueryParamsResponse) String() string { return proto.CompactTextString(m) }
func (*QueryParamsResponse) ProtoMessage()    {}
func (*QueryParamsResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_75c94980d888f50c, []int{3}
}
func (m *QueryParamsResponse) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *QueryParamsResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_QueryParamsResponse.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *QueryParamsResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_QueryParamsResponse.Merge(m, src)
}
func (m *QueryParamsResponse) XXX_Size() int {
	return m.Size()
}
func (m *QueryParamsResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_QueryParamsResponse.DiscardUnknown(m)
}

var xxx_messageInfo_QueryParamsResponse proto.InternalMessageInfo

func (m *QueryParamsResponse) GetParams() Params {
	if m != nil {
		return m.Params
	}
	return Params{}
}

// QueryClaimRecordsRequest is the request type for the Query/ClaimRecords RPC method.
type QueryClaimRecordsRequest struct {
	Address string `protobuf:"bytes,1,opt,name=address,proto3" json:"address,omitempty"`
}

func (m *QueryClaimRecordsRequest) Reset()         { *m = QueryClaimRecordsRequest{} }
func (m *QueryClaimRecordsRequest) String() string { return proto.CompactTextString(m) }
func (*QueryClaimRecordsRequest) ProtoMessage()    {}
func (*QueryClaimRecordsRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_75c94980d888f50c, []int{4}
}
func (m *QueryClaimRecordsRequest) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *QueryClaimRecordsRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_QueryClaimRecordsRequest.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *QueryClaimRecordsRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_QueryClaimRecordsRequest.Merge(m, src)
}
func (m *QueryClaimRecordsRequest) XXX_Size() int {
	return m.Size()
}
func (m *QueryClaimRecordsRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_QueryClaimRecordsRequest.DiscardUnknown(m)
}

var xxx_messageInfo_QueryClaimRecordsRequest proto.InternalMessageInfo

func (m *QueryClaimRecordsRequest) GetAddress() string {
	if m != nil {
		return m.Address
	}
	return ""
}

// QueryClaimRecordsResponse is the response type for the Query/ClaimRecords RPC method.
type QueryClaimRecordsResponse struct {
	// total initial claimable amount for the user
	InitialClaimableAmount github_com_cosmos_cosmos_sdk_types.Int `protobuf:"bytes,1,opt,name=initial_claimable_amount,json=initialClaimableAmount,proto3,customtype=github.com/cosmos/cosmos-sdk/types.Int" json:"initial_claimable_amount"`
	Claims                 []Claim                                `protobuf:"bytes,2,rep,name=claims,proto3" json:"claims"`
}

func (m *QueryClaimRecordsResponse) Reset()         { *m = QueryClaimRecordsResponse{} }
func (m *QueryClaimRecordsResponse) String() string { return proto.CompactTextString(m) }
func (*QueryClaimRecordsResponse) ProtoMessage()    {}
func (*QueryClaimRecordsResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_75c94980d888f50c, []int{5}
}
func (m *QueryClaimRecordsResponse) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *QueryClaimRecordsResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_QueryClaimRecordsResponse.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *QueryClaimRecordsResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_QueryClaimRecordsResponse.Merge(m, src)
}
func (m *QueryClaimRecordsResponse) XXX_Size() int {
	return m.Size()
}
func (m *QueryClaimRecordsResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_QueryClaimRecordsResponse.DiscardUnknown(m)
}

var xxx_messageInfo_QueryClaimRecordsResponse proto.InternalMessageInfo

func (m *QueryClaimRecordsResponse) GetClaims() []Claim {
	if m != nil {
		return m.Claims
	}
	return nil
}

func init() {
	proto.RegisterType((*QueryTotalUnclaimedRequest)(nil), "evmos.claim.v1.QueryTotalUnclaimedRequest")
	proto.RegisterType((*QueryTotalUnclaimedResponse)(nil), "evmos.claim.v1.QueryTotalUnclaimedResponse")
	proto.RegisterType((*QueryParamsRequest)(nil), "evmos.claim.v1.QueryParamsRequest")
	proto.RegisterType((*QueryParamsResponse)(nil), "evmos.claim.v1.QueryParamsResponse")
	proto.RegisterType((*QueryClaimRecordsRequest)(nil), "evmos.claim.v1.QueryClaimRecordsRequest")
	proto.RegisterType((*QueryClaimRecordsResponse)(nil), "evmos.claim.v1.QueryClaimRecordsResponse")
}

func init() { proto.RegisterFile("evmos/claim/v1/query.proto", fileDescriptor_75c94980d888f50c) }

var fileDescriptor_75c94980d888f50c = []byte{
	// 556 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x8c, 0x93, 0xc1, 0x6e, 0xd3, 0x30,
	0x1c, 0xc6, 0xeb, 0x8d, 0x15, 0xe1, 0xa1, 0x1d, 0xcc, 0x98, 0xb2, 0x30, 0xa5, 0x23, 0x48, 0xac,
	0x03, 0x61, 0xd3, 0x6e, 0x0f, 0x00, 0xed, 0x09, 0x71, 0x81, 0x08, 0x2e, 0x5c, 0x2a, 0xb7, 0xb5,
	0x52, 0x8b, 0xd6, 0x4e, 0x63, 0xa7, 0x62, 0x42, 0x48, 0x88, 0x27, 0x40, 0xc0, 0x33, 0x70, 0xe0,
	0x11, 0x78, 0x82, 0x1d, 0x27, 0x71, 0x41, 0x1c, 0x06, 0x6a, 0x79, 0x10, 0x14, 0xdb, 0x41, 0x6b,
	0x14, 0x50, 0x4f, 0x6d, 0xf2, 0xff, 0xfc, 0x7d, 0x3f, 0xdb, 0x5f, 0xa0, 0xcf, 0x66, 0x13, 0xa9,
	0xc8, 0x60, 0x4c, 0xf9, 0x84, 0xcc, 0x5a, 0x64, 0x9a, 0xb1, 0xf4, 0x04, 0x27, 0xa9, 0xd4, 0x12,
	0x6d, 0x99, 0x19, 0x36, 0x33, 0x3c, 0x6b, 0xf9, 0xdb, 0xb1, 0x8c, 0xa5, 0x19, 0x91, 0xfc, 0x9f,
	0x55, 0xf9, 0x7b, 0xb1, 0x94, 0xf1, 0x98, 0x11, 0x9a, 0x70, 0x42, 0x85, 0x90, 0x9a, 0x6a, 0x2e,
	0x85, 0x72, 0xd3, 0x60, 0x20, 0x55, 0x1e, 0xd0, 0xa7, 0x8a, 0x91, 0x59, 0xab, 0xcf, 0x34, 0x6d,
	0x91, 0x81, 0xe4, 0xc2, 0xcd, 0xcb, 0xf9, 0x36, 0xcc, 0x39, 0x97, 0x66, 0x31, 0x13, 0x4c, 0x71,
	0xe7, 0x1c, 0xee, 0x41, 0xff, 0x69, 0x0e, 0xfb, 0x4c, 0x6a, 0x3a, 0x7e, 0x2e, 0x8c, 0x8a, 0x0d,
	0x23, 0x36, 0xcd, 0x98, 0xd2, 0xe1, 0x5b, 0x00, 0x6f, 0x54, 0x8e, 0x55, 0x22, 0x85, 0x62, 0x88,
	0xc2, 0x8d, 0x9c, 0x42, 0x79, 0x60, 0x7f, 0xbd, 0xb9, 0xd9, 0xde, 0xc5, 0x96, 0x13, 0xe7, 0x9c,
	0xd8, 0x71, 0xe2, 0xae, 0xe4, 0xa2, 0x73, 0xff, 0xf4, 0xbc, 0x51, 0xfb, 0xf2, 0xb3, 0xd1, 0x8c,
	0xb9, 0x1e, 0x65, 0x7d, 0x3c, 0x90, 0x13, 0xe2, 0x36, 0x65, 0x7f, 0xee, 0xa9, 0xe1, 0x4b, 0xa2,
	0x4f, 0x12, 0xa6, 0xcc, 0x02, 0x15, 0x59, 0xe7, 0x70, 0x1b, 0x22, 0x43, 0xf0, 0x84, 0xa6, 0x74,
	0xa2, 0x0a, 0xb0, 0xc7, 0xf0, 0xda, 0xd2, 0x5b, 0xc7, 0x73, 0x0c, 0xeb, 0x89, 0x79, 0xe3, 0x81,
	0x7d, 0xd0, 0xdc, 0x6c, 0xef, 0xe0, 0xe5, 0xc3, 0xc7, 0x56, 0xdf, 0xb9, 0x94, 0xd3, 0x44, 0x4e,
	0x1b, 0x1e, 0x43, 0xcf, 0x98, 0x75, 0x73, 0x55, 0xc4, 0x06, 0x32, 0x1d, 0x16, 0x41, 0xc8, 0x83,
	0x97, 0xe9, 0x70, 0x98, 0x32, 0x65, 0x2d, 0xaf, 0x44, 0xc5, 0x63, 0xf8, 0x15, 0xc0, 0xdd, 0x8a,
	0x65, 0x8e, 0x64, 0x04, 0x3d, 0x2e, 0xb8, 0xe6, 0x74, 0xdc, 0x33, 0xe1, 0xb4, 0x3f, 0x66, 0x3d,
	0x3a, 0x91, 0x99, 0xd0, 0xd6, 0xa8, 0x83, 0x73, 0x86, 0x1f, 0xe7, 0x8d, 0xdb, 0x2b, 0x9c, 0xc8,
	0x23, 0xa1, 0xa3, 0x1d, 0xe7, 0xd7, 0x2d, 0xec, 0x1e, 0x1a, 0x37, 0x74, 0x04, 0xeb, 0x26, 0x41,
	0x79, 0x6b, 0xe6, 0x12, 0xae, 0x97, 0xf7, 0x6c, 0x16, 0x14, 0x5b, 0xb6, 0xd2, 0xf6, 0xe7, 0x75,
	0xb8, 0x61, 0xe0, 0xd1, 0x07, 0x00, 0xb7, 0x96, 0x6f, 0x17, 0xdd, 0x29, 0x3b, 0xfc, 0xbb, 0x21,
	0xfe, 0xdd, 0x95, 0xb4, 0xf6, 0x50, 0xc2, 0x83, 0x77, 0xdf, 0x7e, 0x7f, 0x5c, 0xbb, 0x89, 0x1a,
	0xa4, 0xd4, 0x49, 0x9d, 0xeb, 0x7b, 0xd9, 0x5f, 0x82, 0x29, 0xac, 0xdb, 0x9b, 0x42, 0x61, 0xa5,
	0xff, 0x52, 0x19, 0xfc, 0x5b, 0xff, 0xd5, 0xb8, 0xec, 0xc0, 0x64, 0x7b, 0x68, 0xa7, 0x9c, 0x6d,
	0x4b, 0x80, 0x3e, 0x01, 0x78, 0xf5, 0xe2, 0x4d, 0xa2, 0x66, 0xa5, 0x6b, 0x45, 0x47, 0xfc, 0xc3,
	0x15, 0x94, 0x8e, 0x82, 0x18, 0x8a, 0x43, 0x74, 0x40, 0xaa, 0xbe, 0xd8, 0x5e, 0x6a, 0xe5, 0xe4,
	0xb5, 0x2b, 0xd9, 0x9b, 0xce, 0x83, 0xd3, 0x79, 0x00, 0xce, 0xe6, 0x01, 0xf8, 0x35, 0x0f, 0xc0,
	0xfb, 0x45, 0x50, 0x3b, 0x5b, 0x04, 0xb5, 0xef, 0x8b, 0xa0, 0xf6, 0xe2, 0x62, 0x6f, 0xf4, 0x88,
	0xa6, 0x8a, 0x2b, 0x67, 0xfa, 0xca, 0xd9, 0x9a, 0xee, 0xf4, 0xeb, 0xe6, 0x43, 0x3f, 0xfa, 0x13,
	0x00, 0x00, 0xff, 0xff, 0x64, 0x6b, 0x49, 0x06, 0xa4, 0x04, 0x00, 0x00,
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// QueryClient is the client API for Query service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type QueryClient interface {
	// TotalUnclaimed queries the total unclaimed tokens from the airdrop
	TotalUnclaimed(ctx context.Context, in *QueryTotalUnclaimedRequest, opts ...grpc.CallOption) (*QueryTotalUnclaimedResponse, error)
	// Params returns the claim module parameters
	Params(ctx context.Context, in *QueryParamsRequest, opts ...grpc.CallOption) (*QueryParamsResponse, error)
	// ClaimRecords returns the claims records for a given address
	ClaimRecords(ctx context.Context, in *QueryClaimRecordsRequest, opts ...grpc.CallOption) (*QueryClaimRecordsResponse, error)
}

type queryClient struct {
	cc grpc1.ClientConn
}

func NewQueryClient(cc grpc1.ClientConn) QueryClient {
	return &queryClient{cc}
}

func (c *queryClient) TotalUnclaimed(ctx context.Context, in *QueryTotalUnclaimedRequest, opts ...grpc.CallOption) (*QueryTotalUnclaimedResponse, error) {
	out := new(QueryTotalUnclaimedResponse)
	err := c.cc.Invoke(ctx, "/evmos.claim.v1.Query/TotalUnclaimed", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *queryClient) Params(ctx context.Context, in *QueryParamsRequest, opts ...grpc.CallOption) (*QueryParamsResponse, error) {
	out := new(QueryParamsResponse)
	err := c.cc.Invoke(ctx, "/evmos.claim.v1.Query/Params", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *queryClient) ClaimRecords(ctx context.Context, in *QueryClaimRecordsRequest, opts ...grpc.CallOption) (*QueryClaimRecordsResponse, error) {
	out := new(QueryClaimRecordsResponse)
	err := c.cc.Invoke(ctx, "/evmos.claim.v1.Query/ClaimRecords", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// QueryServer is the server API for Query service.
type QueryServer interface {
	// TotalUnclaimed queries the total unclaimed tokens from the airdrop
	TotalUnclaimed(context.Context, *QueryTotalUnclaimedRequest) (*QueryTotalUnclaimedResponse, error)
	// Params returns the claim module parameters
	Params(context.Context, *QueryParamsRequest) (*QueryParamsResponse, error)
	// ClaimRecords returns the claims records for a given address
	ClaimRecords(context.Context, *QueryClaimRecordsRequest) (*QueryClaimRecordsResponse, error)
}

// UnimplementedQueryServer can be embedded to have forward compatible implementations.
type UnimplementedQueryServer struct {
}

func (*UnimplementedQueryServer) TotalUnclaimed(ctx context.Context, req *QueryTotalUnclaimedRequest) (*QueryTotalUnclaimedResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method TotalUnclaimed not implemented")
}
func (*UnimplementedQueryServer) Params(ctx context.Context, req *QueryParamsRequest) (*QueryParamsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Params not implemented")
}
func (*UnimplementedQueryServer) ClaimRecords(ctx context.Context, req *QueryClaimRecordsRequest) (*QueryClaimRecordsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ClaimRecords not implemented")
}

func RegisterQueryServer(s grpc1.Server, srv QueryServer) {
	s.RegisterService(&_Query_serviceDesc, srv)
}

func _Query_TotalUnclaimed_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(QueryTotalUnclaimedRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(QueryServer).TotalUnclaimed(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/evmos.claim.v1.Query/TotalUnclaimed",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(QueryServer).TotalUnclaimed(ctx, req.(*QueryTotalUnclaimedRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Query_Params_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(QueryParamsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(QueryServer).Params(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/evmos.claim.v1.Query/Params",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(QueryServer).Params(ctx, req.(*QueryParamsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Query_ClaimRecords_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(QueryClaimRecordsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(QueryServer).ClaimRecords(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/evmos.claim.v1.Query/ClaimRecords",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(QueryServer).ClaimRecords(ctx, req.(*QueryClaimRecordsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _Query_serviceDesc = grpc.ServiceDesc{
	ServiceName: "evmos.claim.v1.Query",
	HandlerType: (*QueryServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "TotalUnclaimed",
			Handler:    _Query_TotalUnclaimed_Handler,
		},
		{
			MethodName: "Params",
			Handler:    _Query_Params_Handler,
		},
		{
			MethodName: "ClaimRecords",
			Handler:    _Query_ClaimRecords_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "evmos/claim/v1/query.proto",
}

func (m *QueryTotalUnclaimedRequest) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *QueryTotalUnclaimedRequest) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *QueryTotalUnclaimedRequest) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	return len(dAtA) - i, nil
}

func (m *QueryTotalUnclaimedResponse) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *QueryTotalUnclaimedResponse) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *QueryTotalUnclaimedResponse) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.Coins) > 0 {
		for iNdEx := len(m.Coins) - 1; iNdEx >= 0; iNdEx-- {
			{
				size, err := m.Coins[iNdEx].MarshalToSizedBuffer(dAtA[:i])
				if err != nil {
					return 0, err
				}
				i -= size
				i = encodeVarintQuery(dAtA, i, uint64(size))
			}
			i--
			dAtA[i] = 0xa
		}
	}
	return len(dAtA) - i, nil
}

func (m *QueryParamsRequest) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *QueryParamsRequest) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *QueryParamsRequest) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	return len(dAtA) - i, nil
}

func (m *QueryParamsResponse) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *QueryParamsResponse) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *QueryParamsResponse) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	{
		size, err := m.Params.MarshalToSizedBuffer(dAtA[:i])
		if err != nil {
			return 0, err
		}
		i -= size
		i = encodeVarintQuery(dAtA, i, uint64(size))
	}
	i--
	dAtA[i] = 0xa
	return len(dAtA) - i, nil
}

func (m *QueryClaimRecordsRequest) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *QueryClaimRecordsRequest) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *QueryClaimRecordsRequest) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.Address) > 0 {
		i -= len(m.Address)
		copy(dAtA[i:], m.Address)
		i = encodeVarintQuery(dAtA, i, uint64(len(m.Address)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func (m *QueryClaimRecordsResponse) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *QueryClaimRecordsResponse) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *QueryClaimRecordsResponse) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.Claims) > 0 {
		for iNdEx := len(m.Claims) - 1; iNdEx >= 0; iNdEx-- {
			{
				size, err := m.Claims[iNdEx].MarshalToSizedBuffer(dAtA[:i])
				if err != nil {
					return 0, err
				}
				i -= size
				i = encodeVarintQuery(dAtA, i, uint64(size))
			}
			i--
			dAtA[i] = 0x12
		}
	}
	{
		size := m.InitialClaimableAmount.Size()
		i -= size
		if _, err := m.InitialClaimableAmount.MarshalTo(dAtA[i:]); err != nil {
			return 0, err
		}
		i = encodeVarintQuery(dAtA, i, uint64(size))
	}
	i--
	dAtA[i] = 0xa
	return len(dAtA) - i, nil
}

func encodeVarintQuery(dAtA []byte, offset int, v uint64) int {
	offset -= sovQuery(v)
	base := offset
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return base
}
func (m *QueryTotalUnclaimedRequest) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	return n
}

func (m *QueryTotalUnclaimedResponse) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if len(m.Coins) > 0 {
		for _, e := range m.Coins {
			l = e.Size()
			n += 1 + l + sovQuery(uint64(l))
		}
	}
	return n
}

func (m *QueryParamsRequest) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	return n
}

func (m *QueryParamsResponse) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = m.Params.Size()
	n += 1 + l + sovQuery(uint64(l))
	return n
}

func (m *QueryClaimRecordsRequest) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.Address)
	if l > 0 {
		n += 1 + l + sovQuery(uint64(l))
	}
	return n
}

func (m *QueryClaimRecordsResponse) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = m.InitialClaimableAmount.Size()
	n += 1 + l + sovQuery(uint64(l))
	if len(m.Claims) > 0 {
		for _, e := range m.Claims {
			l = e.Size()
			n += 1 + l + sovQuery(uint64(l))
		}
	}
	return n
}

func sovQuery(x uint64) (n int) {
	return (math_bits.Len64(x|1) + 6) / 7
}
func sozQuery(x uint64) (n int) {
	return sovQuery(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *QueryTotalUnclaimedRequest) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowQuery
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: QueryTotalUnclaimedRequest: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: QueryTotalUnclaimedRequest: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		default:
			iNdEx = preIndex
			skippy, err := skipQuery(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthQuery
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func (m *QueryTotalUnclaimedResponse) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowQuery
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: QueryTotalUnclaimedResponse: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: QueryTotalUnclaimedResponse: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Coins", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowQuery
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthQuery
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthQuery
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Coins = append(m.Coins, types.Coin{})
			if err := m.Coins[len(m.Coins)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipQuery(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthQuery
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func (m *QueryParamsRequest) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowQuery
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: QueryParamsRequest: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: QueryParamsRequest: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		default:
			iNdEx = preIndex
			skippy, err := skipQuery(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthQuery
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func (m *QueryParamsResponse) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowQuery
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: QueryParamsResponse: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: QueryParamsResponse: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Params", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowQuery
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthQuery
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthQuery
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := m.Params.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipQuery(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthQuery
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func (m *QueryClaimRecordsRequest) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowQuery
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: QueryClaimRecordsRequest: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: QueryClaimRecordsRequest: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Address", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowQuery
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthQuery
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthQuery
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Address = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipQuery(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthQuery
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func (m *QueryClaimRecordsResponse) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowQuery
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: QueryClaimRecordsResponse: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: QueryClaimRecordsResponse: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field InitialClaimableAmount", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowQuery
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthQuery
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthQuery
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := m.InitialClaimableAmount.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Claims", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowQuery
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthQuery
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthQuery
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Claims = append(m.Claims, Claim{})
			if err := m.Claims[len(m.Claims)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipQuery(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthQuery
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func skipQuery(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	depth := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowQuery
			}
			if iNdEx >= l {
				return 0, io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		wireType := int(wire & 0x7)
		switch wireType {
		case 0:
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return 0, ErrIntOverflowQuery
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				iNdEx++
				if dAtA[iNdEx-1] < 0x80 {
					break
				}
			}
		case 1:
			iNdEx += 8
		case 2:
			var length int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return 0, ErrIntOverflowQuery
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				length |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if length < 0 {
				return 0, ErrInvalidLengthQuery
			}
			iNdEx += length
		case 3:
			depth++
		case 4:
			if depth == 0 {
				return 0, ErrUnexpectedEndOfGroupQuery
			}
			depth--
		case 5:
			iNdEx += 4
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
		if iNdEx < 0 {
			return 0, ErrInvalidLengthQuery
		}
		if depth == 0 {
			return iNdEx, nil
		}
	}
	return 0, io.ErrUnexpectedEOF
}

var (
	ErrInvalidLengthQuery        = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowQuery          = fmt.Errorf("proto: integer overflow")
	ErrUnexpectedEndOfGroupQuery = fmt.Errorf("proto: unexpected end of group")
)
