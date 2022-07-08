package keeper_test

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/evmos/ethermint/tests"
	"github.com/evmos/evmos/v6/testutil"
	"github.com/evmos/evmos/v6/x/claims/types"
)

func (suite *KeeperTestSuite) TestAfterProposalVote() {
	addr := sdk.AccAddress(tests.GenerateAddress().Bytes())

	testCases := []struct {
		name string
		test func()
	}{
		{
			"no claim record",
			func() {
				suite.app.ClaimsKeeper.AfterProposalVote(suite.ctx, 1, addr)
			},
		},
		{
			"claim disabled",
			func() {
				params := suite.app.ClaimsKeeper.GetParams(suite.ctx)
				params.EnableClaims = false
				suite.app.ClaimsKeeper.SetParams(suite.ctx, params)
				suite.app.ClaimsKeeper.SetClaimsRecord(suite.ctx, addr, types.ClaimsRecord{})

				suite.app.ClaimsKeeper.AfterProposalVote(suite.ctx, 1, addr)
			},
		},
		{
			"claim enabled - already claimed",
			func() {
				params := types.Params{
					EnableClaims:       true,
					AirdropStartTime:   suite.ctx.BlockTime().Add(-time.Hour),
					DurationUntilDecay: 2 * time.Hour,
					DurationOfDecay:    time.Hour,
					ClaimsDenom:        types.DefaultClaimsDenom,
				}

				claimRecord := types.NewClaimsRecord(sdk.NewInt(1000))
				claimRecord.MarkClaimed(types.ActionVote)
				suite.app.ClaimsKeeper.SetParams(suite.ctx, params)
				suite.app.ClaimsKeeper.SetClaimsRecord(suite.ctx, addr, claimRecord)

				expBalance := suite.app.BankKeeper.GetBalance(suite.ctx, addr, params.ClaimsDenom)

				suite.app.ClaimsKeeper.AfterProposalVote(suite.ctx, 1, addr)

				newClaimRec, found := suite.app.ClaimsKeeper.GetClaimsRecord(suite.ctx, addr)
				suite.Require().True(found)
				suite.Require().True(newClaimRec.HasClaimedAction(types.ActionVote))

				balance := suite.app.BankKeeper.GetBalance(suite.ctx, addr, params.ClaimsDenom)
				suite.Require().Equal(expBalance, balance)
			},
		},
		{
			"claim enabled - claim",
			func() {
				params := types.Params{
					EnableClaims:       true,
					AirdropStartTime:   suite.ctx.BlockTime().Add(-time.Hour),
					DurationUntilDecay: 2 * time.Hour,
					DurationOfDecay:    time.Hour,
					ClaimsDenom:        types.DefaultClaimsDenom,
				}
				claimRecord := types.NewClaimsRecord(sdk.NewInt(1000))

				suite.app.ClaimsKeeper.SetParams(suite.ctx, params)
				suite.app.ClaimsKeeper.SetClaimsRecord(suite.ctx, addr, claimRecord)

				expBalance := suite.app.BankKeeper.GetBalance(suite.ctx, addr, params.ClaimsDenom)

				coins := sdk.Coins{sdk.NewCoin(params.ClaimsDenom, sdk.NewInt(250))}
				err := testutil.FundModuleAccount(suite.app.BankKeeper, suite.ctx, types.ModuleName, coins)
				suite.Require().NoError(err)

				suite.app.ClaimsKeeper.AfterProposalVote(suite.ctx, 1, addr)

				newClaimRec, found := suite.app.ClaimsKeeper.GetClaimsRecord(suite.ctx, addr)
				suite.Require().True(found)
				suite.Require().True(newClaimRec.HasClaimedAction(types.ActionVote))

				expBalance = expBalance.Add(sdk.NewCoin(params.ClaimsDenom, sdk.NewInt(250)))
				balance := suite.app.BankKeeper.GetBalance(suite.ctx, addr, params.ClaimsDenom)
				suite.Require().Equal(expBalance, balance)
			},
		},
		{
			"no-op: error during claim",
			func() {
				params := types.Params{
					EnableClaims:       true,
					AirdropStartTime:   suite.ctx.BlockTime().Add(-time.Hour),
					DurationUntilDecay: 2 * time.Hour,
					DurationOfDecay:    time.Hour,
					ClaimsDenom:        types.DefaultClaimsDenom,
				}
				claimRecord := types.NewClaimsRecord(sdk.NewInt(1000))

				suite.app.ClaimsKeeper.SetParams(suite.ctx, params)
				suite.app.ClaimsKeeper.SetClaimsRecord(suite.ctx, addr, claimRecord)

				expBalance := suite.app.BankKeeper.GetBalance(suite.ctx, addr, params.ClaimsDenom)

				suite.app.ClaimsKeeper.AfterProposalVote(suite.ctx, 1, addr)

				newClaimRec, found := suite.app.ClaimsKeeper.GetClaimsRecord(suite.ctx, addr)
				suite.Require().True(found)
				suite.Require().False(newClaimRec.HasClaimedAction(types.ActionVote))

				balance := suite.app.BankKeeper.GetBalance(suite.ctx, addr, params.ClaimsDenom)
				suite.Require().Equal(expBalance, balance)
			},
		},
	}
	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			suite.SetupTest() // reset

			suite.app.AccountKeeper.SetAccount(suite.ctx, authtypes.NewBaseAccount(addr, nil, 0, 0))
			tc.test()
		})
	}
}

