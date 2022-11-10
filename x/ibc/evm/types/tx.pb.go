// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: evmos/ibc/evm/v1/tx.proto

package types

import (
	context "context"
	fmt "fmt"
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

// MsgCallEVM defines a Msg to execute an Etherum Tx from an ibc evm enabled
// source chain on an ibc evm enabled destination chain
type MsgCallEVM struct {
	// Amount of the given coin denomination to be sent
	// with the Tx
	Amount string `protobuf:"bytes,1,opt,name=amount,proto3" json:"amount,omitempty"`
	// Coin denomination for the EVM chain
	Denom string `protobuf:"bytes,2,opt,name=denom,proto3" json:"denom,omitempty"`
	// Packet contains the IBC EVM packet information
	Packet *IBCEVMPacketData `protobuf:"bytes,3,opt,name=packet,proto3" json:"packet,omitempty"`
}

func (m *MsgCallEVM) Reset()         { *m = MsgCallEVM{} }
func (m *MsgCallEVM) String() string { return proto.CompactTextString(m) }
func (*MsgCallEVM) ProtoMessage()    {}
func (*MsgCallEVM) Descriptor() ([]byte, []int) {
	return fileDescriptor_a40d98923f08b90e, []int{0}
}
func (m *MsgCallEVM) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *MsgCallEVM) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_MsgCallEVM.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *MsgCallEVM) XXX_Merge(src proto.Message) {
	xxx_messageInfo_MsgCallEVM.Merge(m, src)
}
func (m *MsgCallEVM) XXX_Size() int {
	return m.Size()
}
func (m *MsgCallEVM) XXX_DiscardUnknown() {
	xxx_messageInfo_MsgCallEVM.DiscardUnknown(m)
}

var xxx_messageInfo_MsgCallEVM proto.InternalMessageInfo

func (m *MsgCallEVM) GetAmount() string {
	if m != nil {
		return m.Amount
	}
	return ""
}

func (m *MsgCallEVM) GetDenom() string {
	if m != nil {
		return m.Denom
	}
	return ""
}

func (m *MsgCallEVM) GetPacket() *IBCEVMPacketData {
	if m != nil {
		return m.Packet
	}
	return nil
}

// MsgCallEVMResponse returns no fields
type MsgCallEVMResponse struct {
}

func (m *MsgCallEVMResponse) Reset()         { *m = MsgCallEVMResponse{} }
func (m *MsgCallEVMResponse) String() string { return proto.CompactTextString(m) }
func (*MsgCallEVMResponse) ProtoMessage()    {}
func (*MsgCallEVMResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_a40d98923f08b90e, []int{1}
}
func (m *MsgCallEVMResponse) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *MsgCallEVMResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_MsgCallEVMResponse.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *MsgCallEVMResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_MsgCallEVMResponse.Merge(m, src)
}
func (m *MsgCallEVMResponse) XXX_Size() int {
	return m.Size()
}
func (m *MsgCallEVMResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_MsgCallEVMResponse.DiscardUnknown(m)
}

var xxx_messageInfo_MsgCallEVMResponse proto.InternalMessageInfo

func init() {
	proto.RegisterType((*MsgCallEVM)(nil), "evmos.ibc.evm.v1.MsgCallEVM")
	proto.RegisterType((*MsgCallEVMResponse)(nil), "evmos.ibc.evm.v1.MsgCallEVMResponse")
}

func init() { proto.RegisterFile("evmos/ibc/evm/v1/tx.proto", fileDescriptor_a40d98923f08b90e) }

