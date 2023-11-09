// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package types

import (
	context "context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	channeltypes "github.com/cosmos/ibc-go/v8/modules/core/04-channel/types"
)

// BankKeeper defines the banking contract that must be fulfilled when
// creating a x/claims keeper.
type BankKeeper interface {
	SendCoinsFromModuleToAccount(ctx context.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error
	SendCoinsFromModuleToModule(ctx context.Context, senderPool, recipientPool string, amt sdk.Coins) error
	GetAllBalances(ctx context.Context, addr sdk.AccAddress) sdk.Coins
	GetBalance(ctx context.Context, addr sdk.AccAddress, denom string) sdk.Coin
	MintCoins(ctx context.Context, moduleName string, amt sdk.Coins) error
	BlockedAddr(address sdk.AccAddress) bool
}

// AccountKeeper defines the expected account keeper used for simulations (noalias)
type AccountKeeper interface {
	GetModuleAddress(name string) sdk.AccAddress
	GetModuleAccount(ctx context.Context, name string) sdk.ModuleAccountI
	SetModuleAccount(ctx context.Context, macc sdk.ModuleAccountI)
	GetAccount(context.Context, sdk.AccAddress) sdk.AccountI
	GetSequence(context.Context, sdk.AccAddress) (uint64, error)
	RemoveAccount(ctx context.Context, account sdk.AccountI)
}

// ChannelKeeper is the keeper that controls access to IBC channels
type ChannelKeeper interface {
	GetChannel(ctx sdk.Context, portID, channelID string) (channel channeltypes.Channel, found bool)
}

// DistrKeeper is the keeper of the distribution store
type DistrKeeper interface {
	FundCommunityPool(ctx sdk.Context, amount sdk.Coins, sender sdk.AccAddress) error
}

// StakingKeeper expected staking keeper (noalias)
type StakingKeeper interface {
	// BondDenom - Bondable coin denomination
	BondDenom(sdk.Context) string
}

type (
	LegacyParams = paramtypes.ParamSet
	// Subspace defines an interface that implements the legacy Cosmos SDK x/params Subspace type.
	// NOTE: This is used solely for migration of the Cosmos SDK x/params managed parameters.
	Subspace interface {
		GetParamSetIfExists(ctx sdk.Context, ps LegacyParams)
		WithKeyTable(table paramtypes.KeyTable) paramtypes.Subspace
	}
)
