package keeper_test

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/tharsis/ethermint/crypto/ethsecp256k1"
	"github.com/tharsis/ethermint/tests"
	ethermint "github.com/tharsis/ethermint/types"
	"github.com/tharsis/evmos/testutil"
	inflationtypes "github.com/tharsis/evmos/x/inflation/types"

	"github.com/tharsis/evmos/x/claims/types"
)

func (suite *KeeperTestSuite) SetupClaimTest() {
	suite.SetupTest()
	params := suite.app.ClaimsKeeper.GetParams(suite.ctx)

	coins := sdk.NewCoins(sdk.NewCoin(params.GetClaimsDenom(), sdk.NewInt(10000000)))

	err := suite.app.BankKeeper.MintCoins(suite.ctx, inflationtypes.ModuleName, coins)
	suite.Require().NoError(err)
	err = suite.app.BankKeeper.SendCoinsFromModuleToModule(suite.ctx, inflationtypes.ModuleName, types.ModuleName, coins)
	suite.Require().NoError(err)
}

func (suite *KeeperTestSuite) TestGetClaimableAmountForAction() {
	addr := sdk.AccAddress(tests.GenerateAddress().Bytes())

	testCases := []struct {
		name         string
		claimsRecord types.ClaimsRecord
		params       types.Params
		expAmt       sdk.Int
	}{
		{
			"zero initial claimable amount",
			types.ClaimsRecord{InitialClaimableAmount: sdk.ZeroInt()},
			types.Params{},
			sdk.ZeroInt(),
		},
	}

	for _, tc := range testCases {
		suite.SetupClaimTest()
		action := types.ActionDelegate
		amt := suite.app.ClaimsKeeper.GetClaimableAmountForAction(suite.ctx, addr, tc.claimsRecord, action, tc.params)
		suite.Require().Equal(tc.expAmt.Int64(), amt.Int64())
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
			"zero - actions",
			func() {
				suite.app.ClaimsKeeper.SetClaimsRecord(suite.ctx, addr, types.ClaimsRecord{})
			},
			sdk.ZeroInt(),
		},
	}

	for _, tc := range testCases {
		suite.SetupClaimTest()
		tc.malleate()

		amt := suite.app.ClaimsKeeper.GetUserTotalClaimable(suite.ctx, addr)
		suite.Require().Equal(tc.expAmt.Int64(), amt.Int64())
	}
}

func (suite *KeeperTestSuite) TestHookOfUnclaimableAccount() {
	suite.SetupClaimTest()
	addr1 := sdk.AccAddress(tests.GenerateAddress().Bytes())
	suite.app.AccountKeeper.SetAccount(suite.ctx, authtypes.NewBaseAccount(addr1, nil, 0, 0))

	claim, found := suite.app.ClaimsKeeper.GetClaimsRecord(suite.ctx, addr1)
	suite.Require().False(found)
	suite.Require().Equal(types.ClaimsRecord{}, claim)

	_, err := suite.app.ClaimsKeeper.ClaimCoinsForAction(suite.ctx, addr1, types.ActionEVM)
	suite.Require().NoError(err)

	balances := suite.app.BankKeeper.GetAllBalances(suite.ctx, addr1)
	suite.Equal(sdk.Coins{}, balances)
}

