package keeper_test

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/tharsis/ethermint/tests"
	incentivestypes "github.com/tharsis/evmos/x/incentives/types"
	"github.com/tharsis/evmos/x/inflation/types"
)

func (suite *KeeperTestSuite) TestMintAndAllocateInflation() {
	testCases := []struct {
		name                  string
		mintCoin              sdk.Coin
		teamAddress           string
		malleate              func()
		expStakingRewardAmt   sdk.Coin
		expUsageIncentivesAmt sdk.Coin
		expCommunityPoolAmt   sdk.DecCoin
		expTeamVestingAmt     sdk.Coin
		expPass               bool
	}{
		{
			"pass",
			sdk.NewCoin(denomMint, sdk.NewInt(1_000_000)),
			sdk.AccAddress(tests.GenerateAddress().Bytes()).String(),
			func() {
				// TODO Add case where funds funds are not sufficient
				// coins := sdk.NewCoins(sdk.NewCoin(denomMint, sdk.NewInt(200_000_000)))
				// suite.app.BankKeeper.MintCoins(suite.ctx, types.ModuleName, coins)
				// suite.app.BankKeeper.SendCoinsFromModuleToModule(
				// 	suite.ctx, types.ModuleName, types.UnvestedTeamAccount, coins,
				// )

			},
			sdk.NewCoin(denomMint, sdk.NewInt(533_333)),
			sdk.NewCoin(denomMint, sdk.NewInt(333_333)),
			sdk.NewDecCoin(denomMint, sdk.NewInt(133_334)),
			sdk.NewCoin(denomMint, sdk.NewInt(136_986)),
			true,
		},
	}
	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			suite.SetupTest() // reset

			// Set teamaddress
			params := suite.app.InflationKeeper.GetParams(suite.ctx)
			params.TeamAddress = tc.teamAddress
			suite.app.InflationKeeper.SetParams(suite.ctx, params)
			teamAddress, err := sdk.AccAddressFromBech32(params.TeamAddress)
			suite.Require().NoError(err)

			tc.malleate()

			err = suite.app.InflationKeeper.MintAndAllocateInflation(suite.ctx, tc.mintCoin)

			balance := suite.app.BankKeeper.GetBalance(
				suite.ctx,
				suite.app.AccountKeeper.GetModuleAddress(types.ModuleName),
				denomMint,
			)

			feeCollector := suite.app.AccountKeeper.GetModuleAddress(authtypes.FeeCollectorName)
			balanceStakingRewards := suite.app.BankKeeper.GetBalance(
				suite.ctx,
				feeCollector,
				denomMint,
			)
			incentives := suite.app.AccountKeeper.GetModuleAddress(incentivestypes.ModuleName)
			balanceUsageIncentives := suite.app.BankKeeper.GetBalance(
				suite.ctx,
				incentives,
				denomMint,
			)

			communityPool := suite.app.DistrKeeper.GetFeePoolCommunityCoins(suite.ctx)

			balanceTeamVesting := suite.app.BankKeeper.GetBalance(
				suite.ctx,
				teamAddress,
				denomMint,
			)

			if tc.expPass {
				suite.Require().NoError(err, tc.name)
				suite.Require().True(balance.IsZero())
				suite.Require().Equal(tc.expStakingRewardAmt, balanceStakingRewards)
				suite.Require().Equal(tc.expUsageIncentivesAmt, balanceUsageIncentives)
				if tc.teamAddress == "" {
					expTotalCommunityPool := sdk.NewDecCoins(tc.expCommunityPoolAmt).Add(sdk.NewDecCoinFromCoin(tc.expTeamVestingAmt))
					suite.Require().Equal(expTotalCommunityPool, communityPool)
					suite.Require().True(tc.expTeamVestingAmt.IsZero())
				} else {
					suite.Require().Equal(sdk.NewDecCoins(tc.expCommunityPoolAmt), communityPool)
					suite.Require().Equal(tc.expTeamVestingAmt, balanceTeamVesting)
				}

			} else {
				suite.Require().Error(err)
			}
		})
	}
}
