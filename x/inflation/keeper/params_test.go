package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/tharsis/evmos/x/inflation/types"
)

func (suite *KeeperTestSuite) TestParams() {
	params := suite.app.InflationKeeper.GetParams(suite.ctx)
	expParams := types.DefaultParams()
	// manually set team address at genesis
	expParams.TeamAddress = sdk.AccAddress(suite.address.Bytes()).String()

	suite.Require().Equal(expParams, params)

	params.EpochsPerPeriod = 700
	suite.app.InflationKeeper.SetParams(suite.ctx, params)
	newParams := suite.app.InflationKeeper.GetParams(suite.ctx)
	suite.Require().Equal(newParams, params)
}
