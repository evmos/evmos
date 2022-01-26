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
			"pass - with team address",
			sdk.NewCoin(denomMint, sdk.NewInt(1_000_000)),
			sdk.AccAddress(tests.GenerateAddress().Bytes()).String(),
			func() {},
			sdk.NewCoin(denomMint, sdk.NewInt(533_333)),
			sdk.NewCoin(denomMint, sdk.NewInt(333_333)),
			sdk.NewDecCoin(denomMint, sdk.NewInt(133_334)),
			sdk.NewCoin(denomMint, sdk.NewInt(136_986)),
			true,
		},
		{
			"pass - without team address",
			sdk.NewCoin(denomMint, sdk.NewInt(1_000_000)),
			"",
			func() {},
			sdk.NewCoin(denomMint, sdk.NewInt(533_333)),
			sdk.NewCoin(denomMint, sdk.NewInt(333_333)),
			sdk.NewDecCoin(denomMint, sdk.NewInt(133_334)),
			sdk.NewCoin(denomMint, sdk.NewInt(136_986)),
			true,
		},
		{
			"pass - without team address and no coins minted ",
			sdk.NewCoin(denomMint, sdk.ZeroInt()),
			"",
			func() {},
			sdk.NewCoin(denomMint, sdk.ZeroInt()),
			sdk.NewCoin(denomMint, sdk.ZeroInt()),
			sdk.NewDecCoin(denomMint, sdk.ZeroInt()),
			sdk.NewCoin(denomMint, sdk.NewInt(136_986)),
			true,
		},
		{
			"pass - insufficient escrow balance - no supply",
			sdk.NewCoin(denomMint, sdk.NewInt(1_000_000)),
			sdk.AccAddress(tests.GenerateAddress().Bytes()).String(),
			func() {
				unvestedTeamAccount := suite.app.AccountKeeper.GetModuleAddress(types.UnvestedTeamAccount)
				balances := suite.app.BankKeeper.GetAllBalances(suite.ctx, unvestedTeamAccount)
				burnAcc := sdk.AccAddress(tests.GenerateAddress().Bytes())
				suite.app.BankKeeper.SendCoinsFromModuleToAccount(
					suite.ctx, types.UnvestedTeamAccount, burnAcc, balances,
				)
			},
			sdk.NewCoin(denomMint, sdk.NewInt(533_333)),
			sdk.NewCoin(denomMint, sdk.NewInt(333_333)),
			sdk.NewDecCoin(denomMint, sdk.NewInt(133_334)),
			sdk.NewCoin(denomMint, sdk.ZeroInt()),
			true,
		},
		{
			"pass - insufficient escrow balance - supply lower than team provision",
			sdk.NewCoin(denomMint, sdk.NewInt(1_000_000)),
			sdk.AccAddress(tests.GenerateAddress().Bytes()).String(),
			func() {
				unvestedTeamAccount := suite.app.AccountKeeper.GetModuleAddress(types.UnvestedTeamAccount)
				balances := suite.app.BankKeeper.GetAllBalances(suite.ctx, unvestedTeamAccount)

				coin := balances[0].Sub(sdk.NewCoin(denomMint, sdk.NewInt(10)))
				coins := sdk.NewCoins(coin)

				burnAcc := sdk.AccAddress(tests.GenerateAddress().Bytes())
				suite.app.BankKeeper.SendCoinsFromModuleToAccount(
					suite.ctx, types.UnvestedTeamAccount, burnAcc, coins,
				)

			},
			sdk.NewCoin(denomMint, sdk.NewInt(533_333)),
			sdk.NewCoin(denomMint, sdk.NewInt(333_333)),
			sdk.NewDecCoin(denomMint, sdk.NewInt(133_334)),
			sdk.NewCoin(denomMint, sdk.NewInt(10)),
			true,
		},
	}
	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			suite.SetupTest() // reset

			// Set team address
			params := suite.app.InflationKeeper.GetParams(suite.ctx)
			params.TeamAddress = tc.teamAddress
			suite.app.InflationKeeper.SetParams(suite.ctx, params)
			teamAddress, _ := sdk.AccAddressFromBech32(params.TeamAddress)

			tc.malleate()

			err := suite.app.InflationKeeper.MintAndAllocateInflation(suite.ctx, tc.mintCoin)

			// Get balances
			balanceModule := suite.app.BankKeeper.GetBalance(
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
				suite.Require().True(balanceModule.IsZero())

				suite.Require().Equal(tc.expStakingRewardAmt, balanceStakingRewards)
				suite.Require().Equal(tc.expUsageIncentivesAmt, balanceUsageIncentives)
				if tc.teamAddress == "" {
					expTotalCommunityPool := sdk.NewDecCoins(tc.expCommunityPoolAmt).Add(sdk.NewDecCoinFromCoin(tc.expTeamVestingAmt))
					suite.Require().Equal(expTotalCommunityPool, communityPool)
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
