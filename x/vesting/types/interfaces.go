// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package types

import (
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	v1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
)

// AccountKeeper defines the expected interface contract the vesting module
// requires for storing accounts.
type AccountKeeper interface {
	GetModuleAddress(name string) sdk.AccAddress
	GetAccount(sdk.Context, sdk.AccAddress) authtypes.AccountI
	SetAccount(sdk.Context, authtypes.AccountI)
	IterateAccounts(ctx sdk.Context, process func(authtypes.AccountI) bool)
	RemoveAccount(ctx sdk.Context, acc authtypes.AccountI)
}

// BankKeeper defines the expected interface contract the vesting module requires
// for creating vesting accounts with funds.
type BankKeeper interface {
	SendCoins(ctx sdk.Context, fromAddr sdk.AccAddress, toAddr sdk.AccAddress, amt sdk.Coins) error
	BlockedAddr(addr sdk.AccAddress) bool
}

// StakingKeeper defines the expected interface contract the vesting module
// requires for finding and changing the delegated tokens, used in clawback.
type StakingKeeper interface {
	BondDenom(ctx sdk.Context) string

	// Support functions for Agoric's custom stakingkeeper logic on vestingkeeper
	GetDelegatorUnbonding(ctx sdk.Context, delegator sdk.AccAddress) math.Int
	GetDelegatorBonded(ctx sdk.Context, delegator sdk.AccAddress) math.Int
}

// DistributionKeeper defines the expected interface contract the vesting module
// requires for clawing back unvested coins to the community pool.
type DistributionKeeper interface {
	FundCommunityPool(ctx sdk.Context, amount sdk.Coins, sender sdk.AccAddress) error
}

// GovKeeper defines the expected interface contract the vesting module requires
// for accessing governance related information.
type GovKeeper interface {
	GetParams(ctx sdk.Context) v1.Params
	GetProposal(ctx sdk.Context, proposalID uint64) (v1.Proposal, bool)
}