var fileDescriptor_a40d98923f08b90e = []byte{
	// 325 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0x92, 0x4c, 0x2d, 0xcb, 0xcd,
	0x2f, 0xd6, 0xcf, 0x4c, 0x4a, 0xd6, 0x4f, 0x2d, 0xcb, 0xd5, 0x2f, 0x33, 0xd4, 0x2f, 0xa9, 0xd0,
	0x2b, 0x28, 0xca, 0x2f, 0xc9, 0x17, 0x12, 0x00, 0x4b, 0xe9, 0x65, 0x26, 0x25, 0xeb, 0xa5, 0x96,
	0xe5, 0xea, 0x95, 0x19, 0x4a, 0xc9, 0xa4, 0xe7, 0xe7, 0xa7, 0xe7, 0xa4, 0xea, 0x27, 0x16, 0x64,
	0xea, 0x27, 0xe6, 0xe5, 0xe5, 0x97, 0x24, 0x96, 0x64, 0xe6, 0xe7, 0x15, 0x43, 0xd4, 0x4b, 0x89,
	0xa4, 0xe7, 0xa7, 0xe7, 0x83, 0x99, 0xfa, 0x20, 0x16, 0x54, 0x54, 0x16, 0xc3, 0x82, 0x82, 0xc4,
	0xe4, 0xec, 0xd4, 0x12, 0x88, 0xb4, 0x52, 0x0d, 0x17, 0x97, 0x6f, 0x71, 0xba, 0x73, 0x62, 0x4e,
	0x8e, 0x6b, 0x98, 0xaf, 0x90, 0x18, 0x17, 0x5b, 0x62, 0x6e, 0x7e, 0x69, 0x5e, 0x89, 0x04, 0xa3,
	0x02, 0xa3, 0x06, 0x67, 0x10, 0x94, 0x27, 0x24, 0xc2, 0xc5, 0x9a, 0x92, 0x9a, 0x97, 0x9f, 0x2b,
	0xc1, 0x04, 0x16, 0x86, 0x70, 0x84, 0xac, 0xb8, 0xd8, 0x20, 0x66, 0x49, 0x30, 0x2b, 0x30, 0x6a,
	0x70, 0x1b, 0x29, 0xe9, 0xa1, 0xbb, 0x58, 0xcf, 0xd3, 0xc9, 0xd9, 0x35, 0xcc, 0x37, 0x00, 0xac,
	0xca, 0x25, 0xb1, 0x24, 0x31, 0x08, 0xaa, 0xc3, 0x8a, 0xe5, 0xc5, 0x02, 0x79, 0x06, 0x25, 0x11,
	0x2e, 0x21, 0x84, 0xed, 0x41, 0xa9, 0xc5, 0x05, 0xf9, 0x79, 0xc5, 0xa9, 0x46, 0xd5, 0x5c, 0xcc,
	0xbe, 0xc5, 0xe9, 0x42, 0x25, 0x5c, 0xec, 0x30, 0x77, 0xc9, 0x60, 0x9a, 0x8c, 0xd0, 0x27, 0xa5,
	0x82, 0x4f, 0x16, 0x66, 0xaa, 0x92, 0x6a, 0xd3, 0xe5, 0x27, 0x93, 0x99, 0xe4, 0x85, 0x64, 0xf5,
	0xb1, 0x04, 0xb9, 0x7e, 0x72, 0x62, 0x4e, 0x4e, 0x7c, 0x6a, 0x59, 0xae, 0x93, 0xf3, 0x89, 0x47,
	0x72, 0x8c, 0x17, 0x1e, 0xc9, 0x31, 0x3e, 0x78, 0x24, 0xc7, 0x38, 0xe1, 0xb1, 0x1c, 0xc3, 0x85,
	0xc7, 0x72, 0x0c, 0x37, 0x1e, 0xcb, 0x31, 0x44, 0x69, 0xa6, 0x67, 0x96, 0x64, 0x94, 0x26, 0xe9,
	0x25, 0xe7, 0xe7, 0x42, 0x8d, 0x80, 0x90, 0x65, 0x96, 0xfa, 0x15, 0x70, 0xd3, 0x4a, 0x2a, 0x0b,
	0x52, 0x8b, 0x93, 0xd8, 0xc0, 0x81, 0x6b, 0x0c, 0x08, 0x00, 0x00, 0xff, 0xff, 0xbe, 0x54, 0x0f,
	0xaa, 0xde, 0x01, 0x00, 0x00,
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// MsgClient is the client API for Msg service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type MsgClient interface {
	// CallEVM is called on an ibc source chain to execute an evm tx on the source
	// chain
	CallEVM(ctx context.Context, in *MsgCallEVM, opts ...grpc.CallOption) (*MsgCallEVMResponse, error)
}

type msgClient struct {
	cc grpc1.ClientConn
}

func NewMsgClient(cc grpc1.ClientConn) MsgClient {
	return &msgClient{cc}
}

