package keeper_test

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	"github.com/tharsis/evmos/x/incentives/types"
)

func (suite KeeperTestSuite) TestRegisterIncentive() {
	testCases := []struct {
		name     string
		malleate func()
		expPass  bool
	}{
		{
			"incentives are disabled globally",
			func() {
				params := types.DefaultParams()
				params.EnableIncentives = false
				suite.app.IncentivesKeeper.SetParams(suite.ctx, params)
			},
			false,
		},
		{
			"inventive already registered",
			func() {
				regIn := types.NewIncentive(contract, allocations, epochs)
				suite.app.IncentivesKeeper.SetIncentive(suite.ctx, regIn)
				suite.Commit()
			},
			false,
		},
		{
			"coin doesn't have supply",
			func() {
			},
			false,
		},
		{
			"allocation above allocation limit",
			func() {
				// Make sure the non-mint coin has supply
				err := suite.app.BankKeeper.MintCoins(
					suite.ctx,
					minttypes.ModuleName,
					sdk.Coins{sdk.NewInt64Coin(denomCoin, 1)},
				)
				suite.Require().NoError(err)

				// decrease allocation limit
				params := types.DefaultParams()
				params.AllocationLimit = sdk.NewDecWithPrec(1, 2)
				suite.app.IncentivesKeeper.SetParams(suite.ctx, params)
			},
			false,
		},
		{
			"Total allocation for at least one denom (current + proposed) > 100%",
			func() {
				// Make sure the non-mint coin has supply
				err := suite.app.BankKeeper.MintCoins(
					suite.ctx,
					minttypes.ModuleName,
					sdk.Coins{sdk.NewInt64Coin(denomCoin, 1)},
				)
				suite.Require().NoError(err)

				// increase allocation limit
				params := types.DefaultParams()
				params.AllocationLimit = sdk.NewDecWithPrec(100, 2)
				suite.app.IncentivesKeeper.SetParams(suite.ctx, params)

				// Add incentive which takes up 100% of the allocation
				_, err = suite.app.IncentivesKeeper.RegisterIncentive(
					suite.ctx,
					contract2,
					sdk.DecCoins{
						sdk.NewDecCoinFromDec(denomCoin, sdk.NewDecWithPrec(100, 2)),
					},
					epochs,
				)
				suite.Require().NoError(err)
				suite.Commit()
			},
			false,
		},
		{
			"ok",
			func() {
				// Make sure the non-mint coin has supply
				err := suite.app.BankKeeper.MintCoins(
					suite.ctx,
					minttypes.ModuleName,
					sdk.Coins{sdk.NewInt64Coin(denomCoin, 1)},
				)
				suite.Require().NoError(err)
			},
			true,
		},
	}
	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			suite.SetupTest() // reset

			tc.malleate()

			in, err := suite.app.IncentivesKeeper.RegisterIncentive(
				suite.ctx,
				contract,
				allocations,
				epochs,
			)
			suite.Commit()

			expIn := &types.Incentive{
				Contract:    contract.String(),
				Allocations: allocations,
				Epochs:      epochs,
			}

			if tc.expPass {
				suite.Require().NoError(err, tc.name)
				suite.Require().Equal(expIn, in)
			} else {
				suite.Require().Error(err, tc.name)
			}
		})
	}
}

func (suite KeeperTestSuite) TestCancelIncentive() {
	testCases := []struct {
		name     string
		malleate func()
		expPass  bool
	}{
		{
			"inventive not registered",
			func() {
			},
			false,
		},
		{
			"ok",
			func() {
				_, err := suite.app.IncentivesKeeper.RegisterIncentive(
					suite.ctx,
					contract,
					mintAllocations,
					epochs,
				)
				suite.Require().NoError(err)
				suite.Commit()
			},
			true,
		},
	}
	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			suite.SetupTest() // reset

			tc.malleate()

			err := suite.app.IncentivesKeeper.CancelIncentive(suite.ctx, contract)
			suite.Commit()

			_, ok := suite.app.IncentivesKeeper.GetIncentive(suite.ctx, contract)
			if tc.expPass {
				suite.Require().NoError(err, tc.name)
				suite.Require().False(ok, tc.name)

			} else {
				suite.Require().Error(err, tc.name)
				suite.Require().False(ok, tc.name)
			}
		})
	}
}
