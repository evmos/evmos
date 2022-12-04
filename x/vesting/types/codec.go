package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/auth/vesting/exported"
	sdkvesting "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
)

var (
	amino = codec.NewLegacyAmino()
	// ModuleCdc references the global vesting module codec. Note, the codec should
	// ONLY be used in certain instances of tests and for JSON encoding.
	ModuleCdc = codec.NewProtoCodec(codectypes.NewInterfaceRegistry())
	// AminoCdc is a amino codec created to support amino JSON compatible msgs.
	AminoCdc = codec.NewAminoCodec(amino)
)

const (
	// Amino names
	clawback                     = "evmos/MsgClawback"
	createClawbackVestingAccount = "evmos/MsgCreateClawbackVestingAccount"
	updateVestingFunder          = "evmos/MsgUpdateVestingFunder"
)

// NOTE: This is required for the GetSignBytes function
func init() {
	RegisterLegacyAminoCodec(amino)
	amino.Seal()
}

// RegisterInterface associates protoName with AccountI and VestingAccount
// Interfaces and creates a registry of it's concrete implementations
func RegisterInterfaces(registry codectypes.InterfaceRegistry) {
	// NOTE: BaseVestingAccount is still supported to as it's the underlying embedded
	// vesting account type in the ClawbackVestingAccount
	registry.RegisterInterface(
		"cosmos.vesting.v1beta1.VestingAccount",
		(*exported.VestingAccount)(nil),
		&ClawbackVestingAccount{},
	)

	registry.RegisterImplementations(
		(*authtypes.AccountI)(nil),
		&sdkvesting.BaseVestingAccount{},
		&ClawbackVestingAccount{},
	)

	registry.RegisterImplementations(
		(*authtypes.GenesisAccount)(nil),
		&sdkvesting.BaseVestingAccount{},
		&ClawbackVestingAccount{},
	)

	registry.RegisterImplementations(
		(*sdk.Msg)(nil),
		&MsgClawback{},
		&MsgCreateClawbackVestingAccount{},
		&MsgUpdateVestingFunder{},
	)

	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}

// RegisterLegacyAminoCodec registers the necessary x/vesting interfaces and
// concrete types on the provided LegacyAmino codec. These types are used for
// Amino JSON serialization and EIP-712 compatibility.
func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&MsgClawback{}, clawback, nil)
	cdc.RegisterConcrete(&MsgCreateClawbackVestingAccount{}, createClawbackVestingAccount, nil)
	cdc.RegisterConcrete(&MsgUpdateVestingFunder{}, updateVestingFunder, nil)
}
