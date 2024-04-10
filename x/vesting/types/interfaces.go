// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package types

import (
	"time"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	v1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

// AccountKeeper defines the expected interface contract the vesting module
// requires for storing accounts.
type AccountKeeper interface {
	GetModuleAddress(name string) sdk.AccAddress
	GetAccount(sdk.Context, sdk.AccAddress) authtypes.AccountI
	SetAccount(sdk.Context, authtypes.AccountI)
	NewAccount(ctx sdk.Context, acc authtypes.AccountI) authtypes.AccountI
	NewAccountWithAddress(ctx sdk.Context, addr sdk.AccAddress) authtypes.AccountI
	IterateAccounts(ctx sdk.Context, process func(authtypes.AccountI) bool)
	RemoveAccount(ctx sdk.Context, acc authtypes.AccountI)
}

// BankKeeper defines the expected interface contract the vesting module requires
// for creating vesting accounts with funds.
type BankKeeper interface {
	GetAllBalances(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins
	SendCoins(ctx sdk.Context, fromAddr sdk.AccAddress, toAddr sdk.AccAddress, amt sdk.Coins) error
	SpendableCoins(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins
	BlockedAddr(addr sdk.AccAddress) bool
}

// StakingKeeper defines the expected interface contract the vesting module
// requires for finding and changing the delegated tokens, used in clawback.
type StakingKeeper interface {
	GetParams(ctx sdk.Context) stakingtypes.Params
	BondDenom(ctx sdk.Context) string
	GetDelegatorDelegations(ctx sdk.Context, delegator sdk.AccAddress, maxRetrieve uint16) []stakingtypes.Delegation
	GetUnbondingDelegations(ctx sdk.Context, delegator sdk.AccAddress, maxRetrieve uint16) []stakingtypes.UnbondingDelegation
	GetValidator(ctx sdk.Context, valAddr sdk.ValAddress) (stakingtypes.Validator, bool)

	// Support iterating delegations for use in ante handlers
	IterateDelegations(ctx sdk.Context, delegator sdk.AccAddress, fn func(index int64, delegation stakingtypes.DelegationI) (stop bool))

	// Support functions for Agoric's custom stakingkeeper logic on vestingkeeper
	GetUnbondingDelegation(ctx sdk.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) (stakingtypes.UnbondingDelegation, bool)
	HasMaxUnbondingDelegationEntries(ctx sdk.Context, delegatorAddr sdk.AccAddress, validatorAddr sdk.ValAddress) bool
	SetUnbondingDelegationEntry(ctx sdk.Context, delegatorAddr sdk.AccAddress, validatorAddr sdk.ValAddress, creationHeight int64, minTime time.Time, balance math.Int) stakingtypes.UnbondingDelegation
	InsertUBDQueue(ctx sdk.Context, ubd stakingtypes.UnbondingDelegation, completionTime time.Time)
	RemoveUnbondingDelegation(ctx sdk.Context, ubd stakingtypes.UnbondingDelegation)
	SetUnbondingDelegation(ctx sdk.Context, ubd stakingtypes.UnbondingDelegation)
	GetDelegation(ctx sdk.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) (stakingtypes.Delegation, bool)
	GetRedelegation(ctx sdk.Context, delAddr sdk.AccAddress, valSrcAddr, valDstAddr sdk.ValAddress) (stakingtypes.Redelegation, bool)
	MaxEntries(ctx sdk.Context) uint32
	SetDelegation(ctx sdk.Context, delegation stakingtypes.Delegation)
	RemoveDelegation(ctx sdk.Context, delegation stakingtypes.Delegation) error
	GetRedelegations(ctx sdk.Context, delegator sdk.AccAddress, maxRetrieve uint16) []stakingtypes.Redelegation
	SetRedelegationEntry(ctx sdk.Context, delegatorAddr sdk.AccAddress, validatorSrcAddr, validatorDstAddr sdk.ValAddress, creationHeight int64, minTime time.Time, balance math.Int, sharesSrc, sharesDst math.LegacyDec) stakingtypes.Redelegation
	InsertRedelegationQueue(ctx sdk.Context, red stakingtypes.Redelegation, completionTime time.Time)
	SetRedelegation(ctx sdk.Context, red stakingtypes.Redelegation)
	RemoveRedelegation(ctx sdk.Context, red stakingtypes.Redelegation)
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
