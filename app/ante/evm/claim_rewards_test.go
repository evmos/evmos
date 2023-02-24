package evm_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	distributiontypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
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
		// testcase name
		name string
		// malleate sets up the test case specific state, i.e. delegations and assigning rewards
		malleate func(addr sdk.AccAddress)
		// amount specifies the necessary amount of rewards to be withdrawn
		amount int64
		// expErr defines whether the test case is expected to return an error
		expErr bool
		// expErrContains defines the error message that is expected to be returned
		errContains string
		// postCheck contains assertions that check the state after the test case has been executed
		// to further ensure that no false positives are reported
		postCheck func(addr sdk.AccAddress)
	}{
		{
			name: "pass - sufficient rewards can be withdrawn",
			malleate: func(addr sdk.AccAddress) {
				ctx, err := testutil.PrepareAccountsForDelegationRewards(
					suite.T(), suite.ctx, suite.app, addr, sdk.ZeroInt(), sdk.NewInt(1e18),
				)
				suite.Require().NoError(err, "failed to prepare accounts for delegation rewards")
				suite.ctx = ctx
			},
			amount: 1000,
			expErr: false,
			postCheck: func(addr sdk.AccAddress) {
				// Check that the necessary rewards are withdrawn, which means that there are no outstanding
				// rewards left
				resp, err := suite.app.DistrKeeper.DelegationTotalRewards(
					suite.ctx,
					&distributiontypes.QueryDelegationTotalRewardsRequest{DelegatorAddress: addr.String()},
				)
				suite.Require().NoError(err, "failed to query delegation total rewards")
				suite.Require().Empty(resp.Total, "expected no total rewards to be left")
			},
		},
		{
			name: "pass - Out of multiple outstanding rewards only those necessary are withdrawn",
			malleate: func(addr sdk.AccAddress) {
				ctx, err := testutil.PrepareAccountsForDelegationRewards(
					suite.T(), suite.ctx, suite.app, addr, sdk.ZeroInt(), sdk.NewInt(1e14), sdk.NewInt(2e14), sdk.NewInt(2e13),
				)
				suite.Require().NoError(err, "failed to prepare accounts for delegation rewards")
				suite.ctx = ctx
			},
			amount: 2e14,
			expErr: false,
			postCheck: func(addr sdk.AccAddress) {
				// Check that only the necessary rewards are withdrawn (=1e14+2e14), which means that there is an outstanding
				// reward of 2e13
				resp, err := suite.app.DistrKeeper.DelegationTotalRewards(
					suite.ctx,
					&distributiontypes.QueryDelegationTotalRewardsRequest{DelegatorAddress: addr.String()},
				)
				suite.Require().NoError(err, "failed to query delegation total rewards")
				suite.Require().NotNil(resp.Total, "expected rewards in one denomination yet to be withdrawn")
				suite.Require().Equal(1, len(resp.Total), "expected rewards in one denomination yet to be withdrawn")
				suite.Require().Equal(
					sdk.NewDecCoin(utils.BaseDenom, sdk.NewInt(2e13)),
					resp.Total[0],
					"expected total rewards with an amount of 2e15 yet to be withdrawn",
				)
			},
		},
		{
			name:        "fail - insufficient staking rewards to withdraw",
			malleate:    func(addr sdk.AccAddress) {},
			amount:      1000,
			expErr:      true,
			errContains: "insufficient staking rewards to cover transaction fees",
		},
		{
			name:     "pass - zero amount to be claimed",
			malleate: func(addr sdk.AccAddress) {},
			expErr:   false,
		},
	}

	// ----------------------------------------
	// Run testcases
	//
	for _, tc := range testcases {
		suite.Run(tc.name, func() {
			suite.SetupTest()
			addr, _ := testutiltx.NewAccAddressAndKey()
			tc.malleate(addr)

			amount := sdk.NewCoins(sdk.NewCoin(utils.BaseDenom, sdk.NewInt(tc.amount)))
			err := evm.ClaimSufficientStakingRewards(suite.ctx, suite.app.StakingKeeper, suite.app.DistrKeeper, addr, amount)

			if tc.expErr {
				suite.Require().Error(err)
				suite.Require().ErrorContains(err, tc.errContains)
			} else {
				suite.Require().NoError(err)
			}
			if tc.postCheck != nil {
				tc.postCheck(addr)
			}
		})
	}
}
