package keeper_test

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/evmos/ethermint/crypto/ethsecp256k1"
	"github.com/evmos/ethermint/tests"
	ethermint "github.com/evmos/ethermint/types"
	"github.com/evmos/evmos/v7/testutil"
	inflationtypes "github.com/evmos/evmos/v7/x/inflation/types"

	"github.com/evmos/evmos/v7/x/claims/types"
)

func (suite *KeeperTestSuite) TestGetClaimableAmountForAction() {
	testCases := []struct {
		name         string
		claimsRecord types.ClaimsRecord
		params       types.Params
		expAmt       sdk.Int
		expRemainder sdk.Int
	}{
		{
			"zero initial claimable amount",
			types.ClaimsRecord{InitialClaimableAmount: sdk.ZeroInt()},
			types.Params{},
			sdk.ZeroInt(),
			sdk.ZeroInt(),
		},
		{
			"claims not active",
			types.ClaimsRecord{InitialClaimableAmount: sdk.OneInt()},
			types.Params{},
			sdk.ZeroInt(),
			sdk.ZeroInt(),
		},
		{
			"action already claimed",
			types.ClaimsRecord{InitialClaimableAmount: sdk.OneInt(), ActionsCompleted: []bool{true, true, true, true}},
			types.Params{
				EnableClaims:     true,
				AirdropStartTime: suite.ctx.BlockTime(),
			},
			sdk.ZeroInt(),
			sdk.ZeroInt(),
		},
		{
			"before decay",
			types.NewClaimsRecord(sdk.NewInt(100)),
			types.Params{
				EnableClaims:       true,
				AirdropStartTime:   suite.ctx.BlockTime().Add(-time.Minute),
				DurationUntilDecay: time.Hour,
				DurationOfDecay:    time.Hour,
			},
			sdk.NewInt(25),
			sdk.ZeroInt(),
		},
		{
			"during decay",
			types.NewClaimsRecord(sdk.NewInt(200)),
			types.Params{
				EnableClaims:       true,
				AirdropStartTime:   suite.ctx.BlockTime().Add(-time.Hour),
				DurationUntilDecay: 30 * time.Minute,
				DurationOfDecay:    time.Hour,
				ClaimsDenom:        types.DefaultClaimsDenom,
			},
			sdk.NewInt(25),
			sdk.NewInt(25),
		},
		{
			"during decay - rounded",
			types.NewClaimsRecord(sdk.NewInt(100)),
			types.Params{
				EnableClaims:       true,
				AirdropStartTime:   suite.ctx.BlockTime().Add(-time.Hour),
				DurationUntilDecay: 30 * time.Minute,
				DurationOfDecay:    time.Hour,
				ClaimsDenom:        types.DefaultClaimsDenom,
			},
			sdk.NewInt(12),
			sdk.NewInt(13),
		},
	}

	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			suite.SetupTest() // reset

			action := types.ActionDelegate
			amt, remainder := suite.app.ClaimsKeeper.GetClaimableAmountForAction(suite.ctx, tc.claimsRecord, action, tc.params)
			suite.Require().Equal(tc.expAmt.Int64(), amt.Int64())
			suite.Require().Equal(tc.expRemainder.Int64(), remainder.Int64())
		})
	}
}

func (suite *KeeperTestSuite) TestGetUserTotalClaimable() {
	addr := sdk.AccAddress(tests.GenerateAddress().Bytes())

	testCases := []struct {
		name     string
		malleate func()
		expAmt   sdk.Int
	}{
		{
			"zero - no claim record",
			func() {},
			sdk.ZeroInt(),
		},
		{
			"zero - all actions completed",
			func() {
				cr := types.ClaimsRecord{InitialClaimableAmount: sdk.NewInt(100), ActionsCompleted: []bool{true, true, true, true}}
				suite.app.ClaimsKeeper.SetClaimsRecord(suite.ctx, addr, cr)
			},
			sdk.ZeroInt(),
		},
		{
			"all actions unclaimed, before decay",
			func() {
				cr := types.NewClaimsRecord(sdk.NewInt(100))
				params := suite.app.ClaimsKeeper.GetParams(suite.ctx)
				params.AirdropStartTime = suite.ctx.BlockTime().Add(-time.Minute)
				suite.app.ClaimsKeeper.SetParams(suite.ctx, params)
				suite.app.ClaimsKeeper.SetClaimsRecord(suite.ctx, addr, cr)
			},
			sdk.NewInt(100),
		},
		{
			"all actions unclaimed, claims inactive",
			func() {
				cr := types.NewClaimsRecord(sdk.NewInt(100))
				params := suite.app.ClaimsKeeper.GetParams(suite.ctx)
				params.EnableClaims = false
				suite.app.ClaimsKeeper.SetParams(suite.ctx, params)
				suite.app.ClaimsKeeper.SetClaimsRecord(suite.ctx, addr, cr)
			},
			sdk.ZeroInt(),
		},
		{
			"during decay",
			func() {
				cr := types.NewClaimsRecord(sdk.NewInt(200))
				params := suite.app.ClaimsKeeper.GetParams(suite.ctx)
				params.AirdropStartTime = params.AirdropStartTime.Add(-time.Hour)
				params.DurationUntilDecay = 30 * time.Minute
				params.DurationOfDecay = time.Hour
				suite.app.ClaimsKeeper.SetParams(suite.ctx, params)
				suite.app.ClaimsKeeper.SetClaimsRecord(suite.ctx, addr, cr)
			},
			sdk.NewInt(100),
		},
	}

	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			suite.SetupTest() // reset

			tc.malleate()

			amt := suite.getUserTotalClaimable(suite.ctx, addr)
			suite.Require().Equal(tc.expAmt.Int64(), amt.Int64())
		})
	}
}

