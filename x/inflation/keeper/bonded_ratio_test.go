package keeper_test

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (suite *KeeperTestSuite) TestSetGetBondedRatio() {
	expBondedRatio := sdk.NewDecWithPrec(30, 2)

	testCases := []struct {
		name     string
		malleate func()
		genesis  bool
	}{
		{
			"default BondedRatio",
			func() {},
			true,
		},
		{
			"period BondedRatio",
			func() {
				suite.app.InflationKeeper.SetBondedRatio(suite.ctx, expBondedRatio)
			},
			false,
		},
	}
	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			suite.SetupTest() // reset

			tc.malleate()

			bondedRatio, found := suite.app.InflationKeeper.GetBondedRatio(suite.ctx)
			suite.Require().True(found)
			genesisbondedRatio := sdk.OneDec()
			if tc.genesis {
				suite.Require().Equal(genesisbondedRatio, bondedRatio, tc.name)
			} else {
				suite.Require().Equal(expBondedRatio, bondedRatio, tc.name)
			}
		})
	}
}
