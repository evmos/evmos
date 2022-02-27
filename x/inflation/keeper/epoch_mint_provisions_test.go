package keeper_test

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (suite *KeeperTestSuite) TestSetGetEpochMintProvision() {
	expEpochMintProvision := sdk.NewDec(1_000_000)

	testCases := []struct {
		name     string
		malleate func()
		genesis  bool
	}{
		{
			"default EpochMintProvision",
			func() {},
			true,
		},
		{
			"period EpochMintProvision",
			func() {
				suite.app.InflationKeeper.SetEpochMintProvision(suite.ctx, expEpochMintProvision)
			},
			false,
		},
	}
	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			suite.SetupTest() // reset

			tc.malleate()

			provision, found := suite.app.InflationKeeper.GetEpochMintProvision(suite.ctx)
			suite.Require().True(found)
			genesisProvision := sdk.NewDec(847602).Mul(sdk.DefaultPowerReduction.ToDec())
			if tc.genesis {
				suite.Require().Equal(genesisProvision, provision, tc.name)
			} else {
				suite.Require().Equal(expEpochMintProvision, provision, tc.name)
			}
		})
	}
}
