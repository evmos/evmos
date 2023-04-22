// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package utils

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

// BankKeeper defines the exposed interface for using functionality of the bank keeper
// in the context of the AnteHandler utils package.
type BankKeeper interface {
	GetBalance(ctx sdk.Context, addr sdk.AccAddress, denom string) sdk.Coin
}

// DistributionKeeper defines the exposed interface for using functionality of the distribution
// keeper in the context of the AnteHandler utils package.
type DistributionKeeper interface {
	WithdrawDelegationRewards(ctx sdk.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) (sdk.Coins, error)
}

// StakingKeeper defines the exposed interface for using functionality of the staking keeper
// in the context of the AnteHandler utils package.
type StakingKeeper interface {
	BondDenom(ctx sdk.Context) string
	IterateDelegations(ctx sdk.Context, delegator sdk.AccAddress, fn func(index int64, delegation stakingtypes.DelegationI) (stop bool))
}