func (c *msgClient) CallEVM(ctx context.Context, in *MsgCallEVM, opts ...grpc.CallOption) (*MsgCallEVMResponse, error) {
	out := new(MsgCallEVMResponse)
	err := c.cc.Invoke(ctx, "/evmos.ibc.evm.v1.Msg/CallEVM", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// MsgServer is the server API for Msg service.
type MsgServer interface {
	// CallEVM is called on an ibc source chain to execute an evm tx on the source
	// chain
	CallEVM(context.Context, *MsgCallEVM) (*MsgCallEVMResponse, error)
}

// UnimplementedMsgServer can be embedded to have forward compatible implementations.
type UnimplementedMsgServer struct {
}

func (*UnimplementedMsgServer) CallEVM(ctx context.Context, req *MsgCallEVM) (*MsgCallEVMResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CallEVM not implemented")
}

func RegisterMsgServer(s grpc1.Server, srv MsgServer) {
	s.RegisterService(&_Msg_serviceDesc, srv)
}

func _Msg_CallEVM_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MsgCallEVM)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MsgServer).CallEVM(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/evmos.ibc.evm.v1.Msg/CallEVM",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MsgServer).CallEVM(ctx, req.(*MsgCallEVM))
	}
	return interceptor(ctx, in, info, handler)
}

var _Msg_serviceDesc = grpc.ServiceDesc{
	ServiceName: "evmos.ibc.evm.v1.Msg",
	HandlerType: (*MsgServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "CallEVM",
			Handler:    _Msg_CallEVM_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "evmos/ibc/evm/v1/tx.proto",
}

func (m *MsgCallEVM) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *MsgCallEVM) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *MsgCallEVM) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if m.Packet != nil {
		{
			size, err := m.Packet.MarshalToSizedBuffer(dAtA[:i])
			if err != nil {
				return 0, err
			}
			i -= size
			i = encodeVarintTx(dAtA, i, uint64(size))
		}
		i--
		dAtA[i] = 0x1a
	}
	if len(m.Denom) > 0 {
		i -= len(m.Denom)
		copy(dAtA[i:], m.Denom)
		i = encodeVarintTx(dAtA, i, uint64(len(m.Denom)))
		i--
		dAtA[i] = 0x12
	}
	if len(m.Amount) > 0 {
		i -= len(m.Amount)
		copy(dAtA[i:], m.Amount)
		i = encodeVarintTx(dAtA, i, uint64(len(m.Amount)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func (m *MsgCallEVMResponse) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *MsgCallEVMResponse) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *MsgCallEVMResponse) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	return len(dAtA) - i, nil
}

func encodeVarintTx(dAtA []byte, offset int, v uint64) int {
	offset -= sovTx(v)
	base := offset
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return base
}
func (m *MsgCallEVM) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.Amount)
	if l > 0 {
		n += 1 + l + sovTx(uint64(l))
	}
	l = len(m.Denom)
	if l > 0 {
		n += 1 + l + sovTx(uint64(l))
	}
	if m.Packet != nil {
		l = m.Packet.Size()
		n += 1 + l + sovTx(uint64(l))
	}
	return n
}

func (m *MsgCallEVMResponse) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	return n
}

func sovTx(x uint64) (n int) {
	return (math_bits.Len64(x|1) + 6) / 7
}
func sozTx(x uint64) (n int) {
	return sovTx(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *MsgCallEVM) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowTx
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
			return fmt.Errorf("proto: MsgCallEVM: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: MsgCallEVM: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Amount", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTx
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
				return ErrInvalidLengthTx
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthTx
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Amount = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Denom", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTx
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
				return ErrInvalidLengthTx
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthTx
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Denom = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Packet", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTx
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
				return ErrInvalidLengthTx
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthTx
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if m.Packet == nil {
				m.Packet = &IBCEVMPacketData{}
			}
			if err := m.Packet.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipTx(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthTx
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
func (m *MsgCallEVMResponse) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowTx
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
			return fmt.Errorf("proto: MsgCallEVMResponse: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: MsgCallEVMResponse: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		default:
			iNdEx = preIndex
			skippy, err := skipTx(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthTx
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
func skipTx(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	depth := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowTx
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
					return 0, ErrIntOverflowTx
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
					return 0, ErrIntOverflowTx
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
				return 0, ErrInvalidLengthTx
			}
			iNdEx += length
		case 3:
			depth++
		case 4:
			if depth == 0 {
				return 0, ErrUnexpectedEndOfGroupTx
			}
			depth--
		case 5:
			iNdEx += 4
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
		if iNdEx < 0 {
			return 0, ErrInvalidLengthTx
		}
		if depth == 0 {
			return iNdEx, nil
		}
	}
	return 0, io.ErrUnexpectedEOF
}

var (
	ErrInvalidLengthTx        = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowTx          = fmt.Errorf("proto: integer overflow")
	ErrUnexpectedEndOfGroupTx = fmt.Errorf("proto: unexpected end of group")
)
