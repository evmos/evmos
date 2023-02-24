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
				// NOTE: To enable executing the post check in a deterministic way, we only test with two
				// assigned rewards, of which one is sufficient to cover the transaction fees and the other
				// is not. This is because the iteration over rewards is done in a non-deterministic fashion,
				// This means, that e.g. if reward C is sufficient, but A and B are not,
				// all of the options [A], [B-A], [B-C-A] or [C-A] are possible to be withdrawn, which
				// increases the complexity of assertions.
				ctx, err := testutil.PrepareAccountsForDelegationRewards(
					suite.T(), suite.ctx, suite.app, addr, sdk.ZeroInt(), sdk.NewInt(1e14), sdk.NewInt(2e14),
				)
				suite.Require().NoError(err, "failed to prepare accounts for delegation rewards")
				suite.ctx = ctx
			},
			amount: 2e14,
			expErr: false,
			postCheck: func(addr sdk.AccAddress) {
				balance := suite.app.BankKeeper.GetBalance(suite.ctx, addr, utils.BaseDenom)
				resp, err := suite.app.DistrKeeper.DelegationTotalRewards(
					suite.ctx,
					&distributiontypes.QueryDelegationTotalRewardsRequest{DelegatorAddress: addr.String()},
				)
				suite.Require().NoError(err, "failed to query delegation total rewards")

				// NOTE: The only valid options (because of the non-deterministic iteration over rewards) are a
				// balance of 2e14 (only withdraw reward B) or 3e14 (A+B), which is why we check for both of them.
				switch {
				case balance.Amount.Equal(sdk.NewInt(2e14)):
					suite.Require().NotNil(resp.Total, "expected rewards in one denomination yet to be withdrawn")
					suite.Require().Equal(1, len(resp.Total), "expected rewards in one denomination yet to be withdrawn")
					suite.Require().Equal(
						sdk.NewDecCoin(utils.BaseDenom, sdk.NewInt(1e14)),
						resp.Total[0],
						"expected total rewards with an amount of 1e14 yet to be withdrawn",
					)
				case balance.Amount.Equal(sdk.NewInt(3e14)):
					suite.Require().Nil(resp.Total, "expected no rewards to be left to withdraw")
				default:
					suite.Require().Fail("unexpected balance", "balance: %v", balance)
				}
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