func (suite *KeeperTestSuite) TestHookBeforeAirdropStart() {
	suite.SetupClaimTest()

	airdropStartTime := time.Now().Add(time.Hour)
	params := suite.app.ClaimsKeeper.GetParams(suite.ctx)
	params.AirdropStartTime = airdropStartTime

	suite.app.ClaimsKeeper.SetParams(suite.ctx, params)
	addr1 := sdk.AccAddress(tests.GenerateAddress().Bytes())

	claimsRecord := types.ClaimsRecord{
		InitialClaimableAmount: sdk.NewInt(1000),
		ActionsCompleted:       []bool{false, false, false, false},
	}
	suite.app.AccountKeeper.SetAccount(suite.ctx, authtypes.NewBaseAccount(addr1, nil, 0, 0))

	suite.app.ClaimsKeeper.SetClaimsRecord(suite.ctx, addr1, claimsRecord)

	coins := suite.app.ClaimsKeeper.GetUserTotalClaimable(suite.ctx, addr1)
	suite.Require().Equal(coins, sdk.NewInt(1000))

	coins = suite.app.ClaimsKeeper.GetClaimableAmountForAction(suite.ctx, addr1, claimsRecord, types.ActionVote, suite.app.ClaimsKeeper.GetParams(suite.ctx))
	suite.Require().Equal(coins, sdk.NewInt(250)) // 1/4th of the claimable

	_, err := suite.app.ClaimsKeeper.ClaimCoinsForAction(suite.ctx, addr1, types.ActionVote)
	suite.Require().NoError(err)

	balances := suite.app.BankKeeper.GetAllBalances(suite.ctx, addr1)

	// Now, it is before starting air drop, so claim module should not send the balances to the user
	suite.Require().True(balances.Empty())

	_, err = suite.app.ClaimsKeeper.ClaimCoinsForAction(suite.ctx.WithBlockTime(airdropStartTime), addr1, types.ActionVote)
	suite.Require().NoError(err)

	balances = suite.app.BankKeeper.GetAllBalances(suite.ctx, addr1)
	// Now, it is the time for air drop, so claim module should send the balances to the user
	suite.Require().Equal(claimsRecord.InitialClaimableAmount.Quo(sdk.NewInt(4)), balances.AmountOf(suite.app.ClaimsKeeper.GetParams(suite.ctx).ClaimsDenom))
}

func (suite *KeeperTestSuite) TestHookAfterAirdropEnd() {
	suite.SetupClaimTest()

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

	_, err = suite.app.ClaimsKeeper.ClaimCoinsForAction(suite.ctx, addr1, types.ActionDelegate)
	suite.Require().NoError(err)
}

func (suite *KeeperTestSuite) TestDuplicatedActionNotWithdrawRepeatedly() {
	suite.SetupClaimTest()
	addr1 := sdk.AccAddress(tests.GenerateAddress().Bytes())

	params := suite.app.ClaimsKeeper.GetParams(suite.ctx)

	claimsRecord := types.ClaimsRecord{
		InitialClaimableAmount: sdk.NewInt(1000),
		ActionsCompleted:       []bool{false, false, false, false},
	}

	suite.app.AccountKeeper.SetAccount(suite.ctx, authtypes.NewBaseAccount(addr1, nil, 0, 0))

	suite.app.ClaimsKeeper.SetClaimsRecord(suite.ctx, addr1, claimsRecord)

	coins1 := suite.app.ClaimsKeeper.GetUserTotalClaimable(suite.ctx, addr1)
	suite.Require().Equal(coins1, claimsRecord.InitialClaimableAmount)

	_, err := suite.app.ClaimsKeeper.ClaimCoinsForAction(suite.ctx, addr1, types.ActionEVM)
	suite.Require().NoError(err)

	claim, found := suite.app.ClaimsKeeper.GetClaimsRecord(suite.ctx, addr1)
	suite.Require().True(found)
	suite.Require().True(claim.ActionsCompleted[types.ActionEVM-1])
	claimedCoins := suite.app.BankKeeper.GetAllBalances(suite.ctx, addr1)
	suite.Require().Equal(claimedCoins.AmountOf(params.GetClaimsDenom()), claimsRecord.InitialClaimableAmount.Quo(sdk.NewInt(4)))

	_, err = suite.app.ClaimsKeeper.ClaimCoinsForAction(suite.ctx, addr1, types.ActionEVM)

	suite.NoError(err)
	suite.True(claim.ActionsCompleted[types.ActionEVM-1])
	claimedCoins = suite.app.BankKeeper.GetAllBalances(suite.ctx, addr1)
	suite.Require().Equal(claimedCoins.AmountOf(params.GetClaimsDenom()), claimsRecord.InitialClaimableAmount.Quo(sdk.NewInt(4)))
}