func (suite *KeeperTestSuite) TestClaimCoinsForAction() {
	addr := sdk.AccAddress(tests.GenerateAddress().Bytes())

	testCases := []struct {
		name            string
		malleate        func()
		claimsRecord    types.ClaimsRecord
		action          types.Action
		params          types.Params
		expAmt          sdk.Int
		expClawedBack   sdk.Int
		expError        bool
		expDeleteRecord bool
	}{
		{
			"fail - unspecified action",
			func() {},
			types.ClaimsRecord{},
			types.ActionUnspecified,
			types.Params{},
			sdk.ZeroInt(),
			sdk.ZeroInt(),
			true,
			false,
		},
		{
			"zero - claims inactive",
			func() {},
			types.ClaimsRecord{},
			types.ActionDelegate,
			types.Params{},
			sdk.ZeroInt(),
			sdk.ZeroInt(),
			false,
			false,
		},
		{
			"zero - action claimed",
			func() {},
			types.ClaimsRecord{ActionsCompleted: []bool{true, false, false, false}},
			types.ActionVote,
			types.Params{
				EnableClaims:       true,
				AirdropStartTime:   suite.ctx.BlockTime().Add(-time.Hour),
				DurationUntilDecay: 2 * time.Hour,
				DurationOfDecay:    time.Hour,
				ClaimsDenom:        types.DefaultClaimsDenom,
			},
			sdk.ZeroInt(),
			sdk.ZeroInt(),
			false,
			false,
		},
		{
			"zero - claimable amount is 0",
			func() {},
			types.NewClaimsRecord(sdk.ZeroInt()),
			types.ActionVote,
			types.Params{
				EnableClaims:       true,
				AirdropStartTime:   suite.ctx.BlockTime().Add(-time.Hour),
				DurationUntilDecay: 2 * time.Hour,
				DurationOfDecay:    time.Hour,
				ClaimsDenom:        types.DefaultClaimsDenom,
			},
			sdk.ZeroInt(),
			sdk.ZeroInt(),
			false,
			false,
		},
		{
			"fail - error during transfer from module to account",
			func() {
				// drain the module account funds to test error
				addr := suite.app.ClaimsKeeper.GetModuleAccountAddress()
				coins := suite.app.BankKeeper.GetAllBalances(suite.ctx, addr)
				err := suite.app.BankKeeper.SendCoinsFromModuleToModule(suite.ctx, types.ModuleName, inflationtypes.ModuleName, coins)
				suite.Require().NoError(err)
			},
			types.NewClaimsRecord(sdk.NewInt(200)),
			types.ActionVote,
			types.Params{
				EnableClaims:       true,
				AirdropStartTime:   suite.ctx.BlockTime().Add(-time.Hour),
				DurationUntilDecay: 2 * time.Hour,
				DurationOfDecay:    time.Hour,
				ClaimsDenom:        types.DefaultClaimsDenom,
			},
			sdk.ZeroInt(),
			sdk.ZeroInt(),
			true,
			false,
		},
		{
			"failed - error during community pool fund",
			func() {
				coins := sdk.NewCoins(sdk.NewCoin(types.DefaultClaimsDenom, sdk.NewInt(25)))
				err := testutil.FundModuleAccount(suite.app.BankKeeper, suite.ctx, types.ModuleName, coins)
				suite.Require().NoError(err)
			},
			types.NewClaimsRecord(sdk.NewInt(200)),
			types.ActionVote,
			types.Params{
				EnableClaims:       true,
				AirdropStartTime:   suite.ctx.BlockTime().Add(-time.Hour),
				DurationUntilDecay: 30 * time.Minute,
				DurationOfDecay:    time.Hour,
				ClaimsDenom:        types.DefaultClaimsDenom,
			},
			sdk.ZeroInt(),
			sdk.ZeroInt(),
			true,
			false,
		},
		{
			"success - claim single action",
			func() {
				coins := sdk.NewCoins(sdk.NewCoin(types.DefaultClaimsDenom, sdk.NewInt(50)))
				err := testutil.FundModuleAccount(suite.app.BankKeeper, suite.ctx, types.ModuleName, coins)
				suite.Require().NoError(err)
			},
			types.NewClaimsRecord(sdk.NewInt(200)),
			types.ActionVote,
			types.Params{
				EnableClaims:       true,
				AirdropStartTime:   suite.ctx.BlockTime().Add(-time.Hour),
				DurationUntilDecay: 2 * time.Hour,
				DurationOfDecay:    time.Hour,
				ClaimsDenom:        types.DefaultClaimsDenom,
			},
			sdk.NewInt(50),
			sdk.ZeroInt(),
			false,
			false,
		},
		{
			"success - claim single action during decay",
			func() {
				coins := sdk.NewCoins(sdk.NewCoin(types.DefaultClaimsDenom, sdk.NewInt(50)))
				err := testutil.FundModuleAccount(suite.app.BankKeeper, suite.ctx, types.ModuleName, coins)
				suite.Require().NoError(err)
			},
			types.NewClaimsRecord(sdk.NewInt(200)),
			types.ActionVote,
			types.Params{
				EnableClaims:       true,
				AirdropStartTime:   suite.ctx.BlockTime().Add(-time.Hour),
				DurationUntilDecay: 30 * time.Minute,
				DurationOfDecay:    time.Hour,
				ClaimsDenom:        types.DefaultClaimsDenom,
			},
			sdk.NewInt(25),
			sdk.NewInt(25),
			false,
			false,
		},
		{
			"success - claimed all actions - with VOTE as last action",
			func() {
				coins := sdk.NewCoins(sdk.NewCoin(types.DefaultClaimsDenom, sdk.NewInt(50)))
				err := testutil.FundModuleAccount(suite.app.BankKeeper, suite.ctx, types.ModuleName, coins)
				suite.Require().NoError(err)
			},
			types.ClaimsRecord{
				InitialClaimableAmount: sdk.NewInt(200),
				ActionsCompleted:       []bool{false, true, true, true},
			},
			types.ActionVote,
			types.Params{
				EnableClaims:       true,
				AirdropStartTime:   suite.ctx.BlockTime().Add(-time.Hour),
				DurationUntilDecay: 2 * time.Hour,
				DurationOfDecay:    time.Hour,
				ClaimsDenom:        types.DefaultClaimsDenom,
			},
			sdk.NewInt(50),
			sdk.ZeroInt(),
			false,
			false,
		},
		{
			"success - claimed all actions - with IBC as last action",
			func() {
				coins := sdk.NewCoins(sdk.NewCoin(types.DefaultClaimsDenom, sdk.NewInt(50)))
				err := testutil.FundModuleAccount(suite.app.BankKeeper, suite.ctx, types.ModuleName, coins)
				suite.Require().NoError(err)
			},
			types.ClaimsRecord{
				InitialClaimableAmount: sdk.NewInt(200),
				ActionsCompleted:       []bool{true, true, true, false},
			},
			types.ActionIBCTransfer,
			types.Params{
				EnableClaims:       true,
				AirdropStartTime:   suite.ctx.BlockTime().Add(-time.Hour),
				DurationUntilDecay: 2 * time.Hour,
				DurationOfDecay:    time.Hour,
				ClaimsDenom:        types.DefaultClaimsDenom,
			},
			sdk.NewInt(50),
			sdk.ZeroInt(),
			false,
			false,
		},
	}

	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			suite.SetupTest() // reset

			tc.malleate()

			initialBalance := suite.app.BankKeeper.GetBalance(suite.ctx, addr, types.DefaultClaimsDenom)
			initialCommunityPoolCoins := suite.app.DistrKeeper.GetFeePoolCommunityCoins(suite.ctx)

			amt, err := suite.app.ClaimsKeeper.ClaimCoinsForAction(suite.ctx, addr, tc.claimsRecord, tc.action, tc.params)
			if tc.expError {
				suite.Require().Error(err)
				suite.Require().Equal(int64(0), amt.Int64())
				return
			}

			suite.Require().NoError(err)
			suite.Require().Equal(tc.expAmt.Int64(), amt.Int64())
			if amt.IsZero() {
				return
			}

			expBalance := initialBalance.Add(sdk.NewCoin(types.DefaultClaimsDenom, amt))
			postClaimBalance := suite.app.BankKeeper.GetBalance(suite.ctx, addr, types.DefaultClaimsDenom)
			suite.Require().Equal(expBalance, postClaimBalance)

			if !tc.expClawedBack.IsZero() {
				initialCommunityPoolCoins = initialCommunityPoolCoins.Add(sdk.NewDecCoin(tc.params.ClaimsDenom, tc.expClawedBack))
				funds := suite.app.DistrKeeper.GetFeePoolCommunityCoins(suite.ctx)
				suite.Require().Equal(initialCommunityPoolCoins.String(), funds.String())
			}

			if tc.expDeleteRecord {
				suite.Require().False(suite.app.ClaimsKeeper.HasClaimsRecord(suite.ctx, addr))
			} else {
				cr, found := suite.app.ClaimsKeeper.GetClaimsRecord(suite.ctx, addr)
				suite.Require().True(found)
				suite.Require().True(cr.HasClaimedAction(tc.action))
			}
		})
	}
}

