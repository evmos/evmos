package keeper_test

import (
	"time"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/tharsis/ethermint/tests"
	"github.com/tharsis/evmos/x/claim/types"
)

func (suite *KeeperTestSuite) SetupClaimTest() {
	suite.SetupTest()
	params := suite.app.ClaimKeeper.GetParams(suite.ctx)

	coins := sdk.NewCoins(sdk.NewCoin(params.GetClaimDenom(), sdk.NewInt(10000000)))

	err := suite.app.BankKeeper.MintCoins(suite.ctx, minttypes.ModuleName, coins)
	suite.Require().NoError(err)
	err = suite.app.BankKeeper.SendCoinsFromModuleToModule(suite.ctx, minttypes.ModuleName, types.ModuleName, coins)
	suite.Require().NoError(err)
}

func (suite *KeeperTestSuite) TestGetClaimableAmountForAction() {
	addr := sdk.AccAddress(tests.GenerateAddress().Bytes())

	testCases := []struct {
		name        string
		claimRecord types.ClaimRecord
		params      types.Params
		expAmt      sdk.Int
	}{
		{
			"zero initial claimable amount",
			types.ClaimRecord{InitialClaimableAmount: sdk.ZeroInt()},
			types.Params{},
			sdk.ZeroInt(),
		},
	}

	for _, tc := range testCases {
		suite.SetupClaimTest()
		action := types.ActionDelegate
		amt := suite.app.ClaimKeeper.GetClaimableAmountForAction(suite.ctx, addr, tc.claimRecord, action, tc.params)
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
				suite.app.ClaimKeeper.SetClaimRecord(suite.ctx, addr, types.ClaimRecord{})
			},
			sdk.ZeroInt(),
		},
	}

	for _, tc := range testCases {
		suite.SetupClaimTest()
		tc.malleate()

		amt := suite.app.ClaimKeeper.GetUserTotalClaimable(suite.ctx, addr)
		suite.Require().Equal(tc.expAmt.Int64(), amt.Int64())
	}
}

func (suite *KeeperTestSuite) TestHookOfUnclaimableAccount() {
	suite.SetupClaimTest()

	pub1 := secp256k1.GenPrivKey().PubKey()
	addr1 := sdk.AccAddress(pub1.Address())
	suite.app.AccountKeeper.SetAccount(suite.ctx, authtypes.NewBaseAccount(addr1, nil, 0, 0))

	claim, found := suite.app.ClaimKeeper.GetClaimRecord(suite.ctx, addr1)
	suite.Require().False(found)
	suite.Require().Equal(types.ClaimRecord{}, claim)

	_, err := suite.app.ClaimKeeper.ClaimCoinsForAction(suite.ctx, addr1, types.ActionEVM)
	suite.Require().NoError(err)

	balances := suite.app.BankKeeper.GetAllBalances(suite.ctx, addr1)
	suite.Equal(sdk.Coins{}, balances)
}

