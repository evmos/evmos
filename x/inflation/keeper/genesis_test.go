package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/tharsis/evmos/x/inflation/types"
)

func (suite *KeeperTestSuite) TestInitGenesis() {
	// check calculated epochMintProvison at genesis
	epochMintProvision, _ := suite.app.InflationKeeper.GetEpochMintProvision(suite.ctx)
	suite.Require().Equal(sdk.NewDec(847602), epochMintProvision)

	// check intital account balance on unvested team account
	unvestedTeamAccount := suite.app.AccountKeeper.GetModuleAddress(types.UnvestedTeamAccount)
	balances := suite.app.BankKeeper.GetAllBalances(suite.ctx, unvestedTeamAccount)
	expBalances := sdk.NewCoins(sdk.NewCoin(denomMint, sdk.NewInt(200_000_000)))
	suite.Require().Equal(expBalances, balances)
}
