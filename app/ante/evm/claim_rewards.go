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

package evm

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

// ClaimSufficientStakingRewards checks if the account has enough staking rewards unclaimed
// to cover the given amount. If more than enough rewards are unclaimed, only those up to
// the given amount are claimed.
func ClaimSufficientStakingRewards(ctx sdk.Context, stakingKeeper StakingKeeper, distributionKeeper DistributionKeeper, addr sdk.AccAddress, amount sdk.Coins) error {
	var (
		err     error
		reward  sdk.Coins
		rewards sdk.Coins
	)

	cacheCtx, writeFn := ctx.CacheContext()

	// iterate through all delegations and get the rewards if any are unclaimed.
	stakingKeeper.IterateDelegations(
		cacheCtx, addr, func(_ int64, delegation stakingtypes.DelegationI) (stop bool) {
			reward, err = distributionKeeper.WithdrawDelegationRewards(cacheCtx, addr, delegation.GetValidatorAddr())
			if err != nil {
				return true
			}
			rewards = rewards.Add(reward...)

			// FIXME: is there a better way to do this? probably not necessary to check if ANY is gte but rather check the specific denom.
			return rewards.IsAnyGTE(amount)
		},
	)

	// check if there was an error while iterating delegations
	if err != nil {
		return fmt.Errorf("error while withdrawing delegation rewards: %s", err)
	}

	// commit state changes
	writeFn()

	return nil
}
