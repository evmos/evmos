package evm_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/evmos/evmos/v11/app/ante/evm"
	"github.com/evmos/evmos/v11/testutil"
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
			name: "pass - sufficient rewards can be claimed",
			malleate: func(valAddr sdk.ValAddress) {
				// set distribution module account balance which pays out the rewards
				distrAcc := suite.app.DistrKeeper.GetDistributionAccount(suite.ctx)
				err := testutil.FundModuleAccount(
					suite.ctx, suite.app.BankKeeper, distrAcc.GetName(), sdk.NewCoins(sdk.NewCoin(utils.BaseDenom, sdk.NewInt(1e18))),
				)
				suite.Require().NoError(err, "failed to fund distribution module account")
				suite.app.AccountKeeper.SetModuleAccount(suite.ctx, distrAcc)

				// allocate rewards to validator
				validator := suite.app.StakingKeeper.Validator(suite.ctx, valAddr)
				allocatedRewards := sdk.NewDecCoins(sdk.NewDecCoin(utils.BaseDenom, sdk.NewInt(1e18)))
				suite.app.DistrKeeper.AllocateTokensToValidator(suite.ctx, validator, allocatedRewards)

				// check that the validator historical reference count is 3 (initial creation, delegation + reward allocation)
				historicalCount := suite.app.DistrKeeper.GetValidatorHistoricalReferenceCount(suite.ctx)
				suite.Require().Equal(uint64(3), historicalCount, "historical count should be 3; got %d", historicalCount)
			},
			amount: 1000,
			expErr: false,
		},
		{
			name:        "fail - insufficient staking rewards to claim",
			malleate:    func(valAddr sdk.ValAddress) {},
			amount:      1000,
			expErr:      true,
			errContains: "insufficient staking rewards to cover transaction fees",
		},
		{
			name:     "pass - zero amount to be claimed",
			malleate: func(valAddr sdk.ValAddress) {},
			expErr:   false,
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
