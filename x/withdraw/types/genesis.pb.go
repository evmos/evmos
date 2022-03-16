// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: evmos/withdraw/v1/genesis.proto

package types

import (
	fmt "fmt"
	_ "github.com/gogo/protobuf/gogoproto"
	proto "github.com/gogo/protobuf/proto"
	github_com_gogo_protobuf_types "github.com/gogo/protobuf/types"
	_ "google.golang.org/protobuf/types/known/durationpb"
	io "io"
	math "math"
	math_bits "math/bits"
	time "time"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf
var _ = time.Kitchen

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.GoGoProtoPackageIsVersion3 // please upgrade the proto package

// GenesisState defines the withdraw module's genesis state.
type GenesisState struct {
	// params defines all the paramaters of the module.
	Params Params `protobuf:"bytes,1,opt,name=params,proto3" json:"params"`
}

func (m *GenesisState) Reset()         { *m = GenesisState{} }
func (m *GenesisState) String() string { return proto.CompactTextString(m) }
func (*GenesisState) ProtoMessage()    {}
func (*GenesisState) Descriptor() ([]byte, []int) {
	return fileDescriptor_fd519519d2b9d615, []int{0}
}
func (m *GenesisState) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *GenesisState) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_GenesisState.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *GenesisState) XXX_Merge(src proto.Message) {
	xxx_messageInfo_GenesisState.Merge(m, src)
}
func (m *GenesisState) XXX_Size() int {
	return m.Size()
}
func (m *GenesisState) XXX_DiscardUnknown() {
	xxx_messageInfo_GenesisState.DiscardUnknown(m)
}

var xxx_messageInfo_GenesisState proto.InternalMessageInfo

func (m *GenesisState) GetParams() Params {
	if m != nil {
		return m.Params
	}
	return Params{}
}

// Params holds parameters for the withdraw module
type Params struct {
	// enable withdraw IBC middleware
	EnableWithdraw bool `protobuf:"varint,1,opt,name=enable_withdraw,json=enableWithdraw,proto3" json:"enable_withdraw,omitempty"`
	// duration added to timeout timestamp for balances withdrawn via IBC packets
	PacketTimeoutDuration time.Duration `protobuf:"bytes,2,opt,name=packet_timeout_duration,json=packetTimeoutDuration,proto3,stdduration" json:"packet_timeout_duration"`
}

func (m *Params) Reset()         { *m = Params{} }
func (m *Params) String() string { return proto.CompactTextString(m) }
func (*Params) ProtoMessage()    {}
func (*Params) Descriptor() ([]byte, []int) {
	return fileDescriptor_fd519519d2b9d615, []int{1}
}
func (m *Params) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *Params) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_Params.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *Params) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Params.Merge(m, src)
}
func (m *Params) XXX_Size() int {
	return m.Size()
}
func (m *Params) XXX_DiscardUnknown() {
	xxx_messageInfo_Params.DiscardUnknown(m)
}

var xxx_messageInfo_Params proto.InternalMessageInfo

func (m *Params) GetEnableWithdraw() bool {
	if m != nil {
		return m.EnableWithdraw
	}
	return false
}

func (m *Params) GetPacketTimeoutDuration() time.Duration {
	if m != nil {
		return m.PacketTimeoutDuration
	}
	return 0
}

func init() {
	proto.RegisterType((*GenesisState)(nil), "evmos.withdraw.v1.GenesisState")
	proto.RegisterType((*Params)(nil), "evmos.withdraw.v1.Params")
}

func init() { proto.RegisterFile("evmos/withdraw/v1/genesis.proto", fileDescriptor_fd519519d2b9d615) }

