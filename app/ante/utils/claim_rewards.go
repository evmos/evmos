// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package utils

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	errortypes "github.com/cosmos/cosmos-sdk/types/errors"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

// ClaimStakingRewardsIfNecessary checks if the given address has enough balance to cover the
// given amount. If not, it attempts to claim enough staking rewards to cover the amount.
func ClaimStakingRewardsIfNecessary(
	ctx sdk.Context,
	bankKeeper BankKeeper,
	distributionKeeper DistributionKeeper,
	stakingKeeper StakingKeeper,
	addr sdk.AccAddress,
	amount sdk.Coins,
) error {
	stakingDenom := stakingKeeper.BondDenom(ctx)
	found, amountInStakingDenom := amount.Find(stakingDenom)
	if !found {
		return errortypes.ErrInsufficientFee.Wrapf(
			"wrong fee denomination; got: %s; required: %s", amount, stakingDenom,
		)
	}

	balance := bankKeeper.GetBalance(ctx, addr, stakingDenom)
	if balance.IsNegative() {
		return errortypes.ErrInsufficientFunds.Wrapf("balance of %s in %s is negative", addr, stakingDenom)
	}

	// check if the account has enough balance to cover the fees
	if balance.IsGTE(amountInStakingDenom) {
		return nil
	}

	// Calculate the amount of staking rewards needed to cover the fees
	difference := amountInStakingDenom.Sub(balance)

	// attempt to claim enough staking rewards to cover the fees
	return ClaimSufficientStakingRewards(
		ctx, stakingKeeper, distributionKeeper, addr, difference,
	)
}

// ClaimSufficientStakingRewards checks if the account has enough staking rewards unclaimed
// to cover the given amount. If more than enough rewards are unclaimed, only those up to
// the given amount are claimed.
func ClaimSufficientStakingRewards(
	ctx sdk.Context,
	stakingKeeper StakingKeeper,
	distributionKeeper DistributionKeeper,
	addr sdk.AccAddress,
	amount sdk.Coin,
) error {
	var (
		err     error
		reward  sdk.Coins
		rewards sdk.Coins
	)

	// Allocate a cached context to avoid writing to state if there are not enough rewards
	cacheCtx, writeFn := ctx.CacheContext()

	// Iterate through delegations and get the rewards if any are unclaimed.
	// The loop stops once a sufficient amount was withdrawn.
	stakingKeeper.IterateDelegations(
		cacheCtx,
		addr,
		func(_ int64, delegation stakingtypes.DelegationI) (stop bool) {
			reward, err = distributionKeeper.WithdrawDelegationRewards(cacheCtx, addr, delegation.GetValidatorAddr())
			if err != nil {
				return true
			}
			rewards = rewards.Add(reward...)

			return rewards.AmountOf(amount.Denom).GTE(amount.Amount)
		},
	)

	// check if there was an error while iterating delegations
	if err != nil {
		return errorsmod.Wrap(err, "error while withdrawing delegation rewards")
	}

	// only write to state if there are enough rewards to cover the transaction fees
	if rewards.AmountOf(amount.Denom).LT(amount.Amount) {
		return errortypes.ErrInsufficientFee.Wrapf("insufficient staking rewards to cover transaction fees")
	}
	writeFn() // commit state changes
	return nil
}
