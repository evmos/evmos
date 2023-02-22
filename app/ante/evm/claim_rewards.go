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
	"log"

	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/evmos/evmos/v11/utils"
)

// ClaimSufficientStakingRewards checks if the account has enough staking rewards unclaimed
// to cover the given amount. If more than enough rewards are unclaimed, only those up to
// the given amount are claimed.
func ClaimSufficientStakingRewards(ctx sdk.Context, stakingKeeper StakingKeeper, distributionKeeper DistributionKeeper, addr sdk.AccAddress, amount sdk.Coins) error {
	var (
		err     error
		reward  sdk.Coins
		rewards sdk.Coins

		// baseAmount defines the amount to be claimed in the base denomination
		baseAmount = amount.AmountOf(utils.BaseDenom)
	)

	cacheCtx, writeFn := ctx.CacheContext()

	// iterate through all delegations and get the rewards if any are unclaimed.
	// TODO: remove logging
	log.Print("iterating through delegations...")
	stakingKeeper.IterateDelegations(
		cacheCtx, addr, func(_ int64, delegation stakingtypes.DelegationI) (stop bool) {
			log.Printf("delegation: %s -> %s", delegation.GetDelegatorAddr().String(), delegation.GetValidatorAddr().String())
			reward, err = distributionKeeper.WithdrawDelegationRewards(cacheCtx, addr, delegation.GetValidatorAddr())
			if err != nil {
				log.Printf("error while withdrawing delegation rewards: %s", err)
				return true
			}
			log.Printf("reward: %s", reward.String())
			rewards = rewards.Add(reward...)

			return rewards.AmountOf(utils.BaseDenom).GTE(baseAmount)
		},
	)

	// check if there was an error while iterating delegations
	if err != nil {
		return fmt.Errorf("error while withdrawing delegation rewards: %s", err)
	}

	// only write to state if there are enough rewards to cover the transaction fees
	if rewards.AmountOf(utils.BaseDenom).LT(baseAmount) {
		return fmt.Errorf("insufficient staking rewards to cover transaction fees")
	}
	writeFn() // commit state changes
	return nil
}
