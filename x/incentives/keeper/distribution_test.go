package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/tharsis/evmos/x/incentives/types"
)

func (suite *KeeperTestSuite) TestDistributeIncentives() {
	const (
		mintAmount   int64  = 100
		gasUsed      uint64 = 500
		totalGasUsed uint64 = 1000
	)

	var (
		regIn types.Incentive
	)

	testCases := []struct {
		name     string
		malleate func()
		expPass  bool
	}{
		// {
		// TODO pass - with mint denom and ramaining epochs",
		// },

		{
			"pass - with non-mint denom and no remaining epochs",
			func() {
				// create incentive
				regIn = types.NewIncentive(contract2, allocations, 1)
				suite.app.IncentivesKeeper.SetIncentive(suite.ctx, regIn)

				// Mint tokens in module account
				err := suite.app.BankKeeper.MintCoins(
					suite.ctx,
					types.ModuleName,
					sdk.Coins{sdk.NewInt64Coin(denomCoin, mintAmount)},
				)
				suite.Require().NoError(err)

				// check module balance
				moduleAddr := suite.app.AccountKeeper.GetModuleAddress(types.ModuleName)
				balance := suite.app.BankKeeper.GetBalance(suite.ctx, moduleAddr, denomCoin)
				suite.Require().True(balance.IsPositive())

				// create Gas Meter
				gm := types.NewGasMeter(contract2, participant, gasUsed)
				suite.app.IncentivesKeeper.SetGasMeter(suite.ctx, gm)

				// Set total gas meter
				suite.app.IncentivesKeeper.SetIncentiveTotalGas(
					suite.ctx,
					regIn,
					totalGasUsed,
				)
				suite.Commit()
			},
			true,
		},
		{
			"pass - with non-mint denom and remaining epochs",
			func() {
				// create incentive
				regIn = types.NewIncentive(contract2, allocations, epochs)
				suite.app.IncentivesKeeper.SetIncentive(suite.ctx, regIn)

				// Mint tokens in module account
				err := suite.app.BankKeeper.MintCoins(
					suite.ctx,
					types.ModuleName,
					sdk.Coins{sdk.NewInt64Coin(denomCoin, mintAmount)},
				)
				suite.Require().NoError(err)

				// check module balance
				moduleAddr := suite.app.AccountKeeper.GetModuleAddress(types.ModuleName)
				balance := suite.app.BankKeeper.GetBalance(suite.ctx, moduleAddr, denomCoin)
				suite.Require().True(balance.IsPositive())

				// create Gas Meter
				gm := types.NewGasMeter(contract2, participant, gasUsed)
				suite.app.IncentivesKeeper.SetGasMeter(suite.ctx, gm)

				// Set total gas meter
				suite.app.IncentivesKeeper.SetIncentiveTotalGas(
					suite.ctx,
					regIn,
					totalGasUsed,
				)
				suite.Commit()
			},
			true,
		},
	}
	for _, tc := range testCases {
		suite.SetupTest() // reset

		tc.malleate()

		err := suite.app.IncentivesKeeper.DistributeIncentives(suite.ctx)

		if tc.expPass {
			suite.Require().NoError(err)

			// distributes the rewards to all paricpants
			sdkParticipant := sdk.AccAddress(participant.Bytes())
			balance := suite.app.BankKeeper.GetBalance(suite.ctx, sdkParticipant, denomCoin)
			gasRatio := float64(gasUsed) / float64(totalGasUsed)
			coinAllocated := float64(mintAmount) * (float64(allocationRate) / 100)
			expBalance := int64(coinAllocated * gasRatio)
			suite.Require().Equal(expBalance, balance.Amount.Int64())

			// deletes all gas meters
			_, found := suite.app.IncentivesKeeper.GetIncentiveGasMeter(suite.ctx, contract2, participant)
			suite.Require().False(found)

			// updates the remaining epochs of each incentive and sets the cumulative
			// totalGas to zero OR deletes incentive
			regIn, found = suite.app.IncentivesKeeper.GetIncentive(suite.ctx, contract2)
			if regIn.IsActive() {
				suite.Require().True(found)
				suite.Require().Equal(epochs-1, regIn.Epochs)

				expTotalGas := suite.app.IncentivesKeeper.GetIncentiveTotalGas(suite.ctx, regIn)
				suite.Require().Zero(expTotalGas)
			} else {
				suite.Require().False(found)
			}

		} else {
			suite.Require().Error(err)
		}
	}
}
