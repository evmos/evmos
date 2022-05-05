package keeper_test

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/tharsis/evmos/v4/x/incentives/types"
)

func (suite *KeeperTestSuite) TestDistributeIncentives() {
	const (
		mintAmount   int64  = 100
		gasUsed      uint64 = 500
		totalGasUsed uint64 = 1000
	)

	testCases := []struct {
		name        string
		allocations sdk.DecCoins
		epochs      uint32
		denom       string
		mintAmount  int64
		expPass     bool
	}{
		{
			"pass - with capped reward",
			mintAllocations,
			epochs,
			denomMint,
			1000000,
			true,
		},
		{
			"pass - with non-mint denom and no remaining epochs",
			allocations,
			1,
			denomCoin,
			mintAmount,
			true,
		},
		{
			"pass - with non-mint denom and remaining epochs",
			allocations,
			epochs,
			denomCoin,
			mintAmount,
			true,
		},
		{
			"pass - with mint denom and ramaining epochs",
			mintAllocations,
			epochs,
			denomMint,
			mintAmount,
			true,
		},
	}
	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			suite.SetupTest() // reset

			// Mint tokens in module account
			err := suite.app.BankKeeper.MintCoins(
				suite.ctx,
				types.ModuleName,
				sdk.Coins{sdk.NewInt64Coin(tc.denom, tc.mintAmount)},
			)
			suite.Require().NoError(err)

			// create incentive
			_, err = suite.app.IncentivesKeeper.RegisterIncentive(
				suite.ctx,
				contract,
				tc.allocations,
				tc.epochs,
			)
			suite.Require().NoError(err)

			regIn, found := suite.app.IncentivesKeeper.GetIncentive(suite.ctx, contract)
			suite.Require().True(found)

			// check module balance
			moduleAddr := suite.app.AccountKeeper.GetModuleAddress(types.ModuleName)
			balance := suite.app.BankKeeper.GetBalance(suite.ctx, moduleAddr, tc.denom)
			suite.Require().True(balance.IsPositive())

			// create Gas Meter
			gm := types.NewGasMeter(contract, participant, gasUsed)
			suite.app.IncentivesKeeper.SetGasMeter(suite.ctx, gm)

			// Set total gas meter
			suite.app.IncentivesKeeper.SetIncentiveTotalGas(
				suite.ctx,
				regIn,
				totalGasUsed,
			)
			suite.Commit()

			err = suite.app.IncentivesKeeper.DistributeIncentives(suite.ctx)

			if tc.expPass {
				suite.Require().NoError(err, tc.name)

				// distributes the rewards to all participants
				sdkParticipant := sdk.AccAddress(participant.Bytes())
				balance := suite.app.BankKeeper.GetBalance(suite.ctx, sdkParticipant, tc.denom)

				gasRatio := sdk.NewDec(int64(gasUsed)).QuoInt64(int64(totalGasUsed))
				coinAllocated := sdk.NewDec(tc.mintAmount).MulInt64(allocationRate).QuoInt64(100)
				expBalance := coinAllocated.Mul(gasRatio)
				params := suite.app.IncentivesKeeper.GetParams(suite.ctx)
				expBalance = sdk.MinDec(expBalance, params.RewardScaler.MulInt64(int64(gasUsed)))

				suite.Require().Equal(expBalance.TruncateInt(), balance.Amount, tc.name)

				// deletes all gas meters
				_, found := suite.app.IncentivesKeeper.GetGasMeter(suite.ctx, contract, participant)
				suite.Require().False(found)

				// updates the remaining epochs of each incentive and sets the cumulative
				// totalGas to zero OR deletes incentive
				regIn, found = suite.app.IncentivesKeeper.GetIncentive(suite.ctx, contract)
				if regIn.IsActive() {
					suite.Require().True(found)
					suite.Require().Equal(tc.epochs-1, regIn.Epochs)

					expTotalGas := regIn.TotalGas
					suite.Require().Zero(expTotalGas)
				} else {
					suite.Require().False(found)
				}

			} else {
				suite.Require().Error(err)
			}
		})
	}
}
