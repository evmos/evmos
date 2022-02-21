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

// RegisterLegacyAminoCodec registers the vesting interfaces and concrete types on the
// provided LegacyAmino codec. These types are used for Amino JSON serialization
func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterInterface((*exported.VestingAccount)(nil), nil)
	cdc.RegisterConcrete(&sdkvesting.BaseVestingAccount{}, "cosmos-sdk/BaseVestingAccount", nil)
	cdc.RegisterConcrete(&sdkvesting.ContinuousVestingAccount{}, "cosmos-sdk/ContinuousVestingAccount", nil)
	cdc.RegisterConcrete(&sdkvesting.DelayedVestingAccount{}, "cosmos-sdk/DelayedVestingAccount", nil)
	cdc.RegisterConcrete(&sdkvesting.PeriodicVestingAccount{}, "cosmos-sdk/PeriodicVestingAccount", nil)
	cdc.RegisterConcrete(&sdkvesting.PermanentLockedAccount{}, "cosmos-sdk/PermanentLockedAccount", nil)
}

// RegisterInterface associates protoName with AccountI and VestingAccount
// Interfaces and creates a registry of it's concrete implementations
func RegisterInterfaces(registry codectypes.InterfaceRegistry) {
	registry.RegisterInterface(
		"cosmos.vesting.v1beta1.VestingAccount",
		(*exported.VestingAccount)(nil),
		&sdkvesting.ContinuousVestingAccount{},
		&sdkvesting.DelayedVestingAccount{},
		&sdkvesting.PeriodicVestingAccount{},
		&sdkvesting.PermanentLockedAccount{},
		&ClawbackVestingAccount{},
	)

	registry.RegisterImplementations(
		(*authtypes.AccountI)(nil),
		&sdkvesting.ContinuousVestingAccount{},
		&sdkvesting.DelayedVestingAccount{},
		&sdkvesting.PeriodicVestingAccount{},
		&sdkvesting.PermanentLockedAccount{},
		&ClawbackVestingAccount{},
	)

	registry.RegisterImplementations(
		(*authtypes.GenesisAccount)(nil),
		&sdkvesting.ContinuousVestingAccount{},
		&sdkvesting.DelayedVestingAccount{},
		&sdkvesting.PeriodicVestingAccount{},
		&sdkvesting.PermanentLockedAccount{},
		&ClawbackVestingAccount{},
	)

	registry.RegisterImplementations(
		(*sdk.Msg)(nil),
		&MsgClawback{},
		&MsgCreateClawbackVestingAccount{},
	)

	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}

var amino = codec.NewLegacyAmino()

func init() {
	RegisterLegacyAminoCodec(amino)
	amino.Seal()
}
