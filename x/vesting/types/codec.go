// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/auth/vesting/exported"
	sdkvesting "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	govv1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	migrationtypes "github.com/evmos/evmos/v20/x/vesting/migrations/types"
)

var (
	amino = codec.NewLegacyAmino()
	// ModuleCdc references the global vesting  module codec. Note, the codec should
	// ONLY be used in certain instances of tests and for JSON encoding.
	ModuleCdc = codec.NewProtoCodec(codectypes.NewInterfaceRegistry())
	// AminoCdc is a amino codec created to support amino JSON compatible msgs.
	AminoCdc = codec.NewAminoCodec(amino) //nolint:staticcheck
)

const (
	// Amino names
	clawback                     = "evmos/MsgClawback"
	createClawbackVestingAccount = "evmos/MsgCreateClawbackVestingAccount"
	updateVestingFunder          = "evmos/MsgUpdateVestingFunder"
	convertVestingAccount        = "evmos/MsgConvertVestingAccount"
	fundVestingAccount           = "evmos/MsgFundVestingAccount"
)

// NOTE: This is required for the GetSignBytes function
func init() {
	RegisterLegacyAminoCodec(amino)
	amino.Seal()
}

// RegisterInterfaces associates protoName with AccountI and VestingAccount
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
		(*sdk.AccountI)(nil),
		&sdkvesting.BaseVestingAccount{},
		&ClawbackVestingAccount{},
		&migrationtypes.ClawbackVestingAccount{},
	)

	registry.RegisterImplementations(
		(*authtypes.GenesisAccount)(nil),
		&sdkvesting.BaseVestingAccount{},
		&ClawbackVestingAccount{},
		&migrationtypes.ClawbackVestingAccount{},
	)

	registry.RegisterImplementations(
		(*sdk.Msg)(nil),
		&MsgClawback{},
		&MsgCreateClawbackVestingAccount{},
		&MsgUpdateVestingFunder{},
		&MsgFundVestingAccount{},
		&MsgConvertVestingAccount{},
	)

	registry.RegisterImplementations(
		(*govv1beta1.Content)(nil),
		&ClawbackProposal{},
	)

	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}

// RegisterLegacyAminoCodec registers the necessary x/erc20 interfaces and
// concrete types on the provided LegacyAmino codec. These types are used for
// Amino JSON serialization and EIP-712 compatibility.
func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&MsgClawback{}, clawback, nil)
	cdc.RegisterConcrete(&MsgCreateClawbackVestingAccount{}, createClawbackVestingAccount, nil)
	cdc.RegisterConcrete(&MsgUpdateVestingFunder{}, updateVestingFunder, nil)
	cdc.RegisterConcrete(&MsgConvertVestingAccount{}, convertVestingAccount, nil)
	cdc.RegisterConcrete(&MsgFundVestingAccount{}, fundVestingAccount, nil)
}