var fileDescriptor_fd519519d2b9d615 = []byte{
	// 298 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x64, 0x90, 0xc1, 0x4b, 0xc3, 0x30,
	0x14, 0xc6, 0x1b, 0x91, 0x31, 0xa2, 0x28, 0x16, 0xc5, 0xb9, 0x43, 0x26, 0xbb, 0xe8, 0x41, 0x12,
	0x36, 0x0f, 0xde, 0x87, 0xb8, 0xab, 0x4c, 0x41, 0xd0, 0x43, 0x49, 0xb7, 0x98, 0x06, 0xd7, 0xa6,
	0x34, 0xaf, 0x9d, 0xfe, 0x13, 0xe2, 0xd1, 0x3f, 0x69, 0xc7, 0x1d, 0x3d, 0xa9, 0xb4, 0xff, 0x88,
	0x2c, 0x69, 0xf5, 0xe0, 0x2d, 0x79, 0xdf, 0xef, 0x7d, 0xdf, 0xc7, 0xc3, 0x3d, 0x51, 0xc4, 0xda,
	0xb0, 0x85, 0x82, 0x68, 0x96, 0xf1, 0x05, 0x2b, 0x06, 0x4c, 0x8a, 0x44, 0x18, 0x65, 0x68, 0x9a,
	0x69, 0xd0, 0xfe, 0x9e, 0x05, 0x68, 0x03, 0xd0, 0x62, 0xd0, 0xdd, 0x97, 0x5a, 0x6a, 0xab, 0xb2,
	0xf5, 0xcb, 0x81, 0x5d, 0x22, 0xb5, 0x96, 0x73, 0xc1, 0xec, 0x2f, 0xcc, 0x1f, 0xd9, 0x2c, 0xcf,
	0x38, 0x28, 0x9d, 0x38, 0xbd, 0x3f, 0xc6, 0xdb, 0x63, 0xe7, 0x7c, 0x03, 0x1c, 0x84, 0x7f, 0x81,
	0x5b, 0x29, 0xcf, 0x78, 0x6c, 0x3a, 0xe8, 0x18, 0x9d, 0x6e, 0x0d, 0x8f, 0xe8, 0xbf, 0x24, 0x7a,
	0x6d, 0x81, 0xd1, 0xe6, 0xf2, 0xb3, 0xe7, 0x4d, 0x6a, 0xbc, 0xff, 0x8a, 0x70, 0xcb, 0x09, 0xfe,
	0x09, 0xde, 0x15, 0x09, 0x0f, 0xe7, 0x22, 0x68, 0xb6, 0xac, 0x59, 0x7b, 0xb2, 0xe3, 0xc6, 0x77,
	0xf5, 0xd4, 0x7f, 0xc0, 0x87, 0x29, 0x9f, 0x3e, 0x09, 0x08, 0x40, 0xc5, 0x42, 0xe7, 0x10, 0x34,
	0xed, 0x3a, 0x1b, 0x75, 0xba, 0xab, 0x4f, 0x9b, 0xfa, 0xf4, 0xb2, 0x06, 0x46, 0xed, 0x75, 0xfa,
	0xfb, 0x57, 0x0f, 0x4d, 0x0e, 0x9c, 0xc7, 0xad, 0xb3, 0xf8, 0x05, 0xae, 0x96, 0x25, 0x41, 0xab,
	0x92, 0xa0, 0xef, 0x92, 0xa0, 0xb7, 0x8a, 0x78, 0xab, 0x8a, 0x78, 0x1f, 0x15, 0xf1, 0xee, 0xcf,
	0xa4, 0x82, 0x28, 0x0f, 0xe9, 0x54, 0xc7, 0x0c, 0x22, 0x9e, 0x19, 0x65, 0x98, 0x3b, 0x78, 0x31,
	0x64, 0xcf, 0x7f, 0x57, 0x87, 0x97, 0x54, 0x98, 0xb0, 0x65, 0xb3, 0xcf, 0x7f, 0x02, 0x00, 0x00,
	0xff, 0xff, 0x80, 0x5c, 0x38, 0x02, 0x94, 0x01, 0x00, 0x00,
}

func (m *GenesisState) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *GenesisState) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *GenesisState) MarshalToSizedBuffer(dAtA []byte) (int, error) {
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
		i = encodeVarintGenesis(dAtA, i, uint64(size))
	}
	i--
	dAtA[i] = 0xa
	return len(dAtA) - i, nil
}

func (m *Params) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *Params) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *Params) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	n2, err2 := github_com_gogo_protobuf_types.StdDurationMarshalTo(m.PacketTimeoutDuration, dAtA[i-github_com_gogo_protobuf_types.SizeOfStdDuration(m.PacketTimeoutDuration):])
	if err2 != nil {
		return 0, err2
	}
	i -= n2
	i = encodeVarintGenesis(dAtA, i, uint64(n2))
	i--
	dAtA[i] = 0x12
	if m.EnableWithdraw {
		i--
		if m.EnableWithdraw {
			dAtA[i] = 1
		} else {
			dAtA[i] = 0
		}
		i--
		dAtA[i] = 0x8
	}
	return len(dAtA) - i, nil
}

func encodeVarintGenesis(dAtA []byte, offset int, v uint64) int {
	offset -= sovGenesis(v)
	base := offset
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return base
}
func (m *GenesisState) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = m.Params.Size()
	n += 1 + l + sovGenesis(uint64(l))
	return n
}

func (m *Params) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if m.EnableWithdraw {
		n += 2
	}
	l = github_com_gogo_protobuf_types.SizeOfStdDuration(m.PacketTimeoutDuration)
	n += 1 + l + sovGenesis(uint64(l))
	return n
}

func sovGenesis(x uint64) (n int) {
	return (math_bits.Len64(x|1) + 6) / 7
}
func sozGenesis(x uint64) (n int) {
	return sovGenesis(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *GenesisState) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowGenesis
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
			return fmt.Errorf("proto: GenesisState: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: GenesisState: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Params", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
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
				return ErrInvalidLengthGenesis
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthGenesis
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
			skippy, err := skipGenesis(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthGenesis
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
func (m *Params) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowGenesis
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
			return fmt.Errorf("proto: Params: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: Params: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field EnableWithdraw", wireType)
			}
			var v int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				v |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			m.EnableWithdraw = bool(v != 0)
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field PacketTimeoutDuration", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
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
				return ErrInvalidLengthGenesis
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthGenesis
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := github_com_gogo_protobuf_types.StdDurationUnmarshal(&m.PacketTimeoutDuration, dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipGenesis(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthGenesis
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
func skipGenesis(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	depth := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowGenesis
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
					return 0, ErrIntOverflowGenesis
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
					return 0, ErrIntOverflowGenesis
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
				return 0, ErrInvalidLengthGenesis
			}
			iNdEx += length
		case 3:
			depth++
		case 4:
			if depth == 0 {
				return 0, ErrUnexpectedEndOfGroupGenesis
			}
			depth--
		case 5:
			iNdEx += 4
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
		if iNdEx < 0 {
			return 0, ErrInvalidLengthGenesis
		}
		if depth == 0 {
			return iNdEx, nil
		}
	}
	return 0, io.ErrUnexpectedEOF
}

var (
	ErrInvalidLengthGenesis        = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowGenesis          = fmt.Errorf("proto: integer overflow")
	ErrUnexpectedEndOfGroupGenesis = fmt.Errorf("proto: unexpected end of group")
)
