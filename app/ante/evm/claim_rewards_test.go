package evm_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	distributiontypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	"github.com/cosmos/cosmos-sdk/x/staking/teststaking"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/evmos/evmos/v11/app/ante/evm"
	"github.com/evmos/evmos/v11/testutil"
	testutiltx "github.com/evmos/evmos/v11/testutil/tx"
	"github.com/evmos/evmos/v11/utils"
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
			name: "pass - claim rewards",
			malleate: func(valAddr sdk.ValAddress) {
				rewards := distributiontypes.ValidatorOutstandingRewards{
					Rewards: sdk.DecCoins{sdk.NewDecCoinFromDec(utils.BaseDenom, sdk.NewDec(100))},
				}
				// FIXME: this doesn't work yet, the rewards are apparently zero, check how to correctly set them
				suite.app.DistrKeeper.SetValidatorOutstandingRewards(suite.ctx, valAddr, rewards)
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

			// TODO: remove logging
			suite.T().Logf("delegations: %v", suite.app.StakingKeeper.GetAllDelegatorDelegations(suite.ctx, addr))
			err := evm.ClaimSufficientStakingRewards(suite.ctx, suite.app.StakingKeeper, suite.app.DistrKeeper, addr, amount)

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
// and funds them with some tokens. It also sets up the staking and distribution keepers.
func (suite *AnteTestSuite) BasicSetupForClaimRewardsTest() (sdk.AccAddress, sdk.ValAddress) {
	balanceAmount := sdk.NewInt(1e18)
	initialBalance := sdk.Coins{sdk.Coin{Amount: balanceAmount, Denom: utils.BaseDenom}}
	fivePercent := sdk.NewDecWithPrec(5, 2)

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
	addr2, priv2 := testutiltx.NewAccAddressAndKey()
	err = testutil.FundAccount(suite.ctx, suite.app.BankKeeper, addr2, initialBalance)
	suite.Require().NoError(err, "failed to fund account")

	stakingParams := suite.app.StakingKeeper.GetParams(suite.ctx)
	stakingParams.BondDenom = utils.BaseDenom
	suite.app.StakingKeeper.SetParams(suite.ctx, stakingParams)

	stakingHelper := teststaking.NewHelper(suite.T(), suite.ctx, suite.app.StakingKeeper)
	stakingHelper.Commission = stakingtypes.NewCommissionRates(fivePercent, fivePercent, fivePercent)
	stakingHelper.Denom = utils.BaseDenom

	valAddr := sdk.ValAddress(addr2.Bytes())
	stakingHelper.CreateValidator(valAddr, priv2.PubKey(), balanceAmount, true)
	stakingHelper.Delegate(addr, valAddr, sdk.NewInt(123456789))

	// TODO: remove logging
	suite.T().Logf("Address 1: %s", addr.String())
	suite.T().Logf("Address 2: %s", addr2.String())
	suite.T().Logf("Val Address: %s", valAddr.String())

	return addr, valAddr
}
