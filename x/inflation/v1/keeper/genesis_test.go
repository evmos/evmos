package keeper_test

import (
	"cosmossdk.io/math"
)

func (suite *KeeperTestSuite) TestInitGenesis() {
	// check calculated epochMintProvision at genesis
	epochMintProvision := suite.app.InflationKeeper.GetEpochMintProvision(suite.ctx)
	expMintProvision := math.LegacyMustNewDecFromStr("282534246575342465753425.000000000000000000")
	suite.Require().Equal(expMintProvision, epochMintProvision)
}
