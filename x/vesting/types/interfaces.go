// Copyright 2022 Evmos Foundation
// This file is part of the Evmos Network packages.
//
// Evmos is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The Evmos packages are distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the Evmos packages. If not, see https://github.com/evmos/evmos/blob/main/LICENSE

package types

import (
	"time"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

// AccountKeeper defines the expected interface contract the vesting module
// requires for storing accounts.
type AccountKeeper interface {
	GetAccount(sdk.Context, sdk.AccAddress) authtypes.AccountI
	SetAccount(sdk.Context, authtypes.AccountI)
	NewAccount(ctx sdk.Context, acc authtypes.AccountI) authtypes.AccountI
	NewAccountWithAddress(ctx sdk.Context, addr sdk.AccAddress) authtypes.AccountI
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
	SetRedelegationEntry(ctx sdk.Context, delegatorAddr sdk.AccAddress, validatorSrcAddr, validatorDstAddr sdk.ValAddress, creationHeight int64, minTime time.Time, balance math.Int, sharesSrc, sharesDst sdk.Dec) stakingtypes.Redelegation
	InsertRedelegationQueue(ctx sdk.Context, red stakingtypes.Redelegation, completionTime time.Time)
	SetRedelegation(ctx sdk.Context, red stakingtypes.Redelegation)
	RemoveRedelegation(ctx sdk.Context, red stakingtypes.Redelegation)
	GetDelegatorUnbonding(ctx sdk.Context, delegator sdk.AccAddress) math.Int
	GetDelegatorBonded(ctx sdk.Context, delegator sdk.AccAddress) math.Int
	// Hooks
	stakingtypes.StakingHooks
}
