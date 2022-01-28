package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (suite *KeeperTestSuite) TestInitGenesis() {
	// check calculated epochMintProvison at genesis
	epochMintProvision, _ := suite.app.InflationKeeper.GetEpochMintProvision(suite.ctx)
	suite.Require().Equal(sdk.NewDec(847602), epochMintProvision)
}