func (suite *KeeperTestSuite) TestMergeClaimRecords() {
	recipient := sdk.AccAddress(tests.GenerateAddress().Bytes())

	params := types.Params{
		EnableClaims:       true,
		AirdropStartTime:   suite.ctx.BlockTime().Add(-time.Hour),
		DurationUntilDecay: 2 * time.Hour,
		DurationOfDecay:    time.Hour,
		ClaimsDenom:        types.DefaultClaimsDenom,
	}

	testCases := []struct {
		name string
		test func()
	}{
		{
			"case 4: actions not completed - during decay",
			func() {
				senderClaimsRecord := types.NewClaimsRecord(sdk.NewInt(200))
				recipientClaimsRecord := types.NewClaimsRecord(sdk.NewInt(200))

				expBalance := suite.app.BankKeeper.GetBalance(suite.ctx, recipient, params.ClaimsDenom)
				initialCommunityPoolCoins := suite.app.DistrKeeper.GetFeePoolCommunityCoins(suite.ctx)

				coins := sdk.Coins{sdk.NewCoin(params.ClaimsDenom, sdk.NewInt(100))}
				err := testutil.FundModuleAccount(suite.app.BankKeeper, suite.ctx, types.ModuleName, coins)
				suite.Require().NoError(err)

				params := types.Params{
					EnableClaims:       true,
					AirdropStartTime:   suite.ctx.BlockTime().Add(-time.Hour),
					DurationUntilDecay: 30 * time.Minute,
					DurationOfDecay:    time.Hour,
					ClaimsDenom:        types.DefaultClaimsDenom,
				}

				mergedRecord, err := suite.app.ClaimsKeeper.MergeClaimsRecords(suite.ctx, recipient, senderClaimsRecord, recipientClaimsRecord, params)
				suite.Require().NoError(err)

				expectedRecord := types.ClaimsRecord{
					InitialClaimableAmount: sdk.NewInt(400),
					ActionsCompleted:       []bool{false, false, false, true},
				}

				suite.Require().Equal(expectedRecord, mergedRecord)

				initialCommunityPoolCoins = initialCommunityPoolCoins.Add(sdk.NewDecCoin(params.ClaimsDenom, sdk.NewInt(50)))
				funds := suite.app.DistrKeeper.GetFeePoolCommunityCoins(suite.ctx)
				suite.Require().Equal(initialCommunityPoolCoins.String(), funds.String())

				expBalance = expBalance.Add(sdk.NewCoin(params.ClaimsDenom, sdk.NewInt(50)))
				balance := suite.app.BankKeeper.GetBalance(suite.ctx, recipient, params.ClaimsDenom)
				suite.Require().Equal(expBalance.String(), balance.String())
			},
		},
		{
			"case 4: actions not completed",
			func() {
				senderClaimsRecord := types.NewClaimsRecord(sdk.NewInt(200))
				recipientClaimsRecord := types.NewClaimsRecord(sdk.NewInt(200))

				expBalance := suite.app.BankKeeper.GetBalance(suite.ctx, recipient, params.ClaimsDenom)

				coins := sdk.Coins{sdk.NewCoin(params.ClaimsDenom, sdk.NewInt(100))}
				err := testutil.FundModuleAccount(suite.app.BankKeeper, suite.ctx, types.ModuleName, coins)
				suite.Require().NoError(err)

				mergedRecord, err := suite.app.ClaimsKeeper.MergeClaimsRecords(suite.ctx, recipient, senderClaimsRecord, recipientClaimsRecord, params)
				suite.Require().NoError(err)

				// only IBC action should be claimed
				expectedRecord := types.NewClaimsRecord(sdk.NewInt(400))
				expectedRecord.MarkClaimed(types.ActionIBCTransfer)
				suite.Require().Equal(expectedRecord, mergedRecord)

				expBalance = expBalance.Add(sdk.NewCoin(params.ClaimsDenom, sdk.NewInt(100)))

				balance := suite.app.BankKeeper.GetBalance(suite.ctx, recipient, params.ClaimsDenom)
				suite.Require().Equal(expBalance, balance)
			},
		},
		{
			"case 2: recipient completed all actions, but IBC transfer",
			func() {
				senderClaimsRecord := types.ClaimsRecord{
					InitialClaimableAmount: sdk.NewInt(200),
					ActionsCompleted:       []bool{false, false, false, false},
				}
				recipientClaimsRecord := types.ClaimsRecord{
					InitialClaimableAmount: sdk.NewInt(200),
					ActionsCompleted:       []bool{true, true, true, false},
				}

				expBalance := suite.app.BankKeeper.GetBalance(suite.ctx, recipient, params.ClaimsDenom)

				coins := sdk.Coins{sdk.NewCoin(params.ClaimsDenom, sdk.NewInt(250))}
				err := testutil.FundModuleAccount(suite.app.BankKeeper, suite.ctx, types.ModuleName, coins)
				suite.Require().NoError(err)

				mergedRecord, err := suite.app.ClaimsKeeper.MergeClaimsRecords(suite.ctx, recipient, senderClaimsRecord, recipientClaimsRecord, params)
				suite.Require().NoError(err)

				expectedRecord := types.ClaimsRecord{
					InitialClaimableAmount: sdk.NewInt(400),
					ActionsCompleted:       []bool{true, true, true, true},
				}

				suite.Require().Equal(expectedRecord, mergedRecord)

				expBalance = expBalance.Add(sdk.NewCoin(params.ClaimsDenom, sdk.NewInt(250)))

				balance := suite.app.BankKeeper.GetBalance(suite.ctx, recipient, params.ClaimsDenom)
				suite.Require().Equal(expBalance, balance)
			},
		},
		{
			"fail - sender and recipient completed all",
			func() {
				senderClaimsRecord := types.ClaimsRecord{
					InitialClaimableAmount: sdk.NewInt(200),
					ActionsCompleted:       []bool{true, true, true, true},
				}
				recipientClaimsRecord := types.ClaimsRecord{
					InitialClaimableAmount: sdk.NewInt(200),
					ActionsCompleted:       []bool{true, true, true, true},
				}

				_, err := suite.app.ClaimsKeeper.MergeClaimsRecords(suite.ctx, recipient, senderClaimsRecord, recipientClaimsRecord, params)
				suite.Require().Error(err)
			},
		},
		{
			"fail: error when transferring from escrow account",
			func() {
				senderClaimsRecord := types.NewClaimsRecord(sdk.NewInt(200))
				recipientClaimsRecord := types.NewClaimsRecord(sdk.NewInt(200))

				_, err := suite.app.ClaimsKeeper.MergeClaimsRecords(suite.ctx, recipient, senderClaimsRecord, recipientClaimsRecord, params)
				suite.Require().Error(err)
			},
		},
		{
			"fail: not enough funds to transfer to community pool",
			func() {
				senderClaimsRecord := types.NewClaimsRecord(sdk.NewInt(200))
				recipientClaimsRecord := types.NewClaimsRecord(sdk.NewInt(200))

				params := types.Params{
					EnableClaims:       true,
					AirdropStartTime:   suite.ctx.BlockTime().Add(-time.Hour),
					DurationUntilDecay: 30 * time.Minute,
					DurationOfDecay:    time.Hour,
					ClaimsDenom:        types.DefaultClaimsDenom,
				}
				suite.app.ClaimsKeeper.SetParams(suite.ctx, params)

				coins := sdk.Coins{sdk.NewCoin(params.ClaimsDenom, sdk.NewInt(50))}
				err := testutil.FundModuleAccount(suite.app.BankKeeper, suite.ctx, types.ModuleName, coins)
				suite.Require().NoError(err)

				_, err = suite.app.ClaimsKeeper.MergeClaimsRecords(suite.ctx, recipient, senderClaimsRecord, recipientClaimsRecord, params)
				suite.Require().Error(err)
			},
		},
	}

	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			suite.SetupTest() // reset

			suite.app.ClaimsKeeper.SetParams(suite.ctx, params)

			tc.test()
		})
	}
}

