package keeper_test

import (
	"fmt"

	"github.com/tharsis/evmos/x/incentives/types"
)

func (suite *KeeperTestSuite) TestDistributeIncentives() {
	const (
		mintAmount   = 100
		gasUsed      = 500
		totalGasUsed = 1000
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
		// {
		// TODO pass - with non-mint denom and no remaining epochs",
		// },
		// {
		// 	"pass - with non-mint denom and ramaining epochs",
		// 	func() {
		// 		// create incnetive
		// 		regIn = types.NewIncentive(contract2, allocations, epochs)
		// 		suite.app.IncentivesKeeper.SetIncentive(suite.ctx, regIn)
		// 		// Mint tokens in module account
		// 		err := suite.app.BankKeeper.MintCoins(
		// 			suite.ctx,
		// 			minttypes.ModuleName,
		// 			sdk.Coins{sdk.NewInt64Coin(denomCoin, mintAmount)},
		// 		)
		// 		suite.Require().NoError(err)

		// 		// create Gas Meter
		// 		gm := types.NewGasMeter(contract2, participant, gasUsed)
		// 		suite.app.IncentivesKeeper.SetGasMeter(suite.ctx, gm)

		// 		// Set total gas meter
		// 		suite.app.IncentivesKeeper.SetIncentiveTotalGas(
		// 			suite.ctx,
		// 			regIn,
		// 			totalGasUsed,
		// 		)
		// 		suite.Commit()
		// 	},
		// 	true,
		// },
	}
	for _, tc := range testCases {
		suite.SetupTest() // reset

		tc.malleate()
		fmt.Println(regIn)

		err := suite.app.IncentivesKeeper.DistributeIncentives(suite.ctx)

		if tc.expPass {
			suite.Require().NoError(err)

			// distributes the rewards to all paricpants
			balance := suite.app.BankKeeper.GetBalance(suite.ctx, participant[:], denomCoin)
			expBalance := int64(mintAmount * (allocationRate / 100) * (gasUsed / totalGasUsed))
			suite.Require().Equal(expBalance, balance.Amount.Int64())

			// deletes all gas meters
			_, found := suite.app.IncentivesKeeper.GetIncentiveGasMeter(suite.ctx, contract2, participant)
			suite.Require().False(found)

			// updates the remaining epochs of each incentive and sets the cumulative
			// totalGas to zero OR deletes incentive
			if regIn.IsActive() {
				suite.Require().Equal(epochs-1, regIn.Epochs)

				expTotalGas := suite.app.IncentivesKeeper.GetIncentiveTotalGas(suite.ctx, regIn)
				suite.Require().Zero(expTotalGas)
			} else {
				_, found := suite.app.IncentivesKeeper.GetIncentive(suite.ctx, contract2)
				suite.Require().False(found)
			}

		} else {
			suite.Require().Error(err)

		}
	}
}
