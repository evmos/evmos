package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"

	"github.com/tharsis/ethermint/tests"
	"github.com/tharsis/evmos/x/claims/types"
)

func (suite *KeeperTestSuite) TestTotalUnclaimed() {
	ctx := sdk.WrapSDKContext(suite.ctx)
	coins := sdk.NewCoins(sdk.NewCoin("aevmos", sdk.NewInt(1000)))

	testCases := []struct {
		name       string
		malleate   func()
		expBalance sdk.Coins
	}{
		{
			"empty balance", func() {}, sdk.Coins(nil),
		},
		{
			"non-empty balance",
			func() {
				err := suite.app.BankKeeper.MintCoins(suite.ctx, minttypes.ModuleName, coins)
				suite.Require().NoError(err)
				err = suite.app.BankKeeper.SendCoinsFromModuleToModule(suite.ctx, minttypes.ModuleName, types.ModuleName, coins)
				suite.Require().NoError(err)
			}, coins,
		},
	}

	for _, tc := range testCases {

		tc.malleate()

		res, err := suite.queryClient.TotalUnclaimed(ctx, &types.QueryTotalUnclaimedRequest{})
		suite.Require().NoError(err)
		suite.Require().Equal(tc.expBalance, res.Coins)
	}
}

func (suite *KeeperTestSuite) TestQueryParams() {
	ctx := sdk.WrapSDKContext(suite.ctx)
	expParams := types.DefaultParams()
	expParams.AirdropStartTime = suite.ctx.BlockTime()

	res, err := suite.queryClient.Params(ctx, &types.QueryParamsRequest{})
	suite.Require().NoError(err)
	suite.Require().Equal(expParams, res.Params)
}

func (suite *KeeperTestSuite) TestClaimsRecords() {
	ctx := sdk.WrapSDKContext(suite.ctx)

	addr := sdk.AccAddress(tests.GenerateAddress().Bytes())

	testCases := []struct {
		name          string
		malleate      func()
		expErr        bool
		recordsAmount int
		initialAmount sdk.Int
		actions       []bool
	}{
		{
			"no values", func() {}, false, 0, sdk.ZeroInt(), []bool{},
		},
		{
			"valid, all zero",
			func() {
				claimRecord := types.NewClaimRecord(sdk.ZeroInt())
				suite.app.ClaimsKeeper.SetClaimRecord(suite.ctx, addr, claimRecord)
			},
			false,
			1,
			sdk.ZeroInt(),
			[]bool{false, false, false, false},
		},
		{
			"valid, non empty claimable amounts",
			func() {
				claimRecord := types.NewClaimRecord(sdk.NewInt(1_000_000_000_000))
				suite.app.ClaimsKeeper.SetClaimRecord(suite.ctx, addr, claimRecord)
			},
			false,
			1,
			sdk.NewInt(1_000_000_000_000),
			[]bool{false, false, false, false},
		},
		{
			"valid, half complete half incomplete",
			func() {
				claimRecord := types.NewClaimRecord(sdk.NewInt(1_000_000_000_000))
				claimRecord.ActionsCompleted = []bool{false, false, true, true}
				suite.app.ClaimsKeeper.SetClaimRecord(suite.ctx, addr, claimRecord)
			},
			false,
			1,
			sdk.NewInt(1_000_000_000_000),
			[]bool{false, false, true, true},
		},
	}

	for _, tc := range testCases {

		tc.malleate()

		res, err := suite.queryClient.ClaimRecords(ctx, &types.QueryClaimRecordsRequest{})
		if tc.expErr {
			suite.Require().Error(err)
		} else {
			if tc.recordsAmount == 0 {
				suite.Require().NoError(err)
			} else if tc.recordsAmount == 1 {
				suite.Require().NoError(err)
				suite.Require().Len(res.Claims, 1)
				suite.Require().Equal(res.Claims[0].Address, addr.String())
				suite.Require().Len(res.Claims[0].ActionsCompleted, 4)
				for i, claim := range res.Claims[0].ActionsCompleted {
					suite.Require().Equal(claim, tc.actions[i])

				}
				suite.Require().Equal(res.Claims[0].InitialClaimableAmount.String(), tc.initialAmount.String())

			} else {
				// The test should never reach here
				suite.Require().Equal(true, false)
			}
		}
	}
}

func (suite *KeeperTestSuite) TestClaimRecords() {
	ctx := sdk.WrapSDKContext(suite.ctx)

	req := &types.QueryClaimRecordRequest{}
	addr := sdk.AccAddress(tests.GenerateAddress().Bytes())

	testCases := []struct {
		name     string
		malleate func()
		expErr   bool
	}{
		{
			"empty req", func() {}, true,
		},
		{
			"invalid address",
			func() {
				req = &types.QueryClaimRecordRequest{
					Address: "evmos1",
				}
			},
			true,
		},
		{
			"claim record not found for address",
			func() {
				req = &types.QueryClaimRecordRequest{
					Address: addr.String(),
				}
			},
			true,
		},
		{
			"valid, all zero",
			func() {
				claimRecord := types.NewClaimRecord(sdk.ZeroInt())
				suite.app.ClaimsKeeper.SetClaimRecord(suite.ctx, addr, claimRecord)
				req = &types.QueryClaimRecordRequest{
					Address: addr.String(),
				}
			},
			false,
		},
		{
			"valid, non empty claimable amounts",
			func() {
				claimRecord := types.NewClaimRecord(sdk.NewInt(1_000_000_000_000))
				suite.app.ClaimsKeeper.SetClaimRecord(suite.ctx, addr, claimRecord)
				req = &types.QueryClaimRecordRequest{
					Address: addr.String(),
				}
			},
			false,
		},
	}

	for _, tc := range testCases {

		tc.malleate()

		res, err := suite.queryClient.ClaimRecord(ctx, req)
		if tc.expErr {
			suite.Require().Error(err)
		} else {
			suite.Require().NoError(err)
			suite.Require().Len(res.Claims, 4)
			for _, claim := range res.Claims {
				suite.Require().Equal(res.InitialClaimableAmount.QuoRaw(4).String(), claim.ClaimableAmount.String())
			}
		}
	}
}
