package keeper_test

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/tharsis/ethermint/tests"
	"github.com/tharsis/evmos/v2/x/claims/types"
)

func (suite *KeeperTestSuite) TestAfterProposalVote() {
	addr := sdk.AccAddress(tests.GenerateAddress().Bytes())
	claimRecord := types.ClaimsRecord{
		InitialClaimableAmount: sdk.NewInt(1000),
		ActionsCompleted:       []bool{false, false, false, false},
	}
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
				params := suite.app.ClaimsKeeper.GetParams(suite.ctx)
				params.EnableClaims = true
				suite.app.ClaimsKeeper.SetParams(suite.ctx, params)

				claimedRecord := types.ClaimsRecord{
					InitialClaimableAmount: sdk.NewInt(1000),
					ActionsCompleted:       []bool{true, false, false, false},
				}
				suite.app.ClaimsKeeper.SetClaimsRecord(suite.ctx, addr, claimedRecord)

				suite.app.ClaimsKeeper.AfterProposalVote(suite.ctx, 1, addr)
			},
		},
		{
			"claim enabled - claim",
			func() {
				params := suite.app.ClaimsKeeper.GetParams(suite.ctx)
				params.EnableClaims = true
				params.AirdropStartTime = time.Now().UTC()
				suite.app.ClaimsKeeper.SetParams(suite.ctx, params)

				suite.app.ClaimsKeeper.SetClaimsRecord(suite.ctx, addr, claimRecord)

				suite.ctx = suite.ctx.WithBlockTime(time.Now().UTC().Add(time.Hour))
				suite.app.ClaimsKeeper.AfterProposalVote(suite.ctx, 1, addr)

				newClaimRec, found := suite.app.ClaimsKeeper.GetClaimsRecord(suite.ctx, addr)

				expectedClaimRecord := types.ClaimsRecord{
					InitialClaimableAmount: sdk.NewInt(1000),
					ActionsCompleted:       []bool{true, false, false, false},
				}
				suite.Require().True(found)
				suite.Require().Equal(expectedClaimRecord, newClaimRec)

				balances := suite.app.BankKeeper.GetAllBalances(suite.ctx, addr)
				claimedCoins := sdk.Coins{{Denom: params.ClaimsDenom, Amount: sdk.NewInt(250)}}
				suite.Require().Equal(claimedCoins, balances)
			},
		},
	}
	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			suite.SetupClaimTest() // reset

			suite.app.AccountKeeper.SetAccount(suite.ctx, authtypes.NewBaseAccount(addr, nil, 0, 0))
			tc.test()
		})
	}
}

func (suite *KeeperTestSuite) TestAfterDelegation() {
	addr := sdk.AccAddress(tests.GenerateAddress().Bytes())
	addr2 := sdk.ValAddress(tests.GenerateAddress().Bytes())
	claimRecord := types.ClaimsRecord{
		InitialClaimableAmount: sdk.NewInt(1000),
		ActionsCompleted:       []bool{false, false, false, false},
	}
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
				params := suite.app.ClaimsKeeper.GetParams(suite.ctx)
				params.EnableClaims = false
				suite.app.ClaimsKeeper.SetParams(suite.ctx, params)
				suite.app.ClaimsKeeper.SetClaimsRecord(suite.ctx, addr, types.ClaimsRecord{})

				suite.app.ClaimsKeeper.AfterDelegationModified(suite.ctx, addr, addr2)
			},
		},
		{
			"claim enabled - already claimed",
			func() {
				params := suite.app.ClaimsKeeper.GetParams(suite.ctx)
				params.EnableClaims = true
				params.AirdropStartTime = time.Time{}
				params.DurationUntilDecay = time.Hour
				params.DurationOfDecay = time.Hour
				suite.app.ClaimsKeeper.SetParams(suite.ctx, params)
				claimedRecord := types.ClaimsRecord{
					InitialClaimableAmount: sdk.NewInt(1000),
					ActionsCompleted:       []bool{false, true, false, false},
				}
				suite.app.ClaimsKeeper.SetClaimsRecord(suite.ctx, addr, claimedRecord)

				suite.app.ClaimsKeeper.AfterDelegationModified(suite.ctx, addr, addr2)
			},
		},
		{
			"claim enabled - claim",
			func() {
				params := suite.app.ClaimsKeeper.GetParams(suite.ctx)
				params.EnableClaims = true
				params.AirdropStartTime = time.Now().UTC()
				suite.app.ClaimsKeeper.SetParams(suite.ctx, params)

				suite.app.ClaimsKeeper.SetClaimsRecord(suite.ctx, addr, claimRecord)

				suite.ctx = suite.ctx.WithBlockTime(time.Now().UTC().Add(time.Hour))
				suite.app.ClaimsKeeper.AfterDelegationModified(suite.ctx, addr, addr2)

				newClaimRec, found := suite.app.ClaimsKeeper.GetClaimsRecord(suite.ctx, addr)

				expectedClaimRecord := types.ClaimsRecord{
					InitialClaimableAmount: sdk.NewInt(1000),
					ActionsCompleted:       []bool{false, true, false, false},
				}
				suite.Require().True(found)
				suite.Require().Equal(expectedClaimRecord, newClaimRec)

				balances := suite.app.BankKeeper.GetAllBalances(suite.ctx, addr)
				claimedCoins := sdk.Coins{{Denom: params.ClaimsDenom, Amount: sdk.NewInt(250)}}
				suite.Require().Equal(claimedCoins, balances)
			},
		},
	}
	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			suite.SetupClaimTest() // reset

			suite.app.AccountKeeper.SetAccount(suite.ctx, authtypes.NewBaseAccount(addr, nil, 0, 0))
			tc.test()
		})
	}
}

