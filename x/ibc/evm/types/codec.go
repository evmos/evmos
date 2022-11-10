package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
)

var (
	amino = codec.NewLegacyAmino()

	// AminoCdc is a amino codec created to support JSON compatible msgs.
	AminoCdc = codec.NewAminoCodec(amino)
)

const (
	//Amino names
	callEVMName = "evmos/ibc/MsgCallEVM"
)

// NOTE: This is required for the GetSignBytes function
func init() {
	RegisterLegacyAminoCodec(amino)
	amino.Seal()
}

// RegisterLegacyAminoCodec registers the necessary x/ibc transfer interfaces and concrete types
// on the provided LegacyAmino codec. These types are used for Amino JSON serialization.
func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&MsgCallEVM{}, callEVMName, nil)
}

// RegisterInterfaces register the ibc transfer module interfaces to protobuf
// Any.
func RegisterInterfaces(registry codectypes.InterfaceRegistry) {
	registry.RegisterImplementations((*sdk.Msg)(nil), &MsgCallEVM{})

	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}
