// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package testutil

import (
	"fmt"
	"testing"

	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	distributionkeeper "github.com/cosmos/cosmos-sdk/x/distribution/keeper"
	distributiontypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	"github.com/cosmos/cosmos-sdk/x/staking"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	"github.com/cosmos/cosmos-sdk/x/staking/teststaking"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/evmos/evmos/v12/app"
	testutiltx "github.com/evmos/evmos/v12/testutil/tx"
	"github.com/evmos/evmos/v12/utils"
)

// CreateValidator creates a validator with the provided public key and stake amount
func CreateValidator(ctx sdk.Context, t *testing.T, pubKey cryptotypes.PubKey, sk stakingkeeper.Keeper, stakeAmt sdkmath.Int) {
	zeroDec := sdk.ZeroDec()
	stakingParams := sk.GetParams(ctx)
	stakingParams.BondDenom = sk.BondDenom(ctx)
	stakingParams.MinCommissionRate = zeroDec
	sk.SetParams(ctx, stakingParams)

	stakingHelper := teststaking.NewHelper(t, ctx, sk)
	stakingHelper.Commission = stakingtypes.NewCommissionRates(zeroDec, zeroDec, zeroDec)
	stakingHelper.Denom = sk.BondDenom(ctx)

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