func (suite *KeeperTestSuite) TestHookOfUnclaimableAccount() {
	suite.SetupTestWithEscrow()
	addr1 := sdk.AccAddress(tests.GenerateAddress().Bytes())
	suite.app.AccountKeeper.SetAccount(suite.ctx, authtypes.NewBaseAccount(addr1, nil, 0, 0))

	params := suite.app.ClaimsKeeper.GetParams(suite.ctx)

	claim, found := suite.app.ClaimsKeeper.GetClaimsRecord(suite.ctx, addr1)
	suite.Require().False(found)
	suite.Require().Equal(types.ClaimsRecord{}, claim)

	_, err := suite.app.ClaimsKeeper.ClaimCoinsForAction(suite.ctx, addr1, claim, types.ActionEVM, params)
	suite.Require().NoError(err)

	balances := suite.app.BankKeeper.GetAllBalances(suite.ctx, addr1)
	suite.Require().Equal(sdk.Coins{}, balances)
}

func (suite *KeeperTestSuite) TestHookBeforeAirdropStart() {
	suite.SetupTestWithEscrow()

	airdropStartTime := suite.ctx.BlockTime().Add(time.Hour)
	params := suite.app.ClaimsKeeper.GetParams(suite.ctx)
	params.AirdropStartTime = airdropStartTime
	suite.app.ClaimsKeeper.SetParams(suite.ctx, params)

	addr1 := sdk.AccAddress(tests.GenerateAddress().Bytes())

	claimsRecord := types.NewClaimsRecord(sdk.NewInt(1000))
	acc := suite.app.AccountKeeper.NewAccountWithAddress(suite.ctx, addr1)
	suite.app.AccountKeeper.SetAccount(suite.ctx, acc)
	suite.app.ClaimsKeeper.SetClaimsRecord(suite.ctx, addr1, claimsRecord)

	coins := suite.getUserTotalClaimable(suite.ctx, addr1)
	suite.Require().Equal(sdk.ZeroInt().String(), coins.String())

	coins, remainder := suite.app.ClaimsKeeper.GetClaimableAmountForAction(suite.ctx, claimsRecord, types.ActionVote, params)
	suite.Require().Equal(sdk.ZeroInt().String(), coins.String())
	suite.Require().Equal(sdk.ZeroInt().String(), remainder.String())

	claimedAmount, err := suite.app.ClaimsKeeper.ClaimCoinsForAction(suite.ctx, addr1, claimsRecord, types.ActionVote, params)
	suite.Require().NoError(err)
	suite.Require().Equal(coins.Int64(), claimedAmount.Int64())

	balances := suite.app.BankKeeper.GetAllBalances(suite.ctx, addr1)

	// Now, it is before starting air drop, so claim module should not send the balances to the user
	suite.Require().True(balances.IsZero(), balances.String())

	suite.ctx = suite.ctx.WithBlockTime(suite.ctx.BlockTime().Add(2 * time.Hour))

	_, err = suite.app.ClaimsKeeper.ClaimCoinsForAction(suite.ctx.WithBlockTime(airdropStartTime), addr1, claimsRecord, types.ActionVote, params)
	suite.Require().NoError(err)

	balances = suite.app.BankKeeper.GetAllBalances(suite.ctx, addr1)
	// Now, it is the time for air drop, so claim module should send the balances to the user
	suite.Require().Equal(claimsRecord.InitialClaimableAmount.Quo(sdk.NewInt(4)), balances.AmountOf(params.ClaimsDenom))
}

