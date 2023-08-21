package utils_test

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	anteutils "github.com/evmos/evmos/v14/app/ante/utils"
	"github.com/evmos/evmos/v14/testutil"
	testutiltx "github.com/evmos/evmos/v14/testutil/tx"
	"github.com/evmos/evmos/v14/utils"
)

// TestClaimStakingRewardsIfNecessary tests the ClaimStakingRewardsIfNecessary function
func (suite *AnteTestSuite) TestClaimStakingRewardsIfNecessary() {
	testcases := []struct {
		// testcase name
		name string
		// malleate sets up the test case specific state, i.e. delegations and assigning rewards
		malleate func(addr sdk.AccAddress)
		// amount specifies the necessary coins to be withdrawn
		amount sdk.Coins
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
				var err error
				suite.ctx, err = testutil.PrepareAccountsForDelegationRewards(
					suite.T(), suite.ctx, suite.app, addr, sdk.ZeroInt(), sdk.NewInt(1e18),
				)
				suite.Require().NoError(err, "failed to prepare accounts for delegation rewards")
				suite.ctx, err = testutil.Commit(suite.ctx, suite.app, time.Second*0, nil)
				suite.Require().NoError(err)
			},
			amount: sdk.Coins{sdk.Coin{Denom: utils.BaseDenom, Amount: sdk.NewInt(1000)}},
			expErr: false,
			postCheck: func(addr sdk.AccAddress) {
				// Check that the necessary rewards are withdrawn, which means that there are no outstanding
				// rewards left
				rewards, err := testutil.GetTotalDelegationRewards(suite.ctx, suite.app.DistrKeeper, addr)
				suite.Require().NoError(err, "failed to query delegation total rewards")
				suite.Require().Empty(rewards, "expected no total rewards to be left")
			},
		},
		{
			name: "pass - out of multiple outstanding rewards only those necessary are withdrawn",
			malleate: func(addr sdk.AccAddress) {
				// NOTE: To enable executing the post check in a deterministic way, we only test with two
				// assigned rewards, of which one is sufficient to cover the transaction fees and the other
				// is not. This is because the iteration over rewards is done in a non-deterministic fashion,
				// This means, that e.g. if reward C is sufficient, but A and B are not,
				// all of the options [A], [B-A], [B-C-A] or [C-A] are possible to be withdrawn, which
				// increases the complexity of assertions.
				var err error
				suite.ctx, err = testutil.PrepareAccountsForDelegationRewards(
					suite.T(), suite.ctx, suite.app, addr, sdk.ZeroInt(), sdk.NewInt(1e14), sdk.NewInt(2e14),
				)
				suite.Require().NoError(err, "failed to prepare accounts for delegation rewards")
				suite.ctx, err = testutil.Commit(suite.ctx, suite.app, time.Second*0, nil)
				suite.Require().NoError(err)
			},
			amount: sdk.Coins{sdk.Coin{Denom: utils.BaseDenom, Amount: sdk.NewInt(2e14)}},
			expErr: false,
			postCheck: func(addr sdk.AccAddress) {
				balance := suite.app.BankKeeper.GetBalance(suite.ctx, addr, utils.BaseDenom)
				rewards, err := testutil.GetTotalDelegationRewards(suite.ctx, suite.app.DistrKeeper, addr)
				suite.Require().NoError(err, "failed to query delegation total rewards")

				// NOTE: The only valid options (because of the non-deterministic iteration over rewards, see comment above)
				// are a balance of 2e14 (only withdraw reward B) or 3e14 (A+B), which is why we check for both of them.
				// Any other balance fails the test.
				switch {
				case balance.Amount.Equal(sdk.NewInt(2e14)):
					suite.Require().Equal(
						sdk.DecCoins{sdk.DecCoin{Denom: utils.BaseDenom, Amount: sdk.NewDec(1e14)}},
						rewards,
						"expected total rewards with an amount of 1e14 yet to be withdrawn",
					)
				case balance.Amount.Equal(sdk.NewInt(3e14)):
					suite.Require().Empty(rewards, "expected no rewards left to withdraw")
				default:
					suite.Require().Fail("unexpected balance", "balance: %v", balance)
				}
			},
		},
		{
			name: "pass - user has enough balance to cover transaction fees",
			malleate: func(addr sdk.AccAddress) {
				var err error
				suite.ctx, err = testutil.PrepareAccountsForDelegationRewards(
					suite.T(), suite.ctx, suite.app, addr, sdk.NewInt(1e15), sdk.NewInt(1e18),
				)
				suite.Require().NoError(err, "failed to prepare accounts for delegation rewards")
				suite.ctx, err = testutil.Commit(suite.ctx, suite.app, time.Second*0, nil)
				suite.Require().NoError(err)
			},
			amount: sdk.Coins{sdk.Coin{Denom: utils.BaseDenom, Amount: sdk.NewInt(1000)}},
			expErr: false,
			postCheck: func(addr sdk.AccAddress) {
				// balance should be unchanged as no rewards should have been withdrawn
				balance := suite.app.BankKeeper.GetBalance(suite.ctx, addr, utils.BaseDenom)
				suite.Require().Equal(sdk.NewInt(1e15), balance.Amount, "expected balance to be unchanged")

				// No rewards should be withdrawn
				rewards, err := testutil.GetTotalDelegationRewards(suite.ctx, suite.app.DistrKeeper, addr)
				suite.Require().NoError(err, "failed to query delegation total rewards")
				suite.Require().Equal(
					sdk.DecCoins{sdk.DecCoin{Denom: utils.BaseDenom, Amount: sdk.NewDec(1e18)}},
					rewards,
					"expected total rewards with an amount of 1e18 yet to be withdrawn",
				)
			},
		},
		{
			name:        "fail - insufficient staking rewards to withdraw",
			malleate:    func(addr sdk.AccAddress) {},
			amount:      sdk.Coins{sdk.Coin{Denom: utils.BaseDenom, Amount: sdk.NewInt(1000)}},
			expErr:      true,
			errContains: "insufficient staking rewards to cover transaction fees",
		},
		{
			name:     "pass - zero amount to be claimed",
			malleate: func(addr sdk.AccAddress) {},
			amount:   sdk.Coins{sdk.Coin{Denom: utils.BaseDenom, Amount: sdk.ZeroInt()}},
			expErr:   false,
		},
		{
			name:        "fail - wrong coin denom",
			malleate:    func(addr sdk.AccAddress) {},
			amount:      sdk.Coins{sdk.Coin{Denom: "wrongCoin", Amount: sdk.NewInt(1000)}},
			expErr:      true,
			errContains: "wrong fee denomination",
		},
	}

	for _, tc := range testcases {
		suite.Run(tc.name, func() {
			suite.SetupTest()
			addr, _ := testutiltx.NewAccAddressAndKey()
			tc.malleate(addr)

			err := anteutils.ClaimStakingRewardsIfNecessary(suite.ctx, suite.app.BankKeeper, suite.app.DistrKeeper, suite.app.StakingKeeper, addr, tc.amount)

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
