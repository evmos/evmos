package keeper_test

import (
	"fmt"

	"github.com/tharsis/evmos/x/incentives/types"
)

func (suite *KeeperTestSuite) TestIncentivesGasMeters() {
	var expRes []types.GasMeter

	testCases := []struct {
		name     string
		malleate func()
	}{
		{
			"no gas meter registered",
			func() { expRes = []types.GasMeter{} },
		},
		// TODO: Fix test
		// {
		// 	"1 gas meter registered",
		// 	func() {
		// 		gm := types.NewGasMeter(contract, participant, 1)
		// 		suite.app.IncentivesKeeper.SetGasMeter(suite.ctx, gm)
		// 		suite.Commit()

		// 		expRes = []types.GasMeter{gm}
		// 	},
		// },
		// {
		// 	"2 gas meters registered",
		// 	func() {
		// 		gm := types.NewGasMeter(contract, participant, 1)
		// 		gm2 := types.NewGasMeter(contract2, participant, 1)
		// 		suite.app.IncentivesKeeper.SetGasMeter(suite.ctx, gm)
		// 		suite.app.IncentivesKeeper.SetGasMeter(suite.ctx, gm2)
		// 		suite.Commit()

		// 		expRes = []types.GasMeter{gm, gm2}
		// 	},
		// },
	}
	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			suite.SetupTest() // reset

			tc.malleate()
			res := suite.app.IncentivesKeeper.GetIncentivesGasMeters(suite.ctx)
			suite.Require().ElementsMatch(expRes, res, tc.name)
		})
	}
}

// func (suite *KeeperTestSuite) TestGetIncnetive() {
// 	expIn := types.NewIncentive(contract, allocations, epochs)
// 	suite.app.IncentivesKeeper.SetIncentive(suite.ctx, expIn)
// 	suite.Commit()

// 	testCases := []struct {
// 		name     string
// 		contract common.Address
// 		ok       bool
// 	}{
// 		{"nil address", common.Address{}, false},
// 		{"valid id", common.HexToAddress(expIn.Contract), true},
// 	}
// 	for _, tc := range testCases {
// 		in, found := suite.app.IncentivesKeeper.GetIncentive(suite.ctx, tc.contract)
// 		if tc.ok {
// 			suite.Require().True(found, tc.name)
// 			suite.Require().Equal(expIn, in, tc.name)
// 		} else {
// 			suite.Require().False(found, tc.name)
// 		}
// 	}
// }

// func (suite *KeeperTestSuite) TestDeleteIncentive() {
// 	regIn := types.NewIncentive(contract, allocations, epochs)
// 	suite.app.IncentivesKeeper.SetIncentive(suite.ctx, regIn)
// 	suite.Commit()

// 	testCases := []struct {
// 		name     string
// 		in       types.Incentive
// 		malleate func()
// 		ok       bool
// 	}{
// 		{"nil incentive", types.Incentive{}, func() {}, false},
// 		{"valid incentive", regIn, func() {}, true},
// 		{
// 			"deteted incentive",
// 			regIn,
// 			func() {
// 				suite.app.IncentivesKeeper.DeleteIncentive(suite.ctx, regIn)
// 			},
// 			false,
// 		},
// 	}
// 	for _, tc := range testCases {
// 		tc.malleate()
// 		in, found := suite.app.IncentivesKeeper.GetIncentive(
// 			suite.ctx,
// 			common.HexToAddress(tc.in.Contract),
// 		)
// 		if tc.ok {
// 			suite.Require().True(found, tc.name)
// 			suite.Require().Equal(regIn, in, tc.name)
// 		} else {
// 			suite.Require().False(found, tc.name)
// 		}
// 	}
// }

// func (suite *KeeperTestSuite) TestIsIncentiveRegistered() {
// 	regIn := types.NewIncentive(contract, allocations, epochs)
// 	suite.app.IncentivesKeeper.SetIncentive(suite.ctx, regIn)
// 	suite.Commit()

// 	testCases := []struct {
// 		name     string
// 		contract common.Address
// 		ok       bool
// 	}{
// 		{"valid id", common.HexToAddress(regIn.Contract), true},
// 		{"pair not found", common.Address{}, false},
// 	}
// 	for _, tc := range testCases {
// 		found := suite.app.IncentivesKeeper.IsIncentiveRegistered(
// 			suite.ctx,
// 			tc.contract,
// 		)
// 		if tc.ok {
// 			suite.Require().True(found, tc.name)
// 		} else {
// 			suite.Require().False(found, tc.name)
// 		}
// 	}
// }
