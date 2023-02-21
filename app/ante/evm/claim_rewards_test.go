package evm_test

import (
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/staking"
	"github.com/cosmos/cosmos-sdk/x/staking/teststaking"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/evmos/evmos/v11/app/ante/evm"
	"github.com/evmos/evmos/v11/testutil"
	testutiltx "github.com/evmos/evmos/v11/testutil/tx"
	"github.com/evmos/evmos/v11/utils"
)

var (
	// define initial balance as sdk coins
	balanceAmount  = sdk.NewInt(1e18)
	initialBalance = sdk.Coins{sdk.Coin{Amount: balanceAmount, Denom: utils.BaseDenom}}

	// 5% commission
	fivePercent = sdk.NewDecWithPrec(5, 2)
)

// TestClaimSufficientStakingRewards tests the ClaimSufficientStakingRewards function
func (suite *AnteTestSuite) TestClaimSufficientStakingRewards() {
	// ----------------------------------------
	// Define testcases
	//
	testcases := []struct {
		name        string
		malleate    func(valAddr sdk.ValAddress)
		amount      int64
		expErr      bool
		errContains string
	}{
		{
			name: "pass - sufficient rewards can be claimed",
			malleate: func(valAddr sdk.ValAddress) {
				// set distribution module account balance
				balancePower := int64(1000)
				balanceTokens := suite.app.StakingKeeper.TokensFromConsensusPower(suite.ctx, balancePower)
				distrAcc := suite.app.DistrKeeper.GetDistributionAccount(suite.ctx)
				err := testutil.FundModuleAccount(
					suite.ctx, suite.app.BankKeeper, distrAcc.GetName(), sdk.NewCoins(sdk.NewCoin(utils.BaseDenom, balanceTokens)),
				)
				suite.Require().NoError(err, "failed to fund distribution module account")
				suite.app.AccountKeeper.SetModuleAccount(suite.ctx, distrAcc)

				// end block and increase block height
				staking.EndBlocker(suite.ctx, suite.app.StakingKeeper)
				suite.ctx = suite.ctx.WithBlockHeight(suite.ctx.BlockHeight() + 1)

				// allocate rewards to validator
				validator := suite.app.StakingKeeper.Validator(suite.ctx, valAddr)
				suite.app.DistrKeeper.AllocateTokensToValidator(suite.ctx, validator, sdk.NewDecCoins(sdk.NewDecCoin(utils.BaseDenom, sdk.NewInt(1000))))

				// check that the historical count is 3 (initial creation, delegation + reward allocation)
				historicalCount := suite.app.DistrKeeper.GetValidatorHistoricalReferenceCount(suite.ctx)
				suite.Require().Equal(uint64(3), historicalCount, "historical count should be 3; got %d", historicalCount)
			},
			amount:      100,
			expErr:      false,
			errContains: "",
		},
		{
			name:        "fail - no staking rewards to claim",
			malleate:    func(valAddr sdk.ValAddress) {},
			amount:      100,
			expErr:      true,
			errContains: "insufficient staking rewards to cover transaction fees",
		},
	}

	// ----------------------------------------
	// Run testcases
	//
	for _, tc := range testcases {
		suite.Run(tc.name, func() {
			suite.SetupTest()
			addr, valAddr := suite.BasicSetupForClaimRewardsTest()
			tc.malleate(valAddr)
			amount := sdk.NewCoins(sdk.NewCoin(utils.BaseDenom, sdk.NewInt(tc.amount)))

			// NOTE: this means that the delegation rewards are only available for the val address itself
			// seems like a self reward, that is done with the AllocateTokensToValidator function -> might need to use a different function
			rewards, err := suite.app.DistrKeeper.WithdrawDelegationRewards(suite.ctx, sdk.AccAddress(valAddr), valAddr)
			suite.Require().NoError(err, "failed to withdraw delegation rewards")
			suite.Require().Equal(sdk.NewCoins(sdk.NewCoin(utils.BaseDenom, sdk.NewInt(1230))), rewards, "expected rewards to be withdrawn")

			// TODO: remove logging
			suite.T().Logf("delegations: %v", suite.app.StakingKeeper.GetAllDelegatorDelegations(suite.ctx, addr))
			err = evm.ClaimSufficientStakingRewards(suite.ctx, suite.app.StakingKeeper, suite.app.DistrKeeper, addr, amount)

			if tc.expErr {
				suite.Require().Error(err)
				suite.Require().ErrorContains(err, tc.errContains)
			} else {
				suite.Require().NoError(err)
			}
		})
	}
}

// BasicSetupForClaimRewardsTest is a helper function, that creates a validator and a delegator
// and funds them with some tokens. It also sets up the staking keeper to include a self-delegation
// of the validator and a delegation from the delegator to the validator.
func (suite *AnteTestSuite) BasicSetupForClaimRewardsTest() (sdk.AccAddress, sdk.ValAddress) {
	// reset historical count
	suite.app.DistrKeeper.DeleteAllValidatorHistoricalRewards(suite.ctx)

	// ----------------------------------------
	// Set up first account
	//
	addr, _ := testutiltx.NewAccAddressAndKey()
	err := testutil.FundAccount(suite.ctx, suite.app.BankKeeper, addr, initialBalance)
	suite.Require().NoError(err, "failed to fund account")

	// ----------------------------------------
	// Set up validator
	//
	// This account gets funded with the same initial balance as the first account, which
	// will be fully used to self-delegate upon creation of the validator.
	//
	privKey := ed25519.GenPrivKey()
	addr2, _ := testutiltx.NewAccAddressAndKey()
	err = testutil.FundAccount(suite.ctx, suite.app.BankKeeper, addr2, initialBalance)
	suite.Require().NoError(err, "failed to fund account")

	stakingParams := suite.app.StakingKeeper.GetParams(suite.ctx)
	stakingParams.BondDenom = utils.BaseDenom
	suite.app.StakingKeeper.SetParams(suite.ctx, stakingParams)

	stakingHelper := teststaking.NewHelper(suite.T(), suite.ctx, suite.app.StakingKeeper)
	stakingHelper.Commission = stakingtypes.NewCommissionRates(fivePercent, fivePercent, fivePercent)
	stakingHelper.Denom = utils.BaseDenom

	valAddr := sdk.ValAddress(addr2.Bytes())
	//stakingHelper.CreateValidatorWithValPower(valAddr, privKey.PubKey(), int64(10), true)
	stakeAmount := suite.app.StakingKeeper.TokensFromConsensusPower(suite.ctx, int64(1))
	suite.T().Logf("stake amount: %s (1e%d)", stakeAmount.String(), len(stakeAmount.String())-1)
	stakingHelper.CreateValidator(valAddr, privKey.PubKey(), stakeAmount, true)
	stakingHelper.Delegate(addr, valAddr, sdk.NewInt(123456789))

	// end block to bond validator and increase block height
	staking.EndBlocker(suite.ctx, suite.app.StakingKeeper)
	suite.ctx = suite.ctx.WithBlockHeight(suite.ctx.BlockHeight() + 1)

	return addr, valAddr
}