func (suite *KeeperTestSuite) TestHookAfterAirdropEnd() {
	suite.SetupTestWithEscrow()

	// airdrop recipient address
	addr1 := sdk.AccAddress(tests.GenerateAddress().Bytes())

	claimsRecord := types.ClaimsRecord{
		InitialClaimableAmount: sdk.NewInt(1000),
		ActionsCompleted:       []bool{false, false, false, false},
	}

	suite.app.AccountKeeper.SetAccount(suite.ctx, authtypes.NewBaseAccount(addr1, nil, 0, 0))
	suite.app.ClaimsKeeper.SetClaimsRecord(suite.ctx, addr1, claimsRecord)

	params := suite.app.ClaimsKeeper.GetParams(suite.ctx)
	suite.ctx = suite.ctx.WithBlockTime(params.AirdropStartTime.Add(params.DurationUntilDecay).Add(params.DurationOfDecay))

	err := suite.app.ClaimsKeeper.EndAirdrop(suite.ctx, params)
	suite.Require().NoError(err)

	_, err = suite.app.ClaimsKeeper.ClaimCoinsForAction(suite.ctx, addr1, claimsRecord, types.ActionDelegate, params)
	suite.Require().NoError(err)
}

func (suite *KeeperTestSuite) TestDuplicatedActionNotWithdrawRepeatedly() {
	suite.SetupTestWithEscrow()
	addr1 := sdk.AccAddress(tests.GenerateAddress().Bytes())

	params := suite.app.ClaimsKeeper.GetParams(suite.ctx)

	claimsRecord := types.ClaimsRecord{
		InitialClaimableAmount: sdk.NewInt(1000),
		ActionsCompleted:       []bool{false, false, false, false},
	}

	suite.app.AccountKeeper.SetAccount(suite.ctx, authtypes.NewBaseAccount(addr1, nil, 0, 0))

	suite.app.ClaimsKeeper.SetClaimsRecord(suite.ctx, addr1, claimsRecord)

	coins1 := suite.getUserTotalClaimable(suite.ctx, addr1)
	suite.Require().Equal(coins1, claimsRecord.InitialClaimableAmount)

	_, err := suite.app.ClaimsKeeper.ClaimCoinsForAction(suite.ctx, addr1, claimsRecord, types.ActionEVM, params)
	suite.Require().NoError(err)

	claim, found := suite.app.ClaimsKeeper.GetClaimsRecord(suite.ctx, addr1)
	suite.Require().True(found)
	suite.Require().True(claim.ActionsCompleted[types.ActionEVM-1])
	claimedCoins := suite.app.BankKeeper.GetAllBalances(suite.ctx, addr1)
	suite.Require().Equal(claimedCoins.AmountOf(params.GetClaimsDenom()), claimsRecord.InitialClaimableAmount.Quo(sdk.NewInt(4)))

	_, err = suite.app.ClaimsKeeper.ClaimCoinsForAction(suite.ctx, addr1, claimsRecord, types.ActionEVM, params)

	suite.NoError(err)
	suite.True(claim.ActionsCompleted[types.ActionEVM-1])
	claimedCoins = suite.app.BankKeeper.GetAllBalances(suite.ctx, addr1)
	suite.Require().Equal(claimedCoins.AmountOf(params.GetClaimsDenom()), claimsRecord.InitialClaimableAmount.Quo(sdk.NewInt(4)))
}

func (suite *KeeperTestSuite) TestDelegationAutoWithdrawAndDelegateMore() {
	suite.SetupTestWithEscrow()

	pub1, _ := ethsecp256k1.GenerateKey()
	pub2, _ := ethsecp256k1.GenerateKey()
	addrs := []sdk.AccAddress{sdk.AccAddress(pub1.PubKey().Address()), sdk.AccAddress(pub2.PubKey().Address())}
	params := suite.app.ClaimsKeeper.GetParams(suite.ctx)

	claimsRecords := []types.ClaimsRecord{
		{
			InitialClaimableAmount: sdk.NewInt(1000),
			ActionsCompleted:       []bool{false, false, false, false},
		},
		{
			InitialClaimableAmount: sdk.NewInt(1000),
			ActionsCompleted:       []bool{false, false, false, false},
		},
	}

	// initialize accts
	for i := 0; i < len(addrs); i++ {
		suite.app.AccountKeeper.SetAccount(suite.ctx, authtypes.NewBaseAccount(addrs[i], nil, 0, 0))
		suite.app.ClaimsKeeper.SetClaimsRecord(suite.ctx, addrs[i], claimsRecords[i])
	}

	// test claim records set
	for i := 0; i < len(addrs); i++ {
		coins := suite.getUserTotalClaimable(suite.ctx, addrs[i])
		suite.Require().Equal(coins, claimsRecords[i].InitialClaimableAmount)
	}

	// set addr[0] as a validator
	validator, err := stakingtypes.NewValidator(sdk.ValAddress(addrs[0]), pub1.PubKey(), stakingtypes.Description{})
	suite.Require().NoError(err)
	validator = stakingkeeper.TestingUpdateValidator(suite.app.StakingKeeper, suite.ctx, validator, true)
	suite.app.StakingKeeper.AfterValidatorCreated(suite.ctx, validator.GetOperator())

	validator, _ = validator.AddTokensFromDel(sdk.TokensFromConsensusPower(1, ethermint.PowerReduction))
	delAmount := sdk.TokensFromConsensusPower(1, ethermint.PowerReduction)

	err = testutil.FundAccount(suite.app.BankKeeper, suite.ctx, addrs[1], sdk.NewCoins(sdk.NewCoin(params.GetClaimsDenom(), delAmount)))
	suite.Require().NoError(err)

	_, err = suite.app.StakingKeeper.Delegate(suite.ctx, addrs[1], delAmount, stakingtypes.Unbonded, validator, true)
	suite.Require().NoError(err)

	// delegation should automatically call claim and withdraw balance
	actualClaimedCoins := suite.app.BankKeeper.GetAllBalances(suite.ctx, addrs[1])
	actualClaimedCoin := actualClaimedCoins.AmountOf(params.GetClaimsDenom())
	expectedClaimedCoin := claimsRecords[1].InitialClaimableAmount.Quo(sdk.NewInt(int64(len(claimsRecords[1].ActionsCompleted))))
	suite.Require().Equal(expectedClaimedCoin.String(), actualClaimedCoin.String())

	_, err = suite.app.StakingKeeper.Delegate(suite.ctx, addrs[1], actualClaimedCoin, stakingtypes.Unbonded, validator, true)
	suite.NoError(err)
}

