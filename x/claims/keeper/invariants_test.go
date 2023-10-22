package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/evmos/evmos/v15/testutil"
	utiltx "github.com/evmos/evmos/v15/testutil/tx"
	"github.com/evmos/evmos/v15/x/claims/types"
)

func (suite *KeeperTestSuite) TestClaimsInvariant() {
	testCases := []struct {
		name      string
		malleate  func()
		expBroken bool
	}{
		{
			"claims inactive",
			func() {
				suite.app.ClaimsKeeper.SetParams(suite.ctx, types.DefaultParams()) //nolint:errcheck
			},
			false,
		},
		{
			"invariant broken - single claim record (nothing completed)",
			func() {
				addr := sdk.AccAddress(utiltx.GenerateAddress().Bytes())
				suite.app.ClaimsKeeper.SetClaimsRecord(suite.ctx, addr, types.NewClaimsRecord(sdk.NewInt(40)))
				suite.Require().True(suite.app.ClaimsKeeper.HasClaimsRecord(suite.ctx, addr))

				coins := sdk.Coins{sdk.NewCoin("aevmos", sdk.NewInt(100))}
				// update the escrowed account balance to maintain the invariant
				err := testutil.FundModuleAccount(suite.ctx, suite.app.BankKeeper, types.ModuleName, coins)
				suite.Require().NoError(err)
				suite.app.Commit()
			},
			true,
		},
		{
			"invariant broken - single claim record (nothing completed), low value",
			func() {
				addr := sdk.AccAddress(utiltx.GenerateAddress().Bytes())
				suite.app.ClaimsKeeper.SetClaimsRecord(suite.ctx, addr, types.NewClaimsRecord(sdk.OneInt()))
				suite.Require().True(suite.app.ClaimsKeeper.HasClaimsRecord(suite.ctx, addr))

				coins := sdk.Coins{sdk.NewCoin("aevmos", sdk.NewInt(2))}
				// update the escrowed account balance to maintain the invariant
				err := testutil.FundModuleAccount(suite.ctx, suite.app.BankKeeper, types.ModuleName, coins)
				suite.Require().NoError(err)
				suite.app.Commit()
			},
			true,
		},
		{
			"invariant broken - single claim record (all completed)",
			func() {
				addr := sdk.AccAddress(utiltx.GenerateAddress().Bytes())
				cr := types.ClaimsRecord{
					InitialClaimableAmount: sdk.NewInt(100),
					ActionsCompleted:       []bool{true, true, true, true},
				}
				suite.app.ClaimsKeeper.SetClaimsRecord(suite.ctx, addr, cr)
				suite.Require().True(suite.app.ClaimsKeeper.HasClaimsRecord(suite.ctx, addr))

				coins := sdk.Coins{sdk.NewCoin("aevmos", sdk.NewInt(100))}
				// update the escrowed account balance to maintain the invariant
				err := testutil.FundModuleAccount(suite.ctx, suite.app.BankKeeper, types.ModuleName, coins)
				suite.Require().NoError(err)
				suite.app.Commit()
			},
			true,
		},
		{
			"invariant NOT broken - single claim record",
			func() {
				addr := sdk.AccAddress(utiltx.GenerateAddress().Bytes())
				cr := types.ClaimsRecord{
					InitialClaimableAmount: sdk.NewInt(100),
					ActionsCompleted:       []bool{false, false, false, false},
				}
				suite.app.ClaimsKeeper.SetClaimsRecord(suite.ctx, addr, cr)
				suite.Require().True(suite.app.ClaimsKeeper.HasClaimsRecord(suite.ctx, addr))

				coins := sdk.Coins{sdk.NewCoin("aevmos", sdk.NewInt(100))}
				// update the escrowed account balance to maintain the invariant
				err := testutil.FundModuleAccount(suite.ctx, suite.app.BankKeeper, types.ModuleName, coins)
				suite.Require().NoError(err)
				suite.app.Commit()
			},
			false,
		},
		{
			"invariant NOT broken - multiple claim records",
			func() {
				addr := sdk.AccAddress(utiltx.GenerateAddress().Bytes())
				addr2 := sdk.AccAddress(utiltx.GenerateAddress().Bytes())
				cr := types.ClaimsRecord{
					InitialClaimableAmount: sdk.NewInt(100),
					ActionsCompleted:       []bool{false, false, false, false},
				}
				cr2 := types.ClaimsRecord{
					InitialClaimableAmount: sdk.NewInt(200),
					ActionsCompleted:       []bool{true, false, true, false},
				}
				suite.app.ClaimsKeeper.SetClaimsRecord(suite.ctx, addr, cr)
				suite.app.ClaimsKeeper.SetClaimsRecord(suite.ctx, addr2, cr2)

				suite.Require().True(suite.app.ClaimsKeeper.HasClaimsRecord(suite.ctx, addr))
				suite.Require().True(suite.app.ClaimsKeeper.HasClaimsRecord(suite.ctx, addr2))

				coins := sdk.Coins{sdk.NewCoin("aevmos", sdk.NewInt(200))}
				// update the escrowed account balance to maintain the invariant
				err := testutil.FundModuleAccount(suite.ctx, suite.app.BankKeeper, types.ModuleName, coins)
				suite.Require().NoError(err)
				suite.app.Commit()
			},
			false,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			suite.SetupTest() // reset

			tc.malleate()

			msg, broken := suite.app.ClaimsKeeper.ClaimsInvariant()(suite.ctx)
			suite.Require().Equal(tc.expBroken, broken, msg)
		})
	}
}