func (suite *KeeperTestSuite) TestDelegationAutoWithdrawAndDelegateMore() {
	suite.SetupClaimTest()

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
		coins := suite.app.ClaimsKeeper.GetUserTotalClaimable(suite.ctx, addrs[i])
		suite.Require().Equal(coins, claimsRecords[i].InitialClaimableAmount)
	}

	// set addr[0] as a validator
	validator, err := stakingtypes.NewValidator(sdk.ValAddress(addrs[0]), pub1.PubKey(), stakingtypes.Description{})
	suite.Require().NoError(err)
	validator = stakingkeeper.TestingUpdateValidator(suite.app.StakingKeeper, suite.ctx, validator, true)
	suite.app.StakingKeeper.AfterValidatorCreated(suite.ctx, validator.GetOperator())

	validator, _ = validator.AddTokensFromDel(sdk.TokensFromConsensusPower(1, sdk.DefaultPowerReduction))
	delAmount := sdk.TokensFromConsensusPower(1, sdk.DefaultPowerReduction)
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
	suite.SetupClaimTest()

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

	coins1 := suite.app.ClaimsKeeper.GetUserTotalClaimable(suite.ctx, addrs[0])
	suite.Require().Equal(coins1, claimsRecords[0].InitialClaimableAmount)

	coins2 := suite.app.ClaimsKeeper.GetUserTotalClaimable(suite.ctx, addrs[1])
	suite.Require().Equal(coins2, claimsRecords[1].InitialClaimableAmount)

	coins3 := suite.app.ClaimsKeeper.GetUserTotalClaimable(suite.ctx, sdk.AccAddress(tests.GenerateAddress().Bytes()))
	suite.Require().True(coins3.IsZero())

	// get rewards amount per action
	coins4 := suite.app.ClaimsKeeper.GetClaimableAmountForAction(suite.ctx, addrs[0], claimsRecords[0], types.ActionDelegate, suite.app.ClaimsKeeper.GetParams(suite.ctx))
	suite.Require().Equal(sdk.NewCoins(sdk.NewInt64Coin(params.GetClaimsDenom(), 25)).AmountOf(params.GetClaimsDenom()), coins4) // 2 = 10.Quo(4)

	// get completed activities
	claimsRecord, found := suite.app.ClaimsKeeper.GetClaimsRecord(suite.ctx, addrs[0])
	suite.Require().True(found)

	for i := 0; i < len(claimsRecord.ActionsCompleted); i++ {
		suite.Require().False(claimsRecord.ActionsCompleted[i])
	}

	// do half of actions
	_, err := suite.app.ClaimsKeeper.ClaimCoinsForAction(suite.ctx, addrs[0], types.ActionEVM)
	suite.Require().NoError(err)
	_, err = suite.app.ClaimsKeeper.ClaimCoinsForAction(suite.ctx, addrs[0], types.ActionDelegate)
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
	bal4 := suite.app.ClaimsKeeper.GetClaimableAmountForAction(suite.ctx, addrs[0], claimsRecord1, types.ActionEVM, params)
	suite.Require().Equal(bal4, sdk.NewInt(0))

	// do rest of actions
	_, err = suite.app.ClaimsKeeper.ClaimCoinsForAction(suite.ctx, addrs[0], types.ActionIBCTransfer)
	suite.Require().NoError(err)
	_, err = suite.app.ClaimsKeeper.ClaimCoinsForAction(suite.ctx, addrs[0], types.ActionVote)
	suite.Require().NoError(err)

	// get balance after rest actions done
	bal1 = suite.app.BankKeeper.GetAllBalances(suite.ctx, addrs[0])
	suite.Require().Equal(bal1.AmountOf(params.GetClaimsDenom()), sdk.NewInt(100))

	// get claimable after withdrawing all
	coins1 = suite.app.ClaimsKeeper.GetUserTotalClaimable(suite.ctx, addrs[0])
	suite.Require().NoError(err)
	suite.Require().True(coins1.IsZero())

	err = suite.app.ClaimsKeeper.EndAirdrop(suite.ctx, suite.app.ClaimsKeeper.GetParams(suite.ctx))
	suite.Require().NoError(err)

	moduleAccAddr := suite.app.AccountKeeper.GetModuleAddress(types.ModuleName)
	coins := suite.app.BankKeeper.GetBalance(suite.ctx, moduleAccAddr, params.GetClaimsDenom())
	suite.Require().Equal(coins, sdk.NewInt64Coin(params.GetClaimsDenom(), 0))

	coins2 = suite.app.ClaimsKeeper.GetUserTotalClaimable(suite.ctx, addrs[1])
	suite.Require().NoError(err)
	suite.Require().Equal(coins2, sdk.NewInt(0))
}