func (suite *KeeperTestSuite) TestAirdropFlow() {
	suite.SetupTestWithEscrow()

	addrs := []sdk.AccAddress{
		sdk.AccAddress(tests.GenerateAddress().Bytes()),
		sdk.AccAddress(tests.GenerateAddress().Bytes()),
	}

	claimsRecords := []types.ClaimsRecord{
		{
			InitialClaimableAmount: sdk.NewInt(100),
			ActionsCompleted:       []bool{false, false, false, false},
		},
		{
			InitialClaimableAmount: sdk.NewInt(200),
			ActionsCompleted:       []bool{false, false, false, false},
		},
	}

	params := suite.app.ClaimsKeeper.GetParams(suite.ctx)

	// initialize accts
	for i := 0; i < len(addrs); i++ {
		suite.app.AccountKeeper.SetAccount(suite.ctx, authtypes.NewBaseAccount(addrs[i], nil, 0, 0))
		suite.app.ClaimsKeeper.SetClaimsRecord(suite.ctx, addrs[i], claimsRecords[i])
	}

	coins1 := suite.getUserTotalClaimable(suite.ctx, addrs[0])
	suite.Require().Equal(coins1, claimsRecords[0].InitialClaimableAmount)

	coins2 := suite.getUserTotalClaimable(suite.ctx, addrs[1])
	suite.Require().Equal(coins2, claimsRecords[1].InitialClaimableAmount)

	coins3 := suite.getUserTotalClaimable(suite.ctx, sdk.AccAddress(tests.GenerateAddress().Bytes()))
	suite.Require().True(coins3.IsZero())

	// get rewards amount per action
	coins4, remainder := suite.app.ClaimsKeeper.GetClaimableAmountForAction(suite.ctx, claimsRecords[0], types.ActionDelegate, suite.app.ClaimsKeeper.GetParams(suite.ctx))
	suite.Require().Equal(sdk.NewCoins(sdk.NewInt64Coin(params.GetClaimsDenom(), 25)).AmountOf(params.GetClaimsDenom()), coins4) // 2 = 10.Quo(4)
	suite.Require().Equal(sdk.ZeroInt(), remainder)

	// get completed activities
	claimsRecord, found := suite.app.ClaimsKeeper.GetClaimsRecord(suite.ctx, addrs[0])
	suite.Require().True(found)

	for i := 0; i < len(claimsRecord.ActionsCompleted); i++ {
		suite.Require().False(claimsRecord.ActionsCompleted[i])
	}

	// do half of actions
	_, err := suite.app.ClaimsKeeper.ClaimCoinsForAction(suite.ctx, addrs[0], claimsRecord, types.ActionEVM, params)
	suite.Require().NoError(err)
	_, err = suite.app.ClaimsKeeper.ClaimCoinsForAction(suite.ctx, addrs[0], claimsRecord, types.ActionDelegate, params)
	suite.Require().NoError(err)

	// check that half are completed
	claimsRecord, found = suite.app.ClaimsKeeper.GetClaimsRecord(suite.ctx, addrs[0])
	suite.Require().True(found)

	suite.Require().True(claimsRecord.HasClaimedAction(types.ActionEVM)) // We have Unspecified action in 0
	suite.Require().True(claimsRecord.HasClaimedAction(types.ActionDelegate))

	// get balance after 2 actions done
	bal1 := suite.app.BankKeeper.GetAllBalances(suite.ctx, addrs[0])
	suite.Require().Equal(bal1.String(), sdk.NewCoins(sdk.NewInt64Coin(params.GetClaimsDenom(), 50)).String())

	// check that claimable for completed activity is 0
	claimsRecord1, _ := suite.app.ClaimsKeeper.GetClaimsRecord(suite.ctx, addrs[0])
	bal4, remainder := suite.app.ClaimsKeeper.GetClaimableAmountForAction(suite.ctx, claimsRecord1, types.ActionEVM, params)
	suite.Require().Equal(sdk.ZeroInt(), bal4)
	suite.Require().Equal(sdk.ZeroInt(), remainder)

	// do rest of actions
	_, err = suite.app.ClaimsKeeper.ClaimCoinsForAction(suite.ctx, addrs[0], claimsRecord, types.ActionIBCTransfer, params)
	suite.Require().NoError(err)
	_, err = suite.app.ClaimsKeeper.ClaimCoinsForAction(suite.ctx, addrs[0], claimsRecord, types.ActionVote, params)
	suite.Require().NoError(err)

	// get balance after rest actions done
	bal1 = suite.app.BankKeeper.GetAllBalances(suite.ctx, addrs[0])
	suite.Require().Equal(bal1.AmountOf(params.GetClaimsDenom()), sdk.NewInt(100))

	// get claimable after withdrawing all
	coins1 = suite.getUserTotalClaimable(suite.ctx, addrs[0])
	suite.Require().NoError(err)
	suite.Require().True(coins1.IsZero())

	err = suite.app.ClaimsKeeper.EndAirdrop(suite.ctx, suite.app.ClaimsKeeper.GetParams(suite.ctx))
	suite.Require().NoError(err)

	moduleAccAddr := suite.app.AccountKeeper.GetModuleAddress(types.ModuleName)
	coins := suite.app.BankKeeper.GetBalance(suite.ctx, moduleAccAddr, params.GetClaimsDenom())
	suite.Require().Equal(coins, sdk.NewInt64Coin(params.GetClaimsDenom(), 0))

	coins2 = suite.getUserTotalClaimable(suite.ctx, addrs[1])
	suite.Require().NoError(err)
	suite.Require().Equal(coins2, sdk.ZeroInt())
}