func (suite *KeeperTestSuite) TestHookBeforeAirdropStart() {
	suite.SetupClaimTest()

	airdropStartTime := time.Now().Add(time.Hour)
	params := suite.app.ClaimKeeper.GetParams(suite.ctx)
	params.AirdropStartTime = airdropStartTime

	suite.app.ClaimKeeper.SetParams(suite.ctx, params)

	pub1 := secp256k1.GenPrivKey().PubKey()
	addr1 := sdk.AccAddress(pub1.Address())

	claimRecord := types.ClaimRecord{
		InitialClaimableAmount: sdk.NewInt(1000),
		ActionsCompleted:       []bool{false, false, false, false},
	}
	suite.app.AccountKeeper.SetAccount(suite.ctx, authtypes.NewBaseAccount(addr1, nil, 0, 0))

	suite.app.ClaimKeeper.SetClaimRecord(suite.ctx, addr1, claimRecord)

	coins := suite.app.ClaimKeeper.GetUserTotalClaimable(suite.ctx, addr1)
	suite.Require().Equal(coins, sdk.NewInt(1000))

	coins = suite.app.ClaimKeeper.GetClaimableAmountForAction(suite.ctx, addr1, claimRecord, types.ActionVote, suite.app.ClaimKeeper.GetParams(suite.ctx))
	suite.Require().Equal(coins, sdk.NewInt(250)) // 1/4th of the claimable

	_, err := suite.app.ClaimKeeper.ClaimCoinsForAction(suite.ctx, addr1, types.ActionVote)
	suite.Require().NoError(err)

	balances := suite.app.BankKeeper.GetAllBalances(suite.ctx, addr1)

	// Now, it is before starting air drop, so claim module should not send the balances to the user
	suite.Require().True(balances.Empty())

	_, err = suite.app.ClaimKeeper.ClaimCoinsForAction(suite.ctx.WithBlockTime(airdropStartTime), addr1, types.ActionVote)
	suite.Require().NoError(err)

	balances = suite.app.BankKeeper.GetAllBalances(suite.ctx, addr1)
	// Now, it is the time for air drop, so claim module should send the balances to the user
	suite.Require().Equal(claimRecord.InitialClaimableAmount.Quo(sdk.NewInt(4)), balances.AmountOf(suite.app.ClaimKeeper.GetParams(suite.ctx).ClaimDenom))
}

func (suite *KeeperTestSuite) TestHookAfterAirdropEnd() {
	suite.SetupClaimTest()

	// airdrop recipient address
	pub1 := secp256k1.GenPrivKey().PubKey()
	addr1 := sdk.AccAddress(pub1.Address())

	claimRecord := types.ClaimRecord{
		InitialClaimableAmount: sdk.NewInt(1000),
		ActionsCompleted:       []bool{false, false, false, false},
	}

	suite.app.AccountKeeper.SetAccount(suite.ctx, authtypes.NewBaseAccount(addr1, nil, 0, 0))
	suite.app.ClaimKeeper.SetClaimRecord(suite.ctx, addr1, claimRecord)

	params := suite.app.ClaimKeeper.GetParams(suite.ctx)
	suite.ctx = suite.ctx.WithBlockTime(params.AirdropStartTime.Add(params.DurationUntilDecay).Add(params.DurationOfDecay))

	err := suite.app.ClaimKeeper.EndAirdrop(suite.ctx, params)
	suite.Require().NoError(err)

	_, err = suite.app.ClaimKeeper.ClaimCoinsForAction(suite.ctx, addr1, types.ActionDelegate)
	suite.Require().NoError(err)
}

func (suite *KeeperTestSuite) TestDuplicatedActionNotWithdrawRepeatedly() {
	suite.SetupClaimTest()

	pub1 := secp256k1.GenPrivKey().PubKey()
	addr1 := sdk.AccAddress(pub1.Address())

	params := suite.app.ClaimKeeper.GetParams(suite.ctx)

	claimRecord := types.ClaimRecord{
		InitialClaimableAmount: sdk.NewInt(1000),
		ActionsCompleted:       []bool{false, false, false, false},
	}

	suite.app.AccountKeeper.SetAccount(suite.ctx, authtypes.NewBaseAccount(addr1, nil, 0, 0))

	suite.app.ClaimKeeper.SetClaimRecord(suite.ctx, addr1, claimRecord)

	coins1 := suite.app.ClaimKeeper.GetUserTotalClaimable(suite.ctx, addr1)
	suite.Require().Equal(coins1, claimRecord.InitialClaimableAmount)

	_, err := suite.app.ClaimKeeper.ClaimCoinsForAction(suite.ctx, addr1, types.ActionEVM)
	suite.Require().NoError(err)

	claim, found := suite.app.ClaimKeeper.GetClaimRecord(suite.ctx, addr1)
	suite.Require().True(found)
	suite.Require().True(claim.ActionsCompleted[types.ActionEVM-1])
	claimedCoins := suite.app.BankKeeper.GetAllBalances(suite.ctx, addr1)
	suite.Require().Equal(claimedCoins.AmountOf(params.GetClaimDenom()), claimRecord.InitialClaimableAmount.Quo(sdk.NewInt(4)))

	_, err = suite.app.ClaimKeeper.ClaimCoinsForAction(suite.ctx, addr1, types.ActionEVM)

	suite.NoError(err)
	suite.True(claim.ActionsCompleted[types.ActionEVM-1])
	claimedCoins = suite.app.BankKeeper.GetAllBalances(suite.ctx, addr1)
	suite.Require().Equal(claimedCoins.AmountOf(params.GetClaimDenom()), claimRecord.InitialClaimableAmount.Quo(sdk.NewInt(4)))
}