func (suite *KeeperTestSuite) TestClaimOfDecayed() {
	suite.SetupClaimTest()

	airdropStartTime := time.Now()
	durationUntilDecay := time.Hour
	durationOfDecay := time.Hour * 4

	addr1 := sdk.AccAddress(tests.GenerateAddress().Bytes())

	params := suite.app.ClaimsKeeper.GetParams(suite.ctx)
	params.AirdropStartTime = airdropStartTime
	params.DurationUntilDecay = durationUntilDecay
	params.DurationOfDecay = durationOfDecay
	suite.app.ClaimsKeeper.SetParams(suite.ctx, params)

	claimsRecord := types.ClaimsRecord{
		InitialClaimableAmount: sdk.NewInt(100),
		ActionsCompleted:       []bool{false, false, false, false},
	}

	t := []struct {
		fn func()
	}{
		{
			fn: func() {
				ctx := suite.ctx.WithBlockTime(airdropStartTime)
				coins := suite.app.ClaimsKeeper.GetClaimableAmountForAction(ctx, addr1, claimsRecord, types.ActionEVM, params)
				suite.Equal(claimsRecord.InitialClaimableAmount.Quo(sdk.NewInt(4)).String(), coins.String())

				_, err := suite.app.ClaimsKeeper.ClaimCoinsForAction(suite.ctx, addr1, types.ActionEVM)
				suite.Require().NoError(err)
				bal := suite.app.BankKeeper.GetAllBalances(ctx, addr1)
				suite.Equal(claimsRecord.InitialClaimableAmount.Quo(sdk.NewInt(4)).String(), bal.AmountOf(params.GetClaimsDenom()).String())
			},
		},
		{
			fn: func() {
				ctx := suite.ctx.WithBlockTime(airdropStartTime.Add(durationUntilDecay))
				coins := suite.app.ClaimsKeeper.GetClaimableAmountForAction(ctx, addr1, claimsRecord, types.ActionEVM, params)
				suite.Equal(claimsRecord.InitialClaimableAmount.Quo(sdk.NewInt(4)).String(), coins.String())

				_, err := suite.app.ClaimsKeeper.ClaimCoinsForAction(suite.ctx, addr1, types.ActionEVM)
				suite.Require().NoError(err)
				bal := suite.app.BankKeeper.GetAllBalances(ctx, addr1)
				suite.Equal(claimsRecord.InitialClaimableAmount.Quo(sdk.NewInt(4)).String(), bal.AmountOf(params.GetClaimsDenom()).String())
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
				coins := suite.app.ClaimsKeeper.GetClaimableAmountForAction(ctx, addr1, claimsRecord, types.ActionEVM, params)

				suite.Require().Equal(claimsRecord.InitialClaimableAmount.ToDec().Mul(claimablePercent).Quo(sdk.NewDec(4)).RoundInt().String(), coins.String())

				_, err := suite.app.ClaimsKeeper.ClaimCoinsForAction(ctx, addr1, types.ActionEVM)
				suite.Require().NoError(err)
				bal := suite.app.BankKeeper.GetAllBalances(ctx, addr1)

				suite.Require().Equal(claimsRecord.InitialClaimableAmount.ToDec().Mul(claimablePercent).Quo(sdk.NewDec(4)).RoundInt().String(),
					bal.AmountOf(params.GetClaimsDenom()).String())
			},
		},
		{
			fn: func() {
				ctx := suite.ctx.WithBlockTime(airdropStartTime.Add(durationUntilDecay).Add(durationOfDecay))
				_, err := suite.app.ClaimsKeeper.ClaimCoinsForAction(ctx, addr1, types.ActionEVM)
				suite.Require().NoError(err)
				bal := suite.app.BankKeeper.GetAllBalances(ctx, addr1)
				suite.Require().True(bal.Empty())
			},
		},
	}

	for _, test := range t {
		suite.SetupClaimTest()

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
	suite.SetupClaimTest()
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
	suite.Require().Equal(bal.Amount, sdk.NewInt(0))

	// claim some amount from airdrop
	addr1 := sdk.AccAddress(tests.GenerateAddress().Bytes())
	initClaim := sdk.NewInt(1000)
	claimsRecord := types.ClaimsRecord{
		InitialClaimableAmount: initClaim,
		ActionsCompleted:       []bool{false, false, false, false},
	}
	suite.app.AccountKeeper.SetAccount(ctx, authtypes.NewBaseAccount(addr1, nil, 0, 1))
	suite.app.ClaimsKeeper.SetClaimsRecord(ctx, addr1, claimsRecord)
	claimedCoins, err := suite.app.ClaimsKeeper.ClaimCoinsForAction(ctx, addr1, types.ActionEVM)
	suite.Require().NoError(err)
	coins = suite.app.ClaimsKeeper.GetModuleAccountBalances(ctx)
	suite.Require().Equal(coins.AmountOf(params.GetClaimsDenom()), escrow.Sub(claimedCoins))

	// End the airdrop
	suite.app.ClaimsKeeper.EndAirdrop(ctx, params)

	// Make sure no one can claim after airdrop ends
	claimedCoinsAfter, _ := suite.app.ClaimsKeeper.ClaimCoinsForAction(ctx, addr1, types.ActionDelegate)
	suite.Require().Equal(claimedCoinsAfter, sdk.NewInt(0))

	// ensure claim is disabled and the module account is empty
	params = suite.app.ClaimsKeeper.GetParams(ctx)
	suite.Require().False(params.EnableClaims)
	coins = suite.app.ClaimsKeeper.GetModuleAccountBalances(ctx)
	suite.Require().Equal(coins.AmountOf(params.GetClaimsDenom()), sdk.NewInt(0))

	// ensure community pool has the unclaimed escrow amount
	bal = suite.app.BankKeeper.GetBalance(ctx, distrModuleAddr, params.GetClaimsDenom())
	suite.Require().Equal(bal.Amount, escrow.Sub(claimedCoins))

	// make sure the claim records is empty
	suite.Require().Empty(suite.app.ClaimsKeeper.GetClaimsRecords(ctx))
}

func (suite *KeeperTestSuite) TestClawbackEmptyAccountsAirdrop() {
	suite.SetupClaimTest()

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
		suite.Require().NoError(err, "err: %s test: %s", err, tc.name)
		suite.app.AccountKeeper.SetAccount(suite.ctx, acc)
		suite.app.ClaimsKeeper.SetClaimsRecord(suite.ctx, addr, tc.claimsRecord)
		coins := sdk.NewCoins(sdk.NewInt64Coin(params.GetClaimsDenom(), 100))
		testutil.FundAccount(suite.app.BankKeeper, suite.ctx, addr, coins)
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
