package keeper_test

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	"github.com/evmos/ethermint/tests"
	"github.com/evmos/evmos/v6/testutil"
	"github.com/evmos/evmos/v6/x/vesting/types"
)

func (suite *KeeperTestSuite) TestBalances() {
	var (
		req    *types.QueryBalancesRequest
		expRes *types.QueryBalancesResponse
	)
	addr := sdk.AccAddress(tests.GenerateAddress().Bytes())

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
					Address: addr.String(),
				}
			},
			false,
		},
		{
			"invalid account - not clawback vesting account",
			func() {
				baseAccount := authtypes.NewBaseAccountWithAddress(addr)
				acc := suite.app.AccountKeeper.NewAccount(suite.ctx, baseAccount)
				suite.app.AccountKeeper.SetAccount(suite.ctx, acc)

				req = &types.QueryBalancesRequest{
					Address: addr.String(),
				}
			},
			false,
		},
		{
			"valid",
			func() {
				vestingStart := s.ctx.BlockTime()
				funder := sdk.AccAddress(types.ModuleName)
				err := testutil.FundAccount(suite.app.BankKeeper, suite.ctx, funder, balances)
				suite.Require().NoError(err)

				msg := types.NewMsgCreateClawbackVestingAccount(
					funder,
					addr,
					vestingStart,
					lockupPeriods,
					vestingPeriods,
					false,
				)
				ctx := sdk.WrapSDKContext(suite.ctx)
				_, err = suite.app.VestingKeeper.CreateClawbackVestingAccount(ctx, msg)
				suite.Require().NoError(err)

				req = &types.QueryBalancesRequest{
					Address: addr.String(),
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