func (suite *KeeperTestSuite) TestDelegationAutoWithdrawAndDelegateMore() {
	suite.SetupClaimTest()

	pub1 := secp256k1.GenPrivKey().PubKey()
	pub2 := secp256k1.GenPrivKey().PubKey()
	addrs := []sdk.AccAddress{sdk.AccAddress(pub1.Address()), sdk.AccAddress(pub2.Address())}
	params := suite.app.ClaimKeeper.GetParams(suite.ctx)

	claimRecords := []types.ClaimRecord{
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
		suite.app.ClaimKeeper.SetClaimRecord(suite.ctx, addrs[i], claimRecords[i])
	}

	// test claim records set
	for i := 0; i < len(addrs); i++ {
		coins := suite.app.ClaimKeeper.GetUserTotalClaimable(suite.ctx, addrs[i])
		suite.Require().Equal(coins, claimRecords[i].InitialClaimableAmount)
	}

	// set addr[0] as a validator
	validator, err := stakingtypes.NewValidator(sdk.ValAddress(addrs[0]), pub1, stakingtypes.Description{})
	suite.Require().NoError(err)
	validator = stakingkeeper.TestingUpdateValidator(suite.app.StakingKeeper, suite.ctx, validator, true)
	suite.app.StakingKeeper.AfterValidatorCreated(suite.ctx, validator.GetOperator())

	validator, _ = validator.AddTokensFromDel(sdk.TokensFromConsensusPower(1, sdk.DefaultPowerReduction))
	delAmount := sdk.TokensFromConsensusPower(1, sdk.DefaultPowerReduction)
	err = simapp.FundAccount(suite.app.BankKeeper, suite.ctx, addrs[1], sdk.NewCoins(sdk.NewCoin(params.GetClaimDenom(), delAmount)))
	suite.Require().NoError(err)

	_, err = suite.app.StakingKeeper.Delegate(suite.ctx, addrs[1], delAmount, stakingtypes.Unbonded, validator, true)
	suite.Require().NoError(err)

	// delegation should automatically call claim and withdraw balance
	actualClaimedCoins := suite.app.BankKeeper.GetAllBalances(suite.ctx, addrs[1])
	actualClaimedCoin := actualClaimedCoins.AmountOf(params.GetClaimDenom())
	expectedClaimedCoin := claimRecords[1].InitialClaimableAmount.Quo(sdk.NewInt(int64(len(claimRecords[1].ActionsCompleted))))
	suite.Require().Equal(expectedClaimedCoin.String(), actualClaimedCoin.String())

	_, err = suite.app.StakingKeeper.Delegate(suite.ctx, addrs[1], actualClaimedCoin, stakingtypes.Unbonded, validator, true)
	suite.NoError(err)
}

