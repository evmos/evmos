package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
)

// ModuleCdc references the global intrarelayer module codec. Note, the codec should
// ONLY be used in certain instances of tests and for JSON encoding.
//
// The actual codec used for serialization should be provided to modules/intrarelayer and
// defined at the application level.
var ModuleCdc = codec.NewProtoCodec(codectypes.NewInterfaceRegistry())

// RegisterLegacyAminoCodec registers concrete types on the LegacyAmino codec
func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(TokenPair{}, "intrarelayer/TokenPair", nil)
	cdc.RegisterConcrete(&RegisterERC20Proposal{}, "intrarelayer/RegisterERC20Proposal", nil)
	cdc.RegisterConcrete(&RegisterCoinProposal{}, "intrarelayer/RegisterCoinProposal", nil)
	cdc.RegisterConcrete(&ToggleTokenRelayProposal{}, "intrarelayer/ToggleTokenRelayProposal", nil)
	cdc.RegisterConcrete(&UpdateTokenPairERC20Proposal{}, "intrarelayer/UpdateTokenPairERC20Proposal", nil)
}

// RegisterInterfaces register implementations
func RegisterInterfaces(registry codectypes.InterfaceRegistry) {
	registry.RegisterImplementations(
		(*sdk.Msg)(nil),
		&MsgConvertCoin{},
		&MsgConvertERC20{},
	)
	registry.RegisterImplementations(
		(*govtypes.Content)(nil),
		&RegisterCoinProposal{},
		&RegisterERC20Proposal{},
		&ToggleTokenRelayProposal{},
		&UpdateTokenPairERC20Proposal{},
	)

	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}
