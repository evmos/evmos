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

package testutil

import (
	"fmt"
	"testing"

	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	sdk "github.com/cosmos/cosmos-sdk/types"
	distributionkeeper "github.com/cosmos/cosmos-sdk/x/distribution/keeper"
	distributiontypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	"github.com/cosmos/cosmos-sdk/x/staking"
	"github.com/cosmos/cosmos-sdk/x/staking/teststaking"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/evmos/evmos/v11/app"
	testutiltx "github.com/evmos/evmos/v11/testutil/tx"
	"github.com/evmos/evmos/v11/utils"
)

// PrepareAccountsForDelegationRewards prepares the test suite for testing to withdraw delegation rewards.
//
// Balance is the amount of tokens that will be left in the account after the setup is done.
// For each defined reward, a validator is created and tokens are allocated to it using the distribution keeper,
// such that the given amount of tokens is outstanding as a staking reward for the account.
//
// The setup is done in the following way:
//   - Fund the account with the given address with the given balance.
//   - If the given balance is zero, the account will be created with zero balance.
//
// For every reward defined in the rewards argument, the following steps are executed:
//   - Set up a validator with zero commission and delegate to it -> the account delegation will be 50% of the total delegation.
//   - Allocate rewards to the validator.
//
// The function returns the updated context along with a potential error.
func PrepareAccountsForDelegationRewards(t *testing.T, ctx sdk.Context, app *app.Evmos, addr sdk.AccAddress, balance sdkmath.Int, rewards ...sdkmath.Int) (sdk.Context, error) {
	// Calculate the necessary amount of tokens to fund the account in order for the desired residual balance to
	// be left after creating validators and delegating to them.
	totalRewards := sdk.ZeroInt()
	for _, reward := range rewards {
		totalRewards = totalRewards.Add(reward)
	}
	totalNeededBalance := balance.Add(totalRewards)

	if totalNeededBalance.IsZero() {
		app.AccountKeeper.SetAccount(ctx, app.AccountKeeper.NewAccountWithAddress(ctx, addr))
	} else {
		// Fund account with enough tokens to stake them
		err := FundAccountWithBaseDenom(ctx, app.BankKeeper, addr, totalNeededBalance.Int64())
		if err != nil {
			return sdk.Context{}, fmt.Errorf("failed to fund account: %s", err.Error())
		}
	}

	if totalRewards.IsZero() {
		return ctx, nil
	}

	// reset historical count in distribution keeper which is necessary
	// for the delegation rewards to be calculated correctly
	app.DistrKeeper.DeleteAllValidatorHistoricalRewards(ctx)

	// set distribution module account balance which pays out the rewards
	distrAcc := app.DistrKeeper.GetDistributionAccount(ctx)
	err := FundModuleAccount(ctx, app.BankKeeper, distrAcc.GetName(), sdk.NewCoins(sdk.NewCoin(utils.BaseDenom, totalRewards)))
	if err != nil {
		return sdk.Context{}, fmt.Errorf("failed to fund distribution module account: %s", err.Error())
	}
	app.AccountKeeper.SetModuleAccount(ctx, distrAcc)

	for _, reward := range rewards {
		if reward.IsZero() {
			continue
		}

		// Set up validator and delegate to it
		privKey := ed25519.GenPrivKey()
		addr2, _ := testutiltx.NewAccAddressAndKey()
		err := FundAccountWithBaseDenom(ctx, app.BankKeeper, addr2, reward.Int64())
		if err != nil {
			return sdk.Context{}, fmt.Errorf("failed to fund validator account: %s", err.Error())
		}

		zeroDec := sdk.ZeroDec()
		stakingParams := app.StakingKeeper.GetParams(ctx)
		stakingParams.BondDenom = utils.BaseDenom
		stakingParams.MinCommissionRate = zeroDec
		app.StakingKeeper.SetParams(ctx, stakingParams)

		stakingHelper := teststaking.NewHelper(t, ctx, app.StakingKeeper)
		stakingHelper.Commission = stakingtypes.NewCommissionRates(zeroDec, zeroDec, zeroDec)
		stakingHelper.Denom = utils.BaseDenom

		valAddr := sdk.ValAddress(addr2.Bytes())
		// self-delegate the same amount of tokens as the delegate address also stakes
		// this ensures, that the delegation rewards are 50% of the total rewards
		stakingHelper.CreateValidator(valAddr, privKey.PubKey(), reward, true)
		stakingHelper.Delegate(addr, valAddr, reward)

		// end block to bond validator and increase block height
		// Not using Commit() here because code panics due to invalid block height
		staking.EndBlocker(ctx, app.StakingKeeper)

		// allocate rewards to validator (of these 50% will be paid out to the delegator)
		validator := app.StakingKeeper.Validator(ctx, valAddr)
		allocatedRewards := sdk.NewDecCoins(sdk.NewDecCoin(utils.BaseDenom, reward.Mul(sdk.NewInt(2))))
		app.DistrKeeper.AllocateTokensToValidator(ctx, validator, allocatedRewards)
	}

	return ctx, nil
}

// GetTotalDelegationRewards returns the total delegation rewards that are currently
// outstanding for the given address.
func GetTotalDelegationRewards(ctx sdk.Context, distributionKeeper distributionkeeper.Keeper, addr sdk.AccAddress) (sdk.DecCoins, error) {
	resp, err := distributionKeeper.DelegationTotalRewards(
		ctx,
		&distributiontypes.QueryDelegationTotalRewardsRequest{
			DelegatorAddress: addr.String(),
		},
	)
	if err != nil {
		return nil, err
	}

	return resp.Total, nil
}
