package keeper_test

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	"github.com/tharsis/ethermint/tests"
	"github.com/tharsis/evmos/testutil"
	"github.com/tharsis/evmos/x/vesting/types"
)

func (suite *KeeperTestSuite) TestUnvested() {
	var (
		req    *types.QueryUnvestedRequest
		expRes *types.QueryUnvestedResponse
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
				req = &types.QueryUnvestedRequest{}
			},
			false,
		},
		{
			"invalid address",
			func() {
				req = &types.QueryUnvestedRequest{
					Address: "evmos1",
				}
			},
			false,
		},
		{
			"invalid account - not found",
			func() {
				req = &types.QueryUnvestedRequest{
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

				req = &types.QueryUnvestedRequest{
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

				req = &types.QueryUnvestedRequest{
					Address: addr.String(),
				}
				expRes = &types.QueryUnvestedResponse{Unvested: balances}
			},
			true,
		},
	}

	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			suite.SetupTest() // reset
			ctx := sdk.WrapSDKContext(suite.ctx)
			tc.malleate()

			res, err := suite.queryClient.Unvested(ctx, req)
			if tc.expPass {
				suite.Require().NoError(err)
				suite.Require().Equal(expRes, res)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestVested() {
	var (
		req    *types.QueryVestedRequest
		expRes *types.QueryVestedResponse
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
				req = &types.QueryVestedRequest{}
			},
			false,
		},
		{
			"invalid address",
			func() {
				req = &types.QueryVestedRequest{
					Address: "evmos1",
				}
			},
			false,
		},
		{
			"invalid account - not found",
			func() {
				req = &types.QueryVestedRequest{
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

				req = &types.QueryVestedRequest{
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

				req = &types.QueryVestedRequest{
					Address: addr.String(),
				}
				expRes = &types.QueryVestedResponse{}
			},
			true,
		},
	}

	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			suite.SetupTest() // reset
			ctx := sdk.WrapSDKContext(suite.ctx)
			tc.malleate()

			res, err := suite.queryClient.Vested(ctx, req)
			if tc.expPass {
				suite.Require().NoError(err)
				suite.Require().Equal(expRes, res)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestLocked() {
	var (
		req    *types.QueryLockedRequest
		expRes *types.QueryLockedResponse
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
				req = &types.QueryLockedRequest{}
			},
			false,
		},
		{
			"invalid address",
			func() {
				req = &types.QueryLockedRequest{
					Address: "evmos1",
				}
			},
			false,
		},
		{
			"invalid account - not found",
			func() {
				req = &types.QueryLockedRequest{
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

				req = &types.QueryLockedRequest{
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

				req = &types.QueryLockedRequest{
					Address: addr.String(),
				}
				expRes = &types.QueryLockedResponse{Locked: balances}
			},
			true,
		},
	}

	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			suite.SetupTest() // reset
			ctx := sdk.WrapSDKContext(suite.ctx)
			tc.malleate()

			res, err := suite.queryClient.Locked(ctx, req)
			if tc.expPass {
				suite.Require().NoError(err)
				suite.Require().Equal(expRes, res)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}
