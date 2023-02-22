package keeper_test

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	utiltx "github.com/evmos/evmos/v11/testutil/tx"
	"github.com/evmos/evmos/v11/x/incentives/types"
)

func (suite KeeperTestSuite) TestRegisterIncentive() { //nolint:govet // we can copy locks here because it is a test
	testCases := []struct {
		name                string
		malleate            func()
		expAllocationMeters []sdk.DecCoin
		expPass             bool
	}{
		{
			"incentives are disabled globally",
			func() {
				params := types.DefaultParams()
				params.EnableIncentives = false
				suite.app.IncentivesKeeper.SetParams(suite.ctx, params) //nolint:errcheck
			},
			[]sdk.DecCoin{},
			false,
		},
		{
			"contract doesn't exist",
			func() {
				contract = utiltx.GenerateAddress()
			},
			[]sdk.DecCoin{},
			false,
		},
		{
			"inventive already registered",
			func() {
				regIn := types.NewIncentive(contract, allocations, epochs)
				suite.app.IncentivesKeeper.SetIncentive(suite.ctx, regIn)
				suite.Commit()
			},
			[]sdk.DecCoin{},
			false,
		},
		{
			"coin doesn't have supply",
			func() {
			},
			[]sdk.DecCoin{},
			false,
		},
		{
			"allocation above allocation limit",
			func() {
				// Make sure the non-mint coin has supply
				err := suite.app.BankKeeper.MintCoins(
					suite.ctx,
					types.ModuleName,
					sdk.Coins{sdk.NewInt64Coin(denomCoin, 1)},
				)
				suite.Require().NoError(err)

				// decrease allocation limit
				params := types.DefaultParams()
				params.AllocationLimit = sdk.NewDecWithPrec(1, 2)
				suite.app.IncentivesKeeper.SetParams(suite.ctx, params) //nolint:errcheck
			},
			[]sdk.DecCoin{},
			false,
		},
		{
			"Total allocation for at least one denom (current + proposed) > 100%",
			func() {
				// Make sure the non-mint coin has supply
				err := suite.app.BankKeeper.MintCoins(
					suite.ctx,
					types.ModuleName,
					sdk.Coins{sdk.NewInt64Coin(denomCoin, 1)},
				)
				suite.Require().NoError(err)

				// increase allocation limit
				params := types.DefaultParams()
				params.AllocationLimit = sdk.NewDecWithPrec(100, 2)
				err = suite.app.IncentivesKeeper.SetParams(suite.ctx, params)
				suite.Require().NoError(err)

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
			[]sdk.DecCoin{sdk.NewDecCoinFromDec(denomCoin, sdk.NewDecWithPrec(100, 2))},
			false,
		},
		{
			"ok",
			func() {
				// Make sure the non-mint coin has supply
				err := suite.app.BankKeeper.MintCoins(
					suite.ctx,
					types.ModuleName,
					sdk.Coins{sdk.NewInt64Coin(denomCoin, 1)},
				)
				suite.Require().NoError(err)
			},
			[]sdk.DecCoin{allocations[1], allocations[0]},
			true,
		},
	}
	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			suite.SetupTest() // reset
			suite.deployContracts()

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
				StartTime:   suite.ctx.BlockTime(),
			}

			allocationMeters := suite.app.IncentivesKeeper.GetAllAllocationMeters(suite.ctx)
			suite.Require().Equal(tc.expAllocationMeters, allocationMeters)

			if tc.expPass {
				suite.Require().NoError(err, tc.name)
				suite.Require().Equal(expIn, in)
			} else {
				suite.Require().Error(err, tc.name)
			}
		})
	}
}

func (suite KeeperTestSuite) TestCancelIncentive() { //nolint:govet // we can copy locks here because it is a test
	testCases := []struct {
		name                string
		malleate            func()
		expAllocationMeters []sdk.DecCoin
		expPass             bool
	}{
		{
			"incentives are disabled globally",
			func() {
				params := types.DefaultParams()
				params.EnableIncentives = false
				suite.app.IncentivesKeeper.SetParams(suite.ctx, params) //nolint:errcheck
			},
			[]sdk.DecCoin{},
			false,
		},
		{
			"inventive not registered",
			func() {
			},
			[]sdk.DecCoin{},
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

				gm := types.NewGasMeter(contract, participant, uint64(100))
				suite.app.IncentivesKeeper.SetGasMeter(suite.ctx, gm)
			},
			[]sdk.DecCoin{},
			true,
		},
	}
	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			suite.SetupTest() // reset
			suite.deployContracts()

			tc.malleate()

			err := suite.app.IncentivesKeeper.CancelIncentive(suite.ctx, contract)
			suite.Commit()

			_, ok := suite.app.IncentivesKeeper.GetIncentive(suite.ctx, contract)

			allocationMeters := suite.app.IncentivesKeeper.GetAllAllocationMeters(suite.ctx)
			suite.Require().Equal(tc.expAllocationMeters, allocationMeters)

			_, found := suite.app.IncentivesKeeper.GetGasMeter(suite.ctx, contract, participant)
			if tc.expPass {
				suite.Require().NoError(err, tc.name)
				suite.Require().False(ok, tc.name)
				suite.Require().False(found)
			} else {
				suite.Require().Error(err, tc.name)
				suite.Require().False(ok, tc.name)
			}
		})
	}
}