func (suite *KeeperTestSuite) TestAfterDelegation() {
	addr := sdk.AccAddress(tests.GenerateAddress().Bytes())
	addr2 := sdk.ValAddress(tests.GenerateAddress().Bytes())

	testCases := []struct {
		name string
		test func()
	}{
		{
			"no claim record",
			func() {
				suite.app.ClaimsKeeper.AfterDelegationModified(suite.ctx, addr, addr2)
			},
		},
		{
			"claim disabled",
			func() {
				params := types.Params{
					EnableClaims:       false,
					AirdropStartTime:   suite.ctx.BlockTime().Add(-time.Hour),
					DurationUntilDecay: 2 * time.Hour,
					DurationOfDecay:    time.Hour,
					ClaimsDenom:        types.DefaultClaimsDenom,
				}
				claimRecord := types.NewClaimsRecord(sdk.NewInt(1000))

				suite.app.ClaimsKeeper.SetParams(suite.ctx, params)
				suite.app.ClaimsKeeper.SetClaimsRecord(suite.ctx, addr, claimRecord)

				suite.app.ClaimsKeeper.AfterDelegationModified(suite.ctx, addr, addr2)
			},
		},
		{
			"claim enabled - already claimed",
			func() {
				params := types.Params{
					EnableClaims:       true,
					AirdropStartTime:   suite.ctx.BlockTime().Add(-time.Hour),
					DurationUntilDecay: 2 * time.Hour,
					DurationOfDecay:    time.Hour,
					ClaimsDenom:        types.DefaultClaimsDenom,
				}

				claimRecord := types.NewClaimsRecord(sdk.NewInt(1000))
				claimRecord.MarkClaimed(types.ActionDelegate)

				suite.app.ClaimsKeeper.SetParams(suite.ctx, params)
				suite.app.ClaimsKeeper.SetClaimsRecord(suite.ctx, addr, claimRecord)

				suite.app.ClaimsKeeper.AfterDelegationModified(suite.ctx, addr, addr2)
			},
		},
		{
			"claim enabled - claim",
			func() {
				params := types.Params{
					EnableClaims:       true,
					AirdropStartTime:   suite.ctx.BlockTime().Add(-time.Hour),
					DurationUntilDecay: 2 * time.Hour,
					DurationOfDecay:    time.Hour,
					ClaimsDenom:        types.DefaultClaimsDenom,
				}

				suite.app.ClaimsKeeper.SetParams(suite.ctx, params)

				claimRecord := types.NewClaimsRecord(sdk.NewInt(1000))
				suite.app.ClaimsKeeper.SetClaimsRecord(suite.ctx, addr, claimRecord)

				expBalance := suite.app.BankKeeper.GetBalance(suite.ctx, addr, params.ClaimsDenom)

				coins := sdk.Coins{sdk.NewCoin(params.ClaimsDenom, sdk.NewInt(250))}
				err := testutil.FundModuleAccount(suite.app.BankKeeper, suite.ctx, types.ModuleName, coins)
				suite.Require().NoError(err)

				suite.app.ClaimsKeeper.AfterDelegationModified(suite.ctx, addr, addr2)

				newClaimRec, found := suite.app.ClaimsKeeper.GetClaimsRecord(suite.ctx, addr)
				suite.Require().True(found)
				suite.Require().True(newClaimRec.HasClaimedAction(types.ActionDelegate))

				expBalance = expBalance.Add(sdk.NewCoin(params.ClaimsDenom, sdk.NewInt(250)))
				balance := suite.app.BankKeeper.GetBalance(suite.ctx, addr, types.DefaultClaimsDenom)

				suite.Require().Equal(expBalance, balance)
			},
		},
		{
			"no-op: error during claim",
			func() {
				params := types.Params{
					EnableClaims:       true,
					AirdropStartTime:   suite.ctx.BlockTime().Add(-time.Hour),
					DurationUntilDecay: 2 * time.Hour,
					DurationOfDecay:    time.Hour,
					ClaimsDenom:        types.DefaultClaimsDenom,
				}
				claimRecord := types.NewClaimsRecord(sdk.NewInt(1000))

				suite.app.ClaimsKeeper.SetParams(suite.ctx, params)
				suite.app.ClaimsKeeper.SetClaimsRecord(suite.ctx, addr, claimRecord)

				expBalance := suite.app.BankKeeper.GetBalance(suite.ctx, addr, params.ClaimsDenom)

				suite.app.ClaimsKeeper.AfterDelegationModified(suite.ctx, addr, addr2)

				newClaimRec, found := suite.app.ClaimsKeeper.GetClaimsRecord(suite.ctx, addr)
				suite.Require().True(found)
				suite.Require().False(newClaimRec.HasClaimedAction(types.ActionDelegate))

				balance := suite.app.BankKeeper.GetBalance(suite.ctx, addr, params.ClaimsDenom)
				suite.Require().Equal(expBalance, balance)
			},
		},
	}

	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			suite.SetupTest() // reset

			suite.app.AccountKeeper.SetAccount(suite.ctx, authtypes.NewBaseAccount(addr, nil, 0, 0))
			tc.test()
		})
	}
}