func (suite *KeeperTestSuite) TestClaimOfDecayed() {
	suite.SetupTestWithEscrow()

	airdropStartTime := time.Now().UTC()
	durationUntilDecay := time.Hour
	durationOfDecay := time.Hour * 4

	addr1 := sdk.AccAddress(tests.GenerateAddress().Bytes())

	params := suite.app.ClaimsKeeper.GetParams(suite.ctx)
	params.AirdropStartTime = airdropStartTime
	params.DurationUntilDecay = durationUntilDecay
	params.DurationOfDecay = durationOfDecay
	suite.app.ClaimsKeeper.SetParams(suite.ctx, params)

	var claimsRecord types.ClaimsRecord

	t := []struct {
		fn func()
	}{
		{
			fn: func() {
				ctx := suite.ctx.WithBlockTime(airdropStartTime)

				coins, remainder := suite.app.ClaimsKeeper.GetClaimableAmountForAction(ctx, claimsRecord, types.ActionEVM, params)
				suite.Require().Equal(claimsRecord.InitialClaimableAmount.Quo(sdk.NewInt(4)).String(), coins.String())
				suite.Require().Equal(sdk.ZeroInt(), remainder)

				_, err := suite.app.ClaimsKeeper.ClaimCoinsForAction(suite.ctx, addr1, claimsRecord, types.ActionEVM, params)
				suite.Require().NoError(err)
				bal := suite.app.BankKeeper.GetAllBalances(ctx, addr1)
				suite.Require().Equal(claimsRecord.InitialClaimableAmount.Quo(sdk.NewInt(4)).String(), bal.AmountOf(params.GetClaimsDenom()).String())
			},
		},
		{
			fn: func() {
				ctx := suite.ctx.WithBlockTime(airdropStartTime.Add(durationUntilDecay))

				coins, remainder := suite.app.ClaimsKeeper.GetClaimableAmountForAction(ctx, claimsRecord, types.ActionEVM, params)
				suite.Require().Equal(claimsRecord.InitialClaimableAmount.Quo(sdk.NewInt(4)).String(), coins.String())
				suite.Require().Equal(sdk.ZeroInt(), remainder)

				_, err := suite.app.ClaimsKeeper.ClaimCoinsForAction(suite.ctx, addr1, claimsRecord, types.ActionEVM, params)
				suite.Require().NoError(err)
				bal := suite.app.BankKeeper.GetAllBalances(ctx, addr1)
				suite.Require().Equal(claimsRecord.InitialClaimableAmount.Quo(sdk.NewInt(4)).String(), bal.AmountOf(params.GetClaimsDenom()).String())
			},
		},
		{
			fn: func() {
				blockTime := airdropStartTime.Add(durationUntilDecay).Add(durationOfDecay / 2)
				elapsedAirdropTime := blockTime.Sub(airdropStartTime)
				decayTime := elapsedAirdropTime - durationUntilDecay
				decayPercent := sdk.NewDec(decayTime.Nanoseconds()).QuoInt64(durationOfDecay.Nanoseconds())
				claimablePercent := sdk.OneDec().Sub(decayPercent)

				ctx := suite.ctx.WithBlockTime(blockTime)

				coins, remainder := suite.app.ClaimsKeeper.GetClaimableAmountForAction(ctx, claimsRecord, types.ActionEVM, params)

				suite.Require().Equal(claimsRecord.InitialClaimableAmount.ToDec().Mul(claimablePercent).Quo(sdk.NewDec(4)).RoundInt().String(), coins.String())
				suite.Require().Equal(sdk.NewInt(13).String(), remainder.String())

				_, err := suite.app.ClaimsKeeper.ClaimCoinsForAction(ctx, addr1, claimsRecord, types.ActionEVM, params)
				suite.Require().NoError(err)
				bal := suite.app.BankKeeper.GetAllBalances(ctx, addr1)

				suite.Require().Equal(claimsRecord.InitialClaimableAmount.ToDec().Mul(claimablePercent).Quo(sdk.NewDec(4)).RoundInt().String(),
					bal.AmountOf(params.GetClaimsDenom()).String())
			},
		},
		{
			fn: func() {
				ctx := suite.ctx.WithBlockTime(airdropStartTime.Add(durationUntilDecay).Add(durationOfDecay))
				_, err := suite.app.ClaimsKeeper.ClaimCoinsForAction(ctx, addr1, claimsRecord, types.ActionEVM, params)
				suite.Require().NoError(err)
				bal := suite.app.BankKeeper.GetAllBalances(ctx, addr1)
				suite.Require().True(bal.Empty())
			},
		},
	}

	for _, test := range t {
		suite.SetupTestWithEscrow()

		claimsRecord = types.ClaimsRecord{
			InitialClaimableAmount: sdk.NewInt(100),
			ActionsCompleted:       []bool{false, false, false, false},
		}

		suite.app.ClaimsKeeper.SetParams(suite.ctx, types.Params{
			AirdropStartTime:   airdropStartTime,
			DurationUntilDecay: durationUntilDecay,
			DurationOfDecay:    durationOfDecay,
			EnableClaims:       true,
			ClaimsDenom:        params.GetClaimsDenom(),
		})

		suite.app.AccountKeeper.SetAccount(suite.ctx, authtypes.NewBaseAccount(addr1, nil, 0, 0))
		suite.app.ClaimsKeeper.SetClaimsRecord(suite.ctx, addr1, claimsRecord)

		test.fn()
	}
}

