// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package types

import (
	context "context"
	"time"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

// AccountKeeper defines the expected interface contract the vesting module
// requires for storing accounts.
type AccountKeeper interface {
	GetModuleAddress(name string) sdk.AccAddress
	GetAccount(context.Context, sdk.AccAddress) sdk.AccountI
	SetAccount(context.Context, sdk.AccountI)
	NewAccount(ctx context.Context, acc sdk.AccountI) sdk.AccountI
	NewAccountWithAddress(ctx context.Context, addr sdk.AccAddress) sdk.AccountI
	IterateAccounts(ctx context.Context, process func(sdk.AccountI) bool)
	RemoveAccount(ctx context.Context, acc sdk.AccountI)
}

// BankKeeper defines the expected interface contract the vesting module requires
// for creating vesting accounts with funds.
type BankKeeper interface {
	GetAllBalances(ctx context.Context, addr sdk.AccAddress) sdk.Coins
	SendCoins(ctx context.Context, fromAddr sdk.AccAddress, toAddr sdk.AccAddress, amt sdk.Coins) error
	SpendableCoins(ctx context.Context, addr sdk.AccAddress) sdk.Coins
	BlockedAddr(addr sdk.AccAddress) bool
}

// StakingKeeper defines the expected interface contract the vesting module
// requires for finding and changing the delegated tokens, used in clawback.
type StakingKeeper interface {
	GetParams(ctx context.Context) (stakingtypes.Params, error)
	BondDenom(ctx context.Context) (string, error)
	GetDelegatorDelegations(ctx context.Context, delegator sdk.AccAddress, maxRetrieve uint16) ([]stakingtypes.Delegation, error)
	GetUnbondingDelegations(ctx context.Context, delegator sdk.AccAddress, maxRetrieve uint16) ([]stakingtypes.UnbondingDelegation, error)
	GetValidator(ctx context.Context, valAddr sdk.ValAddress) (stakingtypes.Validator, error)

	// Support iterating delegations for use in ante handlers
	IterateDelegations(ctx context.Context, delegator sdk.AccAddress, fn func(index int64, delegation stakingtypes.DelegationI) (stop bool)) error

	// Support functions for Agoric's custom stakingkeeper logic on vestingkeeper
	GetUnbondingDelegation(ctx context.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) (stakingtypes.UnbondingDelegation, error)
	HasMaxUnbondingDelegationEntries(ctx context.Context, delegatorAddr sdk.AccAddress, validatorAddr sdk.ValAddress) (bool, error)
	SetUnbondingDelegationEntry(ctx context.Context, delegatorAddr sdk.AccAddress, validatorAddr sdk.ValAddress, creationHeight int64, minTime time.Time, balance math.Int) (stakingtypes.UnbondingDelegation, error)
	InsertUBDQueue(ctx context.Context, ubd stakingtypes.UnbondingDelegation, completionTime time.Time) error
	RemoveUnbondingDelegation(ctx context.Context, ubd stakingtypes.UnbondingDelegation) error
	SetUnbondingDelegation(ctx context.Context, ubd stakingtypes.UnbondingDelegation) error
	GetDelegation(ctx context.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) (stakingtypes.Delegation, error)
	GetRedelegation(ctx context.Context, delAddr sdk.AccAddress, valSrcAddr, valDstAddr sdk.ValAddress) (stakingtypes.Redelegation, error)
	MaxEntries(ctx context.Context) (uint32, error)
	SetDelegation(ctx context.Context, delegation stakingtypes.Delegation) error
	RemoveDelegation(ctx context.Context, delegation stakingtypes.Delegation) error
	GetRedelegations(ctx context.Context, delegator sdk.AccAddress, maxRetrieve uint16) ([]stakingtypes.Redelegation, error)
	SetRedelegationEntry(ctx context.Context, delegatorAddr sdk.AccAddress, validatorSrcAddr, validatorDstAddr sdk.ValAddress, creationHeight int64, minTime time.Time, balance math.Int, sharesSrc, sharesDst math.LegacyDec) (stakingtypes.Redelegation, error)
	InsertRedelegationQueue(ctx context.Context, red stakingtypes.Redelegation, completionTime time.Time) error
	SetRedelegation(ctx context.Context, red stakingtypes.Redelegation) error
	RemoveRedelegation(ctx context.Context, red stakingtypes.Redelegation) error
	GetDelegatorUnbonding(ctx context.Context, delegator sdk.AccAddress) (math.Int, error)
	GetDelegatorBonded(ctx context.Context, delegator sdk.AccAddress) (math.Int, error)

	// Hooks
	// Commented this out because go throws compiling error that a Hook is not implemented
	// even though it is implemented
	// stakingtypes.StakingHooks
}

// DistributionKeeper defines the expected interface contract the vesting module
// requires for clawing back unvested coins to the community pool.
type DistributionKeeper interface {
	FundCommunityPool(ctx context.Context, amount sdk.Coins, sender sdk.AccAddress) error
}
