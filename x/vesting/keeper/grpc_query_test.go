package keeper_test

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	"github.com/evmos/evmos/v16/testutil"
	"github.com/evmos/evmos/v16/x/vesting/types"
)

func (suite *KeeperTestSuite) TestBalances() {
	var (
		req    *types.QueryBalancesRequest
		expRes *types.QueryBalancesResponse
	)

	testCases := []struct {
		name        string
		malleate    func()
		expPass     bool
		errContains string
	}{
		{
			name: "nil req",
			malleate: func() {
				req = nil
			},
			expPass:     false,
			errContains: "empty address string is not allowed",
		},
		{
			name: "empty req",
			malleate: func() {
				req = &types.QueryBalancesRequest{}
			},
			expPass:     false,
			errContains: "empty address string is not allowed",
		},
		{
			name: "invalid address",
			malleate: func() {
				req = &types.QueryBalancesRequest{
					Address: "evmos1",
				}
			},
			expPass:     false,
			errContains: "decoding bech32 failed: invalid bech32 string length 6",
		},
		{
			name: "invalid account - not found",
			malleate: func() {
				req = &types.QueryBalancesRequest{
					Address: vestingAddr.String(),
				}
			},
			expPass:     false,
			errContains: "either does not exist or is not a vesting account",
		},
		{
			name: "invalid account - not clawback vesting account",
			malleate: func() {
				baseAccount := authtypes.NewBaseAccountWithAddress(vestingAddr)
				acc := suite.app.AccountKeeper.NewAccount(suite.ctx, baseAccount)
				suite.app.AccountKeeper.SetAccount(suite.ctx, acc)

				req = &types.QueryBalancesRequest{
					Address: vestingAddr.String(),
				}
			},
			expPass:     false,
			errContains: "either does not exist or is not a vesting account",
		},
		{
			name: "valid",
			malleate: func() {
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
			expPass: true,
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
				suite.Require().ErrorContains(err, tc.errContains)
			}
		})
	}
}
