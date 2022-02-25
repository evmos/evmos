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
	ctx := sdk.WrapSDKContext(suite.ctx)

	req := &types.QueryUnvestedRequest{}
	addr := sdk.AccAddress(tests.GenerateAddress().Bytes())

	testCases := []struct {
		name     string
		malleate func()
		expErr   bool
	}{
		// {
		// 	"empty req", func() {}, true,
		// },
		// {
		// 	"invalid address",
		// 	func() {
		// 		req = &types.QueryUnvestedRequest{
		// 			Address: "evmos1",
		// 		}
		// 	},
		// 	true,
		// },
		// {
		// 	"uvnested Tokens not found for address",
		// 	func() {
		// 		req = &types.QueryUnvestedRequest{
		// 			Address: addr.String(),
		// 		}
		// 	},
		// 	true,
		// },
		// {
		// 	"valid, without funding",
		// 	func() {
		// 		// Create and fund periodic vesting account
		// 		vestingStart := s.ctx.BlockTime()
		// 		baseAccount := authtypes.NewBaseAccountWithAddress(addr)
		// 		funder := sdk.AccAddress(types.ModuleName)
		// 		clawbackAccount := types.NewClawbackVestingAccount(
		// 			baseAccount,
		// 			funder,
		// 			balances,
		// 			vestingStart,
		// 			lockupPeriods,
		// 			vestingPeriods,
		// 		)
		// 		acc := s.app.AccountKeeper.NewAccount(s.ctx, clawbackAccount)
		// 		s.app.AccountKeeper.SetAccount(s.ctx, acc)

		// 		req = &types.QueryUnvestedRequest{
		// 			Address: addr.String(),
		// 		}
		// 	},
		// 	false,
		// },
		{
			"valid",
			func() {
				// Create and fund periodic vesting account
				vestingStart := s.ctx.BlockTime()
				baseAccount := authtypes.NewBaseAccountWithAddress(addr)
				funder := sdk.AccAddress(types.ModuleName)
				clawbackAccount := types.NewClawbackVestingAccount(
					baseAccount,
					funder,
					balances,
					vestingStart,
					lockupPeriods,
					vestingPeriods,
				)
				err := testutil.FundAccount(s.app.BankKeeper, s.ctx, addr, balances)
				s.Require().NoError(err)
				acc := s.app.AccountKeeper.NewAccount(s.ctx, clawbackAccount)
				s.app.AccountKeeper.SetAccount(s.ctx, acc)

				req = &types.QueryUnvestedRequest{
					Address: addr.String(),
				}
			},
			false,
		},
		// {
		// 	"valid, non empty claimable amounts",
		// 	func() {
		// 		claimsRecord := types.NewUnvested(sdk.NewInt(1_000_000_000_000))
		// 		suite.app.ClaimsKeeper.SetUnvested(suite.ctx, addr, claimsRecord)
		// 		req = &types.QueryUnvestedRequest{
		// 			Address: addr.String(),
		// 		}
		// 	},
		// 	false,
		// },
	}

	for _, tc := range testCases {
		fmt.Println(addr.String())

		tc.malleate()

		res, err := suite.queryClient.Unvested(ctx, req)

		fmt.Println(fmt.Sprintf("%s\n", res.Unvested))

		if tc.expErr {
			suite.Require().Error(err)
		} else {
			suite.Require().NoError(err)
			suite.Require().Equal("STH", res.Unvested)
		}
	}
}