func (suite *KeeperTestSuite) TestAfterEVMStateTransition() {
	from := tests.GenerateAddress()
	to := tests.GenerateAddress()
	msg := ethtypes.NewMessage(from, &to, 0, nil, 0, nil, nil, nil, nil, nil, false)

	receipt := ethtypes.Receipt{}
	addr := sdk.AccAddress(from.Bytes())

	testCases := []struct {
		name string
		test func()
	}{
		{
			"no claim record",
			func() {
				err := suite.app.ClaimsKeeper.PostTxProcessing(suite.ctx, msg, &receipt)
				suite.Require().NoError(err)
			},
		},
		{
			"claim disabled",
			func() {
				params := types.Params{
					EnableClaims:       false,
					AirdropStartTime:   suite.ctx.BlockTime().Add(-time.Hour),
					DurationUntilDecay: 2 * time.Hour,
					DurationOfDecay:    time.Hour,
					ClaimsDenom:        types.DefaultClaimsDenom,
				}
				claimRecord := types.NewClaimsRecord(sdk.NewInt(1000))

				suite.app.ClaimsKeeper.SetParams(suite.ctx, params)
				suite.app.ClaimsKeeper.SetClaimsRecord(suite.ctx, addr, claimRecord)

				err := suite.app.ClaimsKeeper.PostTxProcessing(suite.ctx, msg, &receipt)
				suite.Require().NoError(err)
			},
		},
		{
			"claim enabled - already claimed",
			func() {
				params := types.Params{
					EnableClaims:       true,
					AirdropStartTime:   suite.ctx.BlockTime().Add(-time.Hour),
					DurationUntilDecay: 2 * time.Hour,
					DurationOfDecay:    time.Hour,
					ClaimsDenom:        types.DefaultClaimsDenom,
				}

				claimRecord := types.NewClaimsRecord(sdk.NewInt(1000))
				claimRecord.MarkClaimed(types.ActionEVM)

				suite.app.ClaimsKeeper.SetParams(suite.ctx, params)
				suite.app.ClaimsKeeper.SetClaimsRecord(suite.ctx, addr, claimRecord)

				err := suite.app.ClaimsKeeper.PostTxProcessing(suite.ctx, msg, &receipt)
				suite.Require().NoError(err)
			},
		},
		{
			"claim enabled - claim",
			func() {
				params := types.Params{
					EnableClaims:       true,
					AirdropStartTime:   suite.ctx.BlockTime().Add(-time.Hour),
					DurationUntilDecay: 2 * time.Hour,
					DurationOfDecay:    time.Hour,
					ClaimsDenom:        types.DefaultClaimsDenom,
				}

				suite.app.ClaimsKeeper.SetParams(suite.ctx, params)

				claimRecord := types.NewClaimsRecord(sdk.NewInt(1000))
				suite.app.ClaimsKeeper.SetClaimsRecord(suite.ctx, addr, claimRecord)

				expBalance := suite.app.BankKeeper.GetBalance(suite.ctx, addr, params.ClaimsDenom)

				coins := sdk.Coins{sdk.NewCoin(params.ClaimsDenom, sdk.NewInt(250))}
				err := testutil.FundModuleAccount(suite.app.BankKeeper, suite.ctx, types.ModuleName, coins)
				suite.Require().NoError(err)

				err = suite.app.ClaimsKeeper.PostTxProcessing(suite.ctx, msg, &receipt)
				suite.Require().NoError(err)

				newClaimRec, found := suite.app.ClaimsKeeper.GetClaimsRecord(suite.ctx, addr)
				suite.Require().True(found)
				suite.Require().True(newClaimRec.HasClaimedAction(types.ActionEVM))

				expBalance = expBalance.Add(sdk.NewCoin(params.ClaimsDenom, sdk.NewInt(250)))
				balance := suite.app.BankKeeper.GetBalance(suite.ctx, addr, types.DefaultClaimsDenom)

				suite.Require().Equal(expBalance, balance)
			},
		},
		{
			"no-op: error during claim",
			func() {
				params := types.Params{
					EnableClaims:       true,
					AirdropStartTime:   suite.ctx.BlockTime().Add(-time.Hour),
					DurationUntilDecay: 2 * time.Hour,
					DurationOfDecay:    time.Hour,
					ClaimsDenom:        types.DefaultClaimsDenom,
				}
				claimRecord := types.NewClaimsRecord(sdk.NewInt(1000))

				suite.app.ClaimsKeeper.SetParams(suite.ctx, params)
				suite.app.ClaimsKeeper.SetClaimsRecord(suite.ctx, addr, claimRecord)

				expBalance := suite.app.BankKeeper.GetBalance(suite.ctx, addr, params.ClaimsDenom)

				err := suite.app.ClaimsKeeper.PostTxProcessing(suite.ctx, msg, &receipt)
				suite.Require().NoError(err)

				newClaimRec, found := suite.app.ClaimsKeeper.GetClaimsRecord(suite.ctx, addr)
				suite.Require().True(found)
				suite.Require().False(newClaimRec.HasClaimedAction(types.ActionEVM))

				balance := suite.app.BankKeeper.GetBalance(suite.ctx, addr, params.ClaimsDenom)
				suite.Require().Equal(expBalance, balance)
			},
		},
	}

	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			suite.SetupTest() // reset

			suite.app.AccountKeeper.SetAccount(suite.ctx, authtypes.NewBaseAccount(addr, nil, 0, 0))
			tc.test()
		})
	}
}