func (suite *KeeperTestSuite) TestAfterEVMStateTransition() {
	from := tests.GenerateAddress()
	to := tests.GenerateAddress()
	receipt := ethtypes.Receipt{}
	addr := sdk.AccAddress(from.Bytes())
	claimRecord := types.ClaimsRecord{
		InitialClaimableAmount: sdk.NewInt(1000),
		ActionsCompleted:       []bool{false, false, false, false},
	}

	testCases := []struct {
		name string
		test func()
	}{
		{
			"no claim record",
			func() {
				err := suite.app.ClaimsKeeper.AfterEVMStateTransition(suite.ctx, from, &to, &receipt)
				suite.Require().NoError(err)
			},
		},
		{
			"claim disabled",
			func() {
				params := suite.app.ClaimsKeeper.GetParams(suite.ctx)
				params.EnableClaims = false
				suite.app.ClaimsKeeper.SetParams(suite.ctx, params)
				suite.app.ClaimsKeeper.SetClaimsRecord(suite.ctx, addr, types.ClaimsRecord{})

				err := suite.app.ClaimsKeeper.AfterEVMStateTransition(suite.ctx, from, &to, &receipt)
				suite.Require().NoError(err)
			},
		},
		{
			"claim enabled - no claim record",
			func() {
				params := suite.app.ClaimsKeeper.GetParams(suite.ctx)
				params.EnableClaims = true
				params.AirdropStartTime = time.Time{}
				suite.app.ClaimsKeeper.SetParams(suite.ctx, params)

				err := suite.app.ClaimsKeeper.AfterEVMStateTransition(suite.ctx, from, &to, &receipt)
				suite.Require().NoError(err)
			},
		},
		{
			"claim enabled - already claimed",
			func() {
				params := suite.app.ClaimsKeeper.GetParams(suite.ctx)
				params.EnableClaims = true
				params.AirdropStartTime = time.Now().UTC()
				suite.app.ClaimsKeeper.SetParams(suite.ctx, params)
				claimedRecord := types.ClaimsRecord{
					InitialClaimableAmount: sdk.NewInt(1000),
					ActionsCompleted:       []bool{false, false, true, false},
				}
				suite.app.ClaimsKeeper.SetClaimsRecord(suite.ctx, addr, claimedRecord)

				err := suite.app.ClaimsKeeper.AfterEVMStateTransition(suite.ctx, from, &to, &receipt)
				suite.Require().NoError(err)
			},
		},
		{
			"claim enabled - claim",
			func() {
				params := suite.app.ClaimsKeeper.GetParams(suite.ctx)
				params.EnableClaims = true
				params.AirdropStartTime = time.Now().UTC().UTC()
				params.DurationUntilDecay = time.Hour
				params.DurationOfDecay = time.Hour
				suite.app.ClaimsKeeper.SetParams(suite.ctx, params)

				suite.app.ClaimsKeeper.SetClaimsRecord(suite.ctx, addr, claimRecord)

				suite.ctx = suite.ctx.WithBlockTime(time.Now().UTC().Add(time.Hour))
				err := suite.app.ClaimsKeeper.AfterEVMStateTransition(suite.ctx, from, &to, &receipt)
				suite.Require().NoError(err)

				newClaimRec, found := suite.app.ClaimsKeeper.GetClaimsRecord(suite.ctx, addr)

				expectedClaimRecord := types.ClaimsRecord{
					InitialClaimableAmount: sdk.NewInt(1000),
					ActionsCompleted:       []bool{false, false, true, false},
				}
				suite.Require().True(found)
				suite.Require().Equal(expectedClaimRecord, newClaimRec)

				balances := suite.app.BankKeeper.GetAllBalances(suite.ctx, addr)
				claimedCoins := sdk.Coins{{Denom: params.ClaimsDenom, Amount: sdk.NewInt(250)}}
				suite.Require().Equal(claimedCoins, balances)
			},
		},
	}
	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			suite.SetupClaimTest() // reset

			suite.app.AccountKeeper.SetAccount(suite.ctx, authtypes.NewBaseAccount(addr, nil, 0, 0))
			tc.test()
		})
	}
}
