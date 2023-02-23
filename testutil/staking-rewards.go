package testutil

import (
	"fmt"
	"testing"

	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/staking"
	"github.com/cosmos/cosmos-sdk/x/staking/teststaking"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/evmos/evmos/v11/app"
	testutiltx "github.com/evmos/evmos/v11/testutil/tx"
	"github.com/evmos/evmos/v11/utils"
)

// PrepareAccountsForDelegationRewards prepares the test suite for testing to withdraw delegation rewards.
//
// The setup is done in the following way:
//   - Fund the account with the given address with the given balance.
//     If the given balance is zero, the account will be created with zero balance.
//   - Set up a validator with zero commission and delegate to it -> the account delegation will be 50% of the total delegation.
//   - Allocate rewards to the validator.
//
// The function returns the updated context along with a potential error.
func PrepareAccountsForDelegationRewards(t *testing.T, ctx sdk.Context, app *app.Evmos, addr sdk.AccAddress, balance, rewards sdkmath.Int) (sdk.Context, error) {
	if balance.IsZero() {
		app.AccountKeeper.SetAccount(ctx, app.AccountKeeper.NewAccountWithAddress(ctx, addr))
	} else {
		// Fund account with enough tokens to stake them
		err := FundAccountWithBaseDenom(ctx, app.BankKeeper, addr, balance.Int64())
		if err != nil {
			return sdk.Context{}, fmt.Errorf("failed to fund account: %s", err.Error())
		}
	}

	if !rewards.IsZero() {
		// reset historical count in distribution keeper which is necessary
		// for the delegation rewards to be calculated correctly
		app.DistrKeeper.DeleteAllValidatorHistoricalRewards(ctx)

		// Set up validator and delegate to it
		privKey := ed25519.GenPrivKey()
		addr2, _ := testutiltx.NewAccAddressAndKey()
		err := FundAccountWithBaseDenom(ctx, app.BankKeeper, addr2, rewards.Int64())
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
		stakingHelper.CreateValidator(valAddr, privKey.PubKey(), rewards, true)
		stakingHelper.Delegate(addr, valAddr, rewards)

		// TODO: Replace this with testutil.Commit?
		// end block to bond validator and increase block height
		staking.EndBlocker(ctx, app.StakingKeeper)
		ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 1)

		// set distribution module account balance which pays out the rewards
		distrAcc := app.DistrKeeper.GetDistributionAccount(ctx)
		err = FundModuleAccount(ctx, app.BankKeeper, distrAcc.GetName(), sdk.NewCoins(sdk.NewCoin(utils.BaseDenom, rewards)))
		if err != nil {
			return sdk.Context{}, fmt.Errorf("failed to fund distribution module account: %s", err.Error())
		}
		app.AccountKeeper.SetModuleAccount(ctx, distrAcc)

		// allocate rewards to validator (of these 50% will be paid out to the delegator)
		validator := app.StakingKeeper.Validator(ctx, valAddr)
		allocatedRewards := sdk.NewDecCoins(sdk.NewDecCoin(utils.BaseDenom, rewards.Mul(sdk.NewInt(2))))
		app.DistrKeeper.AllocateTokensToValidator(ctx, validator, allocatedRewards)
	}

	return ctx, nil
}