func (suite *KeeperTestSuite) TestAirdropFlow() {
	suite.SetupClaimTest()

	addrs := []sdk.AccAddress{
		sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address()),
		sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address()),
	}

	claimRecords := []types.ClaimRecord{
		{
			InitialClaimableAmount: sdk.NewInt(100),
			ActionsCompleted:       []bool{false, false, false, false},
		},
		{
			InitialClaimableAmount: sdk.NewInt(200),
			ActionsCompleted:       []bool{false, false, false, false},
		},
	}

	params := suite.app.ClaimKeeper.GetParams(suite.ctx)

	// initialize accts
	for i := 0; i < len(addrs); i++ {
		suite.app.AccountKeeper.SetAccount(suite.ctx, authtypes.NewBaseAccount(addrs[i], nil, 0, 0))
		suite.app.ClaimKeeper.SetClaimRecord(suite.ctx, addrs[i], claimRecords[i])
	}

	coins1 := suite.app.ClaimKeeper.GetUserTotalClaimable(suite.ctx, addrs[0])
	suite.Require().Equal(coins1, claimRecords[0].InitialClaimableAmount)

	coins2 := suite.app.ClaimKeeper.GetUserTotalClaimable(suite.ctx, addrs[1])
	suite.Require().Equal(coins2, claimRecords[1].InitialClaimableAmount)

	coins3 := suite.app.ClaimKeeper.GetUserTotalClaimable(suite.ctx, sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address()))
	suite.Require().True(coins3.IsZero())

	// get rewards amount per action
	coins4 := suite.app.ClaimKeeper.GetClaimableAmountForAction(suite.ctx, addrs[0], claimRecords[0], types.ActionDelegate, suite.app.ClaimKeeper.GetParams(suite.ctx))
	suite.Require().Equal(sdk.NewCoins(sdk.NewInt64Coin(params.GetClaimDenom(), 25)).AmountOf(params.GetClaimDenom()), coins4) // 2 = 10.Quo(4)

	// get completed activities
	claimRecord, found := suite.app.ClaimKeeper.GetClaimRecord(suite.ctx, addrs[0])
	suite.Require().True(found)

	for i := 0; i < len(claimRecord.ActionsCompleted); i++ {
		suite.Require().False(claimRecord.ActionsCompleted[i])
	}

	// do half of actions
	_, err := suite.app.ClaimKeeper.ClaimCoinsForAction(suite.ctx, addrs[0], types.ActionEVM)
	suite.Require().NoError(err)
	_, err = suite.app.ClaimKeeper.ClaimCoinsForAction(suite.ctx, addrs[0], types.ActionDelegate)
	suite.Require().NoError(err)

	// check that half are completed
	claimRecord, found = suite.app.ClaimKeeper.GetClaimRecord(suite.ctx, addrs[0])
	suite.Require().True(found)

	suite.Require().True(claimRecord.HasClaimedAction(types.ActionEVM)) // We have Unspecified action in 0
	suite.Require().True(claimRecord.HasClaimedAction(types.ActionDelegate))

	// get balance after 2 actions done
	bal1 := suite.app.BankKeeper.GetAllBalances(suite.ctx, addrs[0])
	suite.Require().Equal(bal1.String(), sdk.NewCoins(sdk.NewInt64Coin(params.GetClaimDenom(), 50)).String())

	// check that claimable for completed activity is 0
	claimRecord1, _ := suite.app.ClaimKeeper.GetClaimRecord(suite.ctx, addrs[0])
	bal4 := suite.app.ClaimKeeper.GetClaimableAmountForAction(suite.ctx, addrs[0], claimRecord1, types.ActionEVM, params)
	suite.Require().Equal(bal4, sdk.NewInt(0))

	// do rest of actions
	_, err = suite.app.ClaimKeeper.ClaimCoinsForAction(suite.ctx, addrs[0], types.ActionIBCTransfer)
	suite.Require().NoError(err)
	_, err = suite.app.ClaimKeeper.ClaimCoinsForAction(suite.ctx, addrs[0], types.ActionVote)
	suite.Require().NoError(err)

	// get balance after rest actions done
	bal1 = suite.app.BankKeeper.GetAllBalances(suite.ctx, addrs[0])
	suite.Require().Equal(coins1, sdk.NewInt(100))

	// get claimable after withdrawing all
	coins1 = suite.app.ClaimKeeper.GetUserTotalClaimable(suite.ctx, addrs[0])
	suite.Require().NoError(err)
	suite.Require().True(coins1.IsZero())

	err = suite.app.ClaimKeeper.EndAirdrop(suite.ctx, suite.app.ClaimKeeper.GetParams(suite.ctx))
	suite.Require().NoError(err)

	moduleAccAddr := suite.app.AccountKeeper.GetModuleAddress(types.ModuleName)
	coins := suite.app.BankKeeper.GetBalance(suite.ctx, moduleAccAddr, params.GetClaimDenom())
	suite.Require().Equal(coins, sdk.NewInt64Coin(params.GetClaimDenom(), 0))

	coins2 = suite.app.ClaimKeeper.GetUserTotalClaimable(suite.ctx, addrs[1])
	suite.Require().NoError(err)
	suite.Require().Equal(coins2, sdk.NewInt(0))
}

