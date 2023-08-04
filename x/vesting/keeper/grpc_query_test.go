package keeper_test

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	"github.com/evmos/evmos/v14/testutil"
	"github.com/evmos/evmos/v14/x/vesting/types"
)

func (suite *KeeperTestSuite) TestBalances() {
	var (
		req    *types.QueryBalancesRequest
		expRes *types.QueryBalancesResponse
	)

	testCases := []struct {
		name     string
		malleate func()
		expPass  bool
	}{
		{
			"empty req",
			func() {
				req = &types.QueryBalancesRequest{}
			},
			false,
		},
		{
			"invalid address",
			func() {
				req = &types.QueryBalancesRequest{
					Address: "evmos1",
				}
			},
			false,
		},
		{
			"invalid account - not found",
			func() {
				req = &types.QueryBalancesRequest{
					Address: vestingAddr.String(),
				}
			},
			false,
		},
		{
			"invalid account - not clawback vesting account",
			func() {
				baseAccount := authtypes.NewBaseAccountWithAddress(vestingAddr)
				acc := suite.app.AccountKeeper.NewAccount(suite.ctx, baseAccount)
				suite.app.AccountKeeper.SetAccount(suite.ctx, acc)

				req = &types.QueryBalancesRequest{
					Address: vestingAddr.String(),
				}
			},
			false,
		},
		{
			"valid",
			func() {
				vestingStart := s.ctx.BlockTime()

				// fund the vesting account with coins to initialize it and
				// then send all balances to the funding account
				err = testutil.FundAccount(suite.ctx, suite.app.BankKeeper, vestingAddr, balances)
				suite.Require().NoError(err, "error while funding the target account")
				err = s.app.BankKeeper.SendCoins(suite.ctx, vestingAddr, funder, balances)
				suite.Require().NoError(err, "error while sending coins to the funder account")

				msg := types.NewMsgCreateClawbackVestingAccount(
					funder,
					vestingAddr,
					false,
				)
				_, err = suite.app.VestingKeeper.CreateClawbackVestingAccount(sdk.WrapSDKContext(suite.ctx), msg)
				suite.Require().NoError(err, "error while creating the vesting account")

				msgFund := types.NewMsgFundVestingAccount(
					funder,
					vestingAddr,
					vestingStart,
					lockupPeriods,
					vestingPeriods,
				)
				_, err = suite.app.VestingKeeper.FundVestingAccount(sdk.WrapSDKContext(suite.ctx), msgFund)
				suite.Require().NoError(err, "error while funding the vesting account")

				req = &types.QueryBalancesRequest{
					Address: vestingAddr.String(),
				}
				expRes = &types.QueryBalancesResponse{
					Locked:   balances,
					Unvested: balances,
					Vested:   nil,
				}
			},
			true,
		},
	}

	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			suite.SetupTest() // reset
			ctx := sdk.WrapSDKContext(suite.ctx)
			tc.malleate()
			suite.Commit()

			res, err := suite.queryClient.Balances(ctx, req)
			if tc.expPass {
				suite.Require().NoError(err)
				suite.Require().Equal(expRes, res)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}
