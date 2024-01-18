package keeper_test

import (
	"cosmossdk.io/math"
	inflationkeeper "github.com/evmos/evmos/v16/x/inflation/v1/keeper"
)

func (suite *KeeperTestSuite) TestInitGenesis() {
	// check calculated epochMintProvision at genesis
	epochMintProvision := suite.app.InflationKeeper.GetEpochMintProvision(suite.ctx)
	expMintProvision := math.LegacyMustNewDecFromStr("847602739726027397260274.000000000000000000").Quo(math.LegacyNewDec(inflationkeeper.ReductionFactor))
	suite.Require().Equal(expMintProvision, epochMintProvision)
}