func (suite *KeeperTestSuite) TestClaimOfDecayed() {
	suite.SetupClaimTest()

	airdropStartTime := time.Now()
	durationUntilDecay := time.Hour
	durationOfDecay := time.Hour * 4

	pub1 := secp256k1.GenPrivKey().PubKey()
	addr1 := sdk.AccAddress(pub1.Address())

	params := suite.app.ClaimKeeper.GetParams(suite.ctx)
	params.AirdropStartTime = airdropStartTime
	params.DurationUntilDecay = durationUntilDecay
	params.DurationOfDecay = durationOfDecay
	suite.app.ClaimKeeper.SetParams(suite.ctx, params)

	claimRecord := types.ClaimRecord{
		InitialClaimableAmount: sdk.NewInt(100),
		ActionsCompleted:       []bool{false, false, false, false},
	}

	t := []struct {
		fn func()
	}{
		{
			fn: func() {
				ctx := suite.ctx.WithBlockTime(airdropStartTime)
				coins := suite.app.ClaimKeeper.GetClaimableAmountForAction(ctx, addr1, claimRecord, types.ActionEVM, params)
				suite.Equal(claimRecord.InitialClaimableAmount.Quo(sdk.NewInt(4)).String(), coins.String())

				_, err := suite.app.ClaimKeeper.ClaimCoinsForAction(suite.ctx, addr1, types.ActionEVM)
				suite.Require().NoError(err)
				bal := suite.app.BankKeeper.GetAllBalances(ctx, addr1)
				suite.Equal(claimRecord.InitialClaimableAmount.Quo(sdk.NewInt(4)).String(), bal.AmountOf(params.GetClaimDenom()).String())
			},
		},
		{
			fn: func() {
				ctx := suite.ctx.WithBlockTime(airdropStartTime.Add(durationUntilDecay))
				coins := suite.app.ClaimKeeper.GetClaimableAmountForAction(ctx, addr1, claimRecord, types.ActionEVM, params)
				suite.Equal(claimRecord.InitialClaimableAmount.Quo(sdk.NewInt(4)).String(), coins.String())

				_, err := suite.app.ClaimKeeper.ClaimCoinsForAction(suite.ctx, addr1, types.ActionEVM)
				suite.Require().NoError(err)
				bal := suite.app.BankKeeper.GetAllBalances(ctx, addr1)
				suite.Equal(claimRecord.InitialClaimableAmount.Quo(sdk.NewInt(4)).String(), bal.AmountOf(params.GetClaimDenom()).String())
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
				coins := suite.app.ClaimKeeper.GetClaimableAmountForAction(ctx, addr1, claimRecord, types.ActionEVM, params)

				suite.Require().Equal(claimRecord.InitialClaimableAmount.ToDec().Mul(claimablePercent).Quo(sdk.NewDec(4)).RoundInt().String(), coins.String())

				_, err := suite.app.ClaimKeeper.ClaimCoinsForAction(ctx, addr1, types.ActionEVM)
				suite.Require().NoError(err)
				bal := suite.app.BankKeeper.GetAllBalances(ctx, addr1)

				suite.Require().Equal(claimRecord.InitialClaimableAmount.ToDec().Mul(claimablePercent).Quo(sdk.NewDec(4)).RoundInt().String(),
					bal.AmountOf(params.GetClaimDenom()).String())
			},
		},
		{
			fn: func() {
				ctx := suite.ctx.WithBlockTime(airdropStartTime.Add(durationUntilDecay).Add(durationOfDecay))
				_, err := suite.app.ClaimKeeper.ClaimCoinsForAction(ctx, addr1, types.ActionEVM)
				suite.Require().NoError(err)
				bal := suite.app.BankKeeper.GetAllBalances(ctx, addr1)
				suite.Require().True(bal.Empty())
			},
		},
	}

	for _, test := range t {
		suite.SetupClaimTest()

		suite.app.ClaimKeeper.SetParams(suite.ctx, types.Params{
			AirdropStartTime:   airdropStartTime,
			DurationUntilDecay: durationUntilDecay,
			DurationOfDecay:    durationOfDecay,
			EnableClaim:        true,
			ClaimDenom:         params.GetClaimDenom(),
		})

		suite.app.AccountKeeper.SetAccount(suite.ctx, authtypes.NewBaseAccount(addr1, nil, 0, 0))
		suite.app.ClaimKeeper.SetClaimRecord(suite.ctx, addr1, claimRecord)

		test.fn()
	}
}

