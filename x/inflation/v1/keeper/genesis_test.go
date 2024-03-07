package keeper_test

import (
	"cosmossdk.io/math"
	"github.com/evmos/evmos/v16/testutil/integration/evmos/network"
	inflationkeeper "github.com/evmos/evmos/v16/x/inflation/v1/keeper"
)

func (suite *KeeperTestSuite) TestInitGenesis() {
	nw := network.NewUnitTestNetwork()
	ctx := nw.GetContext()
	// check calculated epochMintProvision at genesis
	epochMintProvision := nw.App.InflationKeeper.GetEpochMintProvision(ctx)
	expMintProvision := math.LegacyMustNewDecFromStr("847602739726027397260274.000000000000000000").Quo(math.LegacyNewDec(inflationkeeper.ReductionFactor))
	suite.Require().Equal(expMintProvision, epochMintProvision)
}
