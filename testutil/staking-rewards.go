// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package testutil

import (
	"fmt"
	"testing"

	"cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	distributionkeeper "github.com/cosmos/cosmos-sdk/x/distribution/keeper"
	distributiontypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	teststaking "github.com/cosmos/cosmos-sdk/x/staking/testutil"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/evmos/evmos/v20/app"
	testutiltx "github.com/evmos/evmos/v20/testutil/tx"
	evmostypes "github.com/evmos/evmos/v20/types"
	"github.com/stretchr/testify/require"
)

// CreateValidator creates a validator with the provided public key and stake amount
func CreateValidator(ctx sdk.Context, t *testing.T, pubKey cryptotypes.PubKey, sk stakingkeeper.Keeper, stakeAmt math.Int) {
	zeroDec := math.LegacyZeroDec()
	stakingParams, err := sk.GetParams(ctx)
	require.NoError(t, err)
	stakingParams.BondDenom, err = sk.BondDenom(ctx)
	require.NoError(t, err)
	stakingParams.MinCommissionRate = zeroDec
	err = sk.SetParams(ctx, stakingParams)
	require.NoError(t, err)

	stakingHelper := teststaking.NewHelper(t, ctx, &sk)
	stakingHelper.Commission = stakingtypes.NewCommissionRates(zeroDec, zeroDec, zeroDec)
	stakingHelper.Denom, err = sk.BondDenom(ctx)
	require.NoError(t, err)

	valAddr := sdk.ValAddress(pubKey.Address())
	stakingHelper.CreateValidator(valAddr, pubKey, stakeAmt, true)
}

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
func PrepareAccountsForDelegationRewards(t *testing.T, ctx sdk.Context, app *app.Evmos, addr sdk.AccAddress, balance math.Int, rewards ...math.Int) (sdk.Context, error) {
	// Calculate the necessary amount of tokens to fund the account in order for the desired residual balance to
	// be left after creating validators and delegating to them.
	totalRewards := math.ZeroInt()
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
	err := FundModuleAccount(ctx, app.BankKeeper, distrAcc.GetName(), sdk.NewCoins(sdk.NewCoin(evmostypes.BaseDenom, totalRewards)))
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

		zeroDec := math.LegacyZeroDec()
		stakingParams, err := app.StakingKeeper.GetParams(ctx)
		if err != nil {
			return sdk.Context{}, fmt.Errorf("failed to get staking params: %s", err.Error())
		}
		stakingParams.BondDenom = evmostypes.BaseDenom
		stakingParams.MinCommissionRate = zeroDec
		err = app.StakingKeeper.SetParams(ctx, stakingParams)
		require.NoError(t, err)

		stakingHelper := teststaking.NewHelper(t, ctx, app.StakingKeeper.Keeper)
		stakingHelper.Commission = stakingtypes.NewCommissionRates(zeroDec, zeroDec, zeroDec)
		stakingHelper.Denom = stakingParams.BondDenom

		valAddr := sdk.ValAddress(addr2.Bytes())
		// self-delegate the same amount of tokens as the delegate address also stakes
		// this ensures, that the delegation rewards are 50% of the total rewards
		stakingHelper.CreateValidator(valAddr, privKey.PubKey(), reward, true)
		stakingHelper.Delegate(addr, valAddr, reward)

		// end block to bond validator and increase block height
		// Not using Commit() here because code panics due to invalid block height
		_, err = app.StakingKeeper.EndBlocker(ctx)
		require.NoError(t, err)

		// allocate rewards to validator (of these 50% will be paid out to the delegator)
		validator, err := app.StakingKeeper.Validator(ctx, valAddr)
		if err != nil {
			return sdk.Context{}, fmt.Errorf("failed to get validator: %s", err.Error())
		}
		allocatedRewards := sdk.NewDecCoins(sdk.NewDecCoin(stakingParams.BondDenom, reward.Mul(math.NewInt(2))))
		if err = app.DistrKeeper.AllocateTokensToValidator(ctx, validator, allocatedRewards); err != nil {
			return sdk.Context{}, fmt.Errorf("failed to allocate tokens to validator: %s", err.Error())
		}
	}

	// Increase block height in ctx for the rewards calculation
	// NOTE: this will only work for unit tests that use the context
	// returned by this function
	currentHeight := ctx.BlockHeight()
	return ctx.WithBlockHeight(currentHeight + 1), nil
}

// GetTotalDelegationRewards returns the total delegation rewards that are currently
// outstanding for the given address.
func GetTotalDelegationRewards(ctx sdk.Context, distributionKeeper distributionkeeper.Keeper, addr sdk.AccAddress) (sdk.DecCoins, error) {
	querier := distributionkeeper.NewQuerier(distributionKeeper)
	resp, err := querier.DelegationTotalRewards(
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
