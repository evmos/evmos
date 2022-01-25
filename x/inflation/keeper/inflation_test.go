package keeper_test

// import (
// 	"fmt"

// 	sdk "github.com/cosmos/cosmos-sdk/types"
// 	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
// 	incentivestypes "github.com/tharsis/evmos/x/incentives/types"
// 	"github.com/tharsis/evmos/x/inflation/types"
// )

// func (suite *KeeperTestSuite) TestMintAndAllocateInflation() {
// 	testCases := []struct {
// 		name               string
// 		coin               sdk.Coin
// 		malleate           func()
// 		expStakingRewards  sdk.Coin
// 		expUsageIncentives sdk.Coin
// 		expCommunityPool   sdk.Coins
// 		expPass            bool
// 	}{
// 		// TODO fix with accounts
// 		{
// 			"pass",
// 			sdk.NewCoin(denomMint, sdk.NewInt(600_000)),
// 			func() {
// 				coins := sdk.NewCoins(sdk.NewCoin(denomMint, sdk.NewInt(600_000)))
// 				suite.app.BankKeeper.MintCoins(suite.ctx, types.UnvestedTharsisAccount, coins)
// 			},
// 			sdk.NewCoin(denomMint, sdk.NewInt(1000)),
// 			sdk.NewCoin(denomMint, sdk.NewInt(1000)),
// 			sdk.NewCoins(sdk.NewCoin(denomMint, sdk.NewInt(1000))),
// 			true,
// 		},
// 	}
// 	for _, tc := range testCases {
// 		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
// 			suite.SetupTest() // reset

// 			tc.malleate()

// 			err := suite.app.InflationKeeper.MintAndAllocateInflation(suite.ctx, tc.coin)

// 			balance := suite.app.BankKeeper.GetBalance(
// 				suite.ctx,
// 				sdk.AccAddress(types.ModuleAddress.Bytes()),
// 				denomMint,
// 			)

// 			feeCollector := suite.app.AccountKeeper.GetModuleAddress(authtypes.FeeCollectorName)
// 			balanceStakingRewards := suite.app.BankKeeper.GetBalance(
// 				suite.ctx,
// 				feeCollector,
// 				denomMint,
// 			)
// 			incentives := suite.app.AccountKeeper.GetModuleAddress(incentivestypes.ModuleName)
// 			balanceUsageIncentives := suite.app.BankKeeper.GetBalance(
// 				suite.ctx,
// 				incentives,
// 				denomMint,
// 			)
// 			communityPool := suite.app.DistrKeeper.GetFeePoolCommunityCoins(suite.ctx)
// 			if tc.expPass {
// 				suite.Require().NoError(err, tc.name)
// 				suite.Require().Zero(balance)
// 				suite.Require().Equal(tc.expStakingRewards, balanceStakingRewards)
// 				suite.Require().Equal(tc.expUsageIncentives, balanceUsageIncentives)
// 				suite.Require().Equal(tc.expCommunityPool, communityPool)
// 			} else {
// 				suite.Require().Error(err)
// 			}
// 		})
// 	}
// }