func (suite *KeeperTestSuite) TestClawbackEscrowedTokens() {
	suite.SetupTestWithEscrow()
	params := suite.app.ClaimsKeeper.GetParams(suite.ctx)

	ctx := suite.ctx.WithBlockTime(params.GetAirdropStartTime())

	escrow := sdk.NewInt(10000000)
	distrModuleAddr := suite.app.AccountKeeper.GetModuleAddress(distrtypes.ModuleName)

	// ensure claim is enabled
	suite.Require().True(params.EnableClaims)

	// ensure module account has the escrow fund
	coins := suite.app.ClaimsKeeper.GetModuleAccountBalances(ctx)
	suite.Require().Equal(coins.AmountOf(params.GetClaimsDenom()), escrow)

	// ensure community pool doesn't have the fund
	bal := suite.app.BankKeeper.GetBalance(ctx, distrModuleAddr, params.GetClaimsDenom())
	suite.Require().Equal(bal.Amount, sdk.ZeroInt())

	// claim some amount from airdrop
	addr1 := sdk.AccAddress(tests.GenerateAddress().Bytes())
	initClaim := sdk.NewInt(1000)
	claimsRecord := types.ClaimsRecord{
		InitialClaimableAmount: initClaim,
		ActionsCompleted:       []bool{false, false, false, false},
	}
	suite.app.AccountKeeper.SetAccount(ctx, authtypes.NewBaseAccount(addr1, nil, 0, 1))
	suite.app.ClaimsKeeper.SetClaimsRecord(ctx, addr1, claimsRecord)
	claimedCoins, err := suite.app.ClaimsKeeper.ClaimCoinsForAction(ctx, addr1, claimsRecord, types.ActionEVM, params)
	suite.Require().NoError(err)
	coins = suite.app.ClaimsKeeper.GetModuleAccountBalances(ctx)
	suite.Require().Equal(coins.AmountOf(params.GetClaimsDenom()), escrow.Sub(claimedCoins))

	// End the airdrop
	suite.app.ClaimsKeeper.EndAirdrop(ctx, params)

	// Make sure no one can claim after airdrop ends
	claimedCoinsAfter, err := suite.app.ClaimsKeeper.ClaimCoinsForAction(ctx, addr1, claimsRecord, types.ActionDelegate, params)
	suite.Require().Error(err)
	suite.Require().Equal(claimedCoinsAfter, sdk.ZeroInt())

	// ensure claim is disabled and the module account is empty
	params = suite.app.ClaimsKeeper.GetParams(ctx)
	suite.Require().False(params.EnableClaims)
	coins = suite.app.ClaimsKeeper.GetModuleAccountBalances(ctx)
	suite.Require().Equal(coins.AmountOf(params.GetClaimsDenom()), sdk.ZeroInt())

	// ensure community pool has the unclaimed escrow amount
	bal = suite.app.BankKeeper.GetBalance(ctx, distrModuleAddr, params.GetClaimsDenom())
	suite.Require().Equal(bal.Amount, escrow.Sub(claimedCoins))

	// make sure the claim records is empty
	suite.Require().Empty(suite.app.ClaimsKeeper.GetClaimsRecords(ctx))
}

func (suite *KeeperTestSuite) TestClawbackEmptyAccountsAirdrop() {
	suite.SetupTestWithEscrow()

	params := suite.app.ClaimsKeeper.GetParams(suite.ctx)
	tests := []struct {
		name           string
		address        string
		sequence       uint64
		expectClawback bool
		claimsRecord   types.ClaimsRecord
	}{
		{
			name:           "address active",
			address:        "evmos1qql8ag4cluz6r4dz28p3w00dnc9w8ueuafmxps",
			sequence:       1,
			expectClawback: false,
			claimsRecord: types.ClaimsRecord{
				InitialClaimableAmount: sdk.NewInt(100),
				ActionsCompleted:       []bool{true, false, true, false},
			},
		},
		{
			name:           "address inactive",
			address:        "evmos1x2w87cvt5mqjncav4lxy8yfreynn273xn5335v",
			sequence:       0,
			expectClawback: true,
			claimsRecord: types.ClaimsRecord{
				InitialClaimableAmount: sdk.NewInt(100),
				ActionsCompleted:       []bool{false, false, false, false},
			},
		},
	}

	for _, tc := range tests {
		addr, err := sdk.AccAddressFromBech32(tc.address)
		suite.Require().NoError(err, tc.name)

		acc := &ethermint.EthAccount{
			BaseAccount: authtypes.NewBaseAccount(sdk.AccAddress(addr.Bytes()), nil, 0, 0),
			CodeHash:    common.BytesToHash(crypto.Keccak256(nil)).String(),
		}

		err = acc.SetSequence(tc.sequence)
		suite.Require().NoError(err, tc.name)
		suite.app.AccountKeeper.SetAccount(suite.ctx, acc)
		suite.app.ClaimsKeeper.SetClaimsRecord(suite.ctx, addr, tc.claimsRecord)
		coins := sdk.NewCoins(sdk.NewInt64Coin(params.GetClaimsDenom(), 100))

		err = testutil.FundAccount(suite.app.BankKeeper, suite.ctx, addr, coins)
		suite.Require().NoError(err, tc.name)
	}

	err := suite.app.ClaimsKeeper.EndAirdrop(suite.ctx, params)
	suite.Require().NoError(err, "err: %s", err)

	for _, tc := range tests {
		addr, err := sdk.AccAddressFromBech32(tc.address)
		suite.Require().NoError(err, "err: %s test: %s", err, tc.name)

		coins := suite.app.BankKeeper.GetAllBalances(suite.ctx, addr)

		if tc.expectClawback {
			suite.Require().Equal(coins.AmountOfNoDenomValidation(params.GetClaimsDenom()), sdk.ZeroInt(),
				"balance incorrect. test: %s", tc.name)
		} else {
			suite.Require().Equal(coins.AmountOfNoDenomValidation(params.GetClaimsDenom()), sdk.NewInt(100),
				"balance incorrect. test: %s", tc.name)
		}
	}
}

// GetUserTotalClaimable returns claimable amount for a specific action done by
// an address at a given block time
func (suite *KeeperTestSuite) getUserTotalClaimable(ctx sdk.Context, addr sdk.AccAddress) sdk.Int {
	totalClaimable := sdk.ZeroInt()

	claimsRecord, found := suite.app.ClaimsKeeper.GetClaimsRecord(ctx, addr)
	if !found {
		return sdk.ZeroInt()
	}

	params := suite.app.ClaimsKeeper.GetParams(ctx)

	actions := []types.Action{types.ActionVote, types.ActionDelegate, types.ActionEVM, types.ActionIBCTransfer}
	for _, action := range actions {
		claimableForAction, _ := suite.app.ClaimsKeeper.GetClaimableAmountForAction(ctx, claimsRecord, action, params)
		totalClaimable = totalClaimable.Add(claimableForAction)
	}

	return totalClaimable
}