func (suite *KeeperTestSuite) TestClawbackAirdrop() {
	suite.SetupClaimTest()

	claimRecord := types.ClaimRecord{
		InitialClaimableAmount: sdk.NewInt(100),
		ActionsCompleted:       []bool{false, false, false, false},
	}

	tests := []struct {
		name           string
		address        string
		sequence       uint64
		expectClawback bool
	}{
		{
			name:           "address active",
			address:        "evmos1qql8ag4cluz6r4dz28p3w00dnc9w8ueuafmxps",
			sequence:       1,
			expectClawback: false,
		},
		{
			name:           "address inactive",
			address:        "evmos1x2w87cvt5mqjncav4lxy8yfreynn273xn5335v",
			sequence:       0,
			expectClawback: true,
		},
		{
			name:           "non airdrop address active",
			address:        sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address()).String(),
			sequence:       1,
			expectClawback: false,
		},
	}

	for _, tc := range tests {
		addr, err := sdk.AccAddressFromBech32(tc.address)
		suite.Require().NoError(err, "err: %s test: %s", err, tc.name)
		acc := authtypes.NewBaseAccountWithAddress(addr)
		err = acc.SetSequence(tc.sequence)
		suite.Require().NoError(err, "err: %s test: %s", err, tc.name)
		suite.app.AccountKeeper.SetAccount(suite.ctx, acc)
		suite.app.ClaimKeeper.SetClaimRecord(suite.ctx, addr, claimRecord)
		coins := sdk.NewCoins(sdk.NewInt64Coin("aevmos", 100))
		simapp.FundAccount(suite.app.BankKeeper, suite.ctx, addr, coins)
	}

	err := suite.app.ClaimKeeper.EndAirdrop(suite.ctx, suite.app.ClaimKeeper.GetParams(suite.ctx))
	suite.Require().NoError(err, "err: %s", err)

	for _, tc := range tests {
		addr, err := sdk.AccAddressFromBech32(tc.address)
		suite.Require().NoError(err, "err: %s test: %s", err, tc.name)
		coins := suite.app.BankKeeper.GetAllBalances(suite.ctx, addr)
		if tc.expectClawback {
			suite.Require().True(coins.AmountOf("aevmos").Equal(sdk.NewInt(0)),
				"balance incorrect. test: %s", tc.name)
		} else {
			suite.Require().True(coins.AmountOf("aevmos").Equal(sdk.NewInt(100)),
				"balance incorrect. test: %s", tc.name)
		}
	}
}
