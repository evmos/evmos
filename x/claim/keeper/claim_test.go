package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/tharsis/ethermint/tests"
	"github.com/tharsis/evmos/x/claim/types"
)

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
		suite.SetupTest()
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
		suite.SetupTest()
		tc.malleate()

		amt := suite.app.ClaimKeeper.GetUserTotalClaimable(suite.ctx, addr)
		suite.Require().Equal(tc.expAmt.Int64(), amt.Int64())
	}
}

// func (suite *KeeperTestSuite) TestHookOfUnclaimableAccount() {
// 	suite.SetupTest()

// 	pub1 := secp256k1.GenPrivKey().PubKey()
// 	addr1 := sdk.AccAddress(pub1.Address())
// 	suite.app.AccountKeeper.SetAccount(suite.ctx, authtypes.NewBaseAccount(addr1, nil, 0, 0))

// 	claim, found := suite.app.ClaimKeeper.GetClaimRecord(suite.ctx, addr1)
// 	suite.Require().True(found)
// 	suite.Require().Equal(types.ClaimRecord{}, claim)

// 	suite.app.ClaimKeeper.AfterSwap(suite.ctx, addr1)

// 	balances := suite.app.BankKeeper.GetAllBalances(suite.ctx, addr1)
// 	suite.Equal(sdk.Coins{}, balances)
// }

// func (suite *KeeperTestSuite) TestHookBeforeAirdropStart() {
// 	suite.SetupTest()

// 	airdropStartTime := time.Now().Add(time.Hour)

// 	suite.app.ClaimKeeper.SetParams(suite.ctx, types.Params{
// 		AirdropStartTime:   airdropStartTime,
// 		DurationUntilDecay: time.Hour,
// 		DurationOfDecay:    time.Hour * 4,
// 	})

// 	pub1 := secp256k1.GenPrivKey().PubKey()
// 	addr1 := sdk.AccAddress(pub1.Address())

// 	claimRecords := []types.ClaimRecord{
// 		{
// 			InitialClaimableAmount: sdk.NewInt(1000),
// 			ActionsCompleted:       []bool{false, false, false, false},
// 		},
// 	}
// 	suite.app.AccountKeeper.SetAccount(suite.ctx, authtypes.NewBaseAccount(addr1, nil, 0, 0))

// 	suite.app.ClaimKeeper.SetClaimRecord(suite.ctx, addr1, claimRecords[0])

// 	coins, err := suite.app.ClaimKeeper.GetUserTotalClaimable(suite.ctx, addr1)
// 	suite.Require().NoError(err)
// 	// Now, it is before starting air drop, so this value should return the empty coins
// 	suite.Require().True(coins.IsZero())

// 	coins, err = suite.app.ClaimKeeper.GetClaimableAmountForAction(suite.ctx, addr1, types.ActionSwap)
// 	suite.Require().NoError(err)
// 	// Now, it is before starting air drop, so this value should return the empty coins
// 	suite.Require().True(coins.Empty())

// 	suite.app.ClaimKeeper.AfterSwap(suite.ctx, addr1)
// 	balances := suite.app.BankKeeper.GetAllBalances(suite.ctx, addr1)
// 	// Now, it is before starting air drop, so claim module should not send the balances to the user after swap.
// 	suite.Require().True(balances.Empty())

// 	suite.app.ClaimKeeper.AfterSwap(suite.ctx.WithBlockTime(airdropStartTime), addr1)
// 	balances = suite.app.BankKeeper.GetAllBalances(suite.ctx, addr1)
// 	// Now, it is the time for air drop, so claim module should send the balances to the user after swap.
// 	suite.Require().Equal(claimRecords[0].InitialClaimableAmount.AmountOf(sdk.DefaultBondDenom).Quo(sdk.NewInt(4)), balances.AmountOf(sdk.DefaultBondDenom))
// }

// func (suite *KeeperTestSuite) TestHookAfterAirdropEnd() {
// 	suite.SetupTest()

// 	// airdrop recipient address
// 	addr1, _ := sdk.AccAddressFromBech32("osmo122fypjdzwscz998aytrrnmvavtaaarjjt6223p")

// 	claimRecords := []types.ClaimRecord{
// 		{
// 			Address:                addr1.String(),
// 			InitialClaimableAmount: sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 1000)),
// 			ActionCompleted:        []bool{false, false, false, false},
// 		},
// 	}
// 	suite.app.AccountKeeper.SetAccount(suite.ctx, authtypes.NewBaseAccount(addr1, nil, 0, 0))
// 	err := suite.app.ClaimKeeper.SetClaimRecords(suite.ctx, claimRecords)
// 	suite.Require().NoError(err)

// 	params, err := suite.app.ClaimKeeper.GetParams(suite.ctx)
// 	suite.Require().NoError(err)
// 	suite.ctx = suite.ctx.WithBlockTime(params.AirdropStartTime.Add(params.DurationUntilDecay).Add(params.DurationOfDecay))

// 	suite.app.ClaimKeeper.EndAirdrop(suite.ctx)

// 	suite.Require().NotPanics(func() {
// 		suite.app.ClaimKeeper.AfterSwap(suite.ctx, addr1)
// 	})
// }

// func (suite *KeeperTestSuite) TestDuplicatedActionNotWithdrawRepeatedly() {
// 	suite.SetupTest()

// 	pub1 := secp256k1.GenPrivKey().PubKey()
// 	addr1 := sdk.AccAddress(pub1.Address())

// 	claimRecords := []types.ClaimRecord{
// 		{
// 			Address:                addr1.String(),
// 			InitialClaimableAmount: sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 1000)),
// 			ActionCompleted:        []bool{false, false, false, false},
// 		},
// 	}
// 	suite.app.AccountKeeper.SetAccount(suite.ctx, authtypes.NewBaseAccount(addr1, nil, 0, 0))

// 	suite.app.ClaimKeeper.SetClaimRecords(suite.ctx, claimRecords)

// 	coins1, err := suite.app.ClaimKeeper.GetUserTotalClaimable(suite.ctx, addr1)
// 	suite.Require().NoError(err)
// 	suite.Require().Equal(coins1, claimRecords[0].InitialClaimableAmount)

// 	suite.app.ClaimKeeper.AfterSwap(suite.ctx, addr1)
// 	claim, found := suite.app.ClaimKeeper.GetClaimRecord(suite.ctx, addr1)
// 	suite.Require().True(found)
// 	suite.Require().True(claim.ActionCompleted[types.ActionSwap])
// 	claimedCoins := suite.app.BankKeeper.GetAllBalances(suite.ctx, addr1)
// 	suite.Require().Equal(claimedCoins.AmountOf(sdk.DefaultBondDenom), claimRecords[0].InitialClaimableAmount.AmountOf(sdk.DefaultBondDenom).Quo(sdk.NewInt(4)))

// 	suite.app.ClaimKeeper.AfterSwap(suite.ctx, addr1)
// 	claim, found = suite.app.ClaimKeeper.GetClaimRecord(suite.ctx, addr1)
// 	suite.NoError(err)
// 	suite.True(claim.ActionCompleted[types.ActionSwap])
// 	claimedCoins = suite.app.BankKeeper.GetAllBalances(suite.ctx, addr1)
// 	suite.Require().Equal(claimedCoins.AmountOf(sdk.DefaultBondDenom), claimRecords[0].InitialClaimableAmount.AmountOf(sdk.DefaultBondDenom).Quo(sdk.NewInt(4)))
// }

// func (suite *KeeperTestSuite) TestDelegationAutoWithdrawAndDelegateMore() {
// 	suite.SetupTest()

// 	pub1 := secp256k1.GenPrivKey().PubKey()
// 	pub2 := secp256k1.GenPrivKey().PubKey()
// 	addrs := []sdk.AccAddress{sdk.AccAddress(pub1.Address()), sdk.AccAddress(pub2.Address())}
// 	claimRecords := []types.ClaimRecord{
// 		{
// 			Address:                addrs[0].String(),
// 			InitialClaimableAmount: sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 1000)),
// 			ActionCompleted:        []bool{false, false, false, false},
// 		},
// 		{
// 			Address:                addrs[1].String(),
// 			InitialClaimableAmount: sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 1000)),
// 			ActionCompleted:        []bool{false, false, false, false},
// 		},
// 	}

// 	// initialize accts
// 	for i := 0; i < len(addrs); i++ {
// 		suite.app.AccountKeeper.SetAccount(suite.ctx, authtypes.NewBaseAccount(addrs[i], nil, 0, 0))
// 	}
// 	// initialize claim records
// 	err := suite.app.ClaimKeeper.SetClaimRecords(suite.ctx, claimRecords)
// 	suite.Require().NoError(err)

// 	// test claim records set
// 	for i := 0; i < len(addrs); i++ {
// 		coins, err := suite.app.ClaimKeeper.GetUserTotalClaimable(suite.ctx, addrs[i])
// 		suite.Require().NoError(err)
// 		suite.Require().Equal(coins, claimRecords[i].InitialClaimableAmount)
// 	}

// 	// set addr[0] as a validator
// 	validator, err := stakingtypes.NewValidator(sdk.ValAddress(addrs[0]), pub1, stakingtypes.Description{})
// 	suite.Require().NoError(err)
// 	validator = stakingkeeper.TestingUpdateValidator(suite.app.StakingKeeper, suite.ctx, validator, true)
// 	suite.app.StakingKeeper.AfterValidatorCreated(suite.ctx, validator.GetOperator())

// 	validator, _ = validator.AddTokensFromDel(sdk.TokensFromConsensusPower(1, sdk.DefaultPowerReduction))
// 	delAmount := sdk.TokensFromConsensusPower(1, sdk.DefaultPowerReduction)
// 	err = simapp.FundAccount(suite.app.BankKeeper, suite.ctx, addrs[1],
// 		sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, delAmount)))
// 	suite.Require().NoError(err)
// 	_, err = suite.app.StakingKeeper.Delegate(suite.ctx, addrs[1], delAmount, stakingtypes.Unbonded, validator, true)
// 	suite.Require().NoError(err)

// 	// delegation should automatically call claim and withdraw balance
// 	actualClaimedCoins := suite.app.BankKeeper.GetAllBalances(suite.ctx, addrs[1])
// 	actualClaimedCoin := actualClaimedCoins.AmountOf(sdk.DefaultBondDenom)
// 	expectedClaimedCoin := claimRecords[1].InitialClaimableAmount.AmountOf(sdk.DefaultBondDenom).Quo(sdk.NewInt(int64(len(claimRecords[1].ActionCompleted))))
// 	suite.Require().Equal(expectedClaimedCoin.String(), actualClaimedCoin.String())

// 	_, err = suite.app.StakingKeeper.Delegate(suite.ctx, addrs[1], actualClaimedCoin, stakingtypes.Unbonded, validator, true)
// 	suite.NoError(err)
// }

// func (suite *KeeperTestSuite) TestAirdropFlow() {
// 	suite.SetupTest()

// 	addr1 := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
// 	addr2 := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
// 	addr3 := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())

// 	claimRecords := []types.ClaimRecord{
// 		{
// 			Address:                addr1.String(),
// 			InitialClaimableAmount: sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 100)),
// 			ActionCompleted:        []bool{false, false, false, false},
// 		},
// 		{
// 			Address:                addr2.String(),
// 			InitialClaimableAmount: sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 200)),
// 			ActionCompleted:        []bool{false, false, false, false},
// 		},
// 	}

// 	err := suite.app.ClaimKeeper.SetClaimRecords(suite.ctx, claimRecords)
// 	suite.Require().NoError(err)

// 	coins1, err := suite.app.ClaimKeeper.GetUserTotalClaimable(suite.ctx, addr1)
// 	suite.Require().NoError(err)
// 	suite.Require().Equal(coins1, claimRecords[0].InitialClaimableAmount, coins1.String())

// 	coins2, err := suite.app.ClaimKeeper.GetUserTotalClaimable(suite.ctx, addr2)
// 	suite.Require().NoError(err)
// 	suite.Require().Equal(coins2, claimRecords[1].InitialClaimableAmount)

// 	coins3, err := suite.app.ClaimKeeper.GetUserTotalClaimable(suite.ctx, addr3)
// 	suite.Require().NoError(err)
// 	suite.Require().Equal(coins3, sdk.Coins{})

// 	// get rewards amount per action
// 	coins4, err := suite.app.ClaimKeeper.GetClaimableAmountForAction(suite.ctx, addr1, types.ActionAddLiquidity)
// 	suite.Require().NoError(err)
// 	suite.Require().Equal(coins4.String(), sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 25)).String()) // 2 = 10.Quo(4)

// 	// get completed activities
// 	claimRecord, err := suite.app.ClaimKeeper.GetClaimRecord(suite.ctx, addr1)
// 	suite.Require().NoError(err)
// 	for i := range types.Action_name {
// 		suite.Require().False(claimRecord.ActionCompleted[i])
// 	}

// 	// do half of actions
// 	suite.app.ClaimKeeper.AfterAddLiquidity(suite.ctx, addr1)
// 	suite.app.ClaimKeeper.AfterSwap(suite.ctx, addr1)

// 	// check that half are completed
// 	claimRecord, err = suite.app.ClaimKeeper.GetClaimRecord(suite.ctx, addr1)
// 	suite.Require().NoError(err)
// 	suite.Require().True(claimRecord.ActionCompleted[types.ActionAddLiquidity])
// 	suite.Require().True(claimRecord.ActionCompleted[types.ActionSwap])

// 	// get balance after 2 actions done
// 	coins1 = suite.app.BankKeeper.GetAllBalances(suite.ctx, addr1)
// 	suite.Require().Equal(coins1.String(), sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 50)).String())

// 	// check that claimable for completed activity is 0
// 	coins4, err = suite.app.ClaimKeeper.GetClaimableAmountForAction(suite.ctx, addr1, types.ActionAddLiquidity)
// 	suite.Require().NoError(err)
// 	suite.Require().Equal(coins4.String(), sdk.Coins{}.String()) // 2 = 10.Quo(4)

// 	// do rest of actions
// 	suite.app.ClaimKeeper.AfterProposalVote(suite.ctx, 1, addr1)
// 	suite.app.ClaimKeeper.AfterDelegationModified(suite.ctx, addr1, sdk.ValAddress(addr1))

// 	// get balance after rest actions done
// 	coins1 = suite.app.BankKeeper.GetAllBalances(suite.ctx, addr1)
// 	suite.Require().Equal(coins1.String(), sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 100)).String())

// 	// get claimable after withdrawing all
// 	coins1, err = suite.app.ClaimKeeper.GetUserTotalClaimable(suite.ctx, addr1)
// 	suite.Require().NoError(err)
// 	suite.Require().True(coins1.Empty())

// 	err = suite.app.ClaimKeeper.EndAirdrop(suite.ctx)
// 	suite.Require().NoError(err)

// 	moduleAccAddr := suite.app.AccountKeeper.GetModuleAddress(types.ModuleName)
// 	coins := suite.app.BankKeeper.GetBalance(suite.ctx, moduleAccAddr, sdk.DefaultBondDenom)
// 	suite.Require().Equal(coins, sdk.NewInt64Coin(sdk.DefaultBondDenom, 0))

// 	coins2, err = suite.app.ClaimKeeper.GetUserTotalClaimable(suite.ctx, addr2)
// 	suite.Require().NoError(err)
// 	suite.Require().Equal(coins2, sdk.Coins{})
// }

// func (suite *KeeperTestSuite) TestClaimOfDecayed() {
// 	airdropStartTime := time.Now()
// 	durationUntilDecay := time.Hour
// 	durationOfDecay := time.Hour * 4

// 	pub1 := secp256k1.GenPrivKey().PubKey()
// 	addr1 := sdk.AccAddress(pub1.Address())

// 	claimRecords := []types.ClaimRecord{
// 		{
// 			Address:                addr1.String(),
// 			InitialClaimableAmount: sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 1000)),
// 			ActionCompleted:        []bool{false, false, false, false},
// 		},
// 	}

// 	tests := []struct {
// 		fn func()
// 	}{
// 		{
// 			fn: func() {
// 				ctx := suite.ctx.WithBlockTime(airdropStartTime)
// 				coins, err := suite.app.ClaimKeeper.GetClaimableAmountForAction(ctx, addr1, types.ActionSwap)
// 				suite.NoError(err)
// 				suite.Equal(claimRecords[0].InitialClaimableAmount.AmountOf(sdk.DefaultBondDenom).Quo(sdk.NewInt(4)).String(), coins.AmountOf(sdk.DefaultBondDenom).String())

// 				suite.app.ClaimKeeper.AfterSwap(ctx, addr1)
// 				coins = suite.app.BankKeeper.GetAllBalances(ctx, addr1)
// 				suite.Equal(claimRecords[0].InitialClaimableAmount.AmountOf(sdk.DefaultBondDenom).Quo(sdk.NewInt(4)).String(), coins.AmountOf(sdk.DefaultBondDenom).String())
// 			},
// 		},
// 		{
// 			fn: func() {
// 				ctx := suite.ctx.WithBlockTime(airdropStartTime.Add(durationUntilDecay))
// 				coins, err := suite.app.ClaimKeeper.GetClaimableAmountForAction(ctx, addr1, types.ActionSwap)
// 				suite.NoError(err)
// 				suite.Equal(claimRecords[0].InitialClaimableAmount.AmountOf(sdk.DefaultBondDenom).Quo(sdk.NewInt(4)).String(), coins.AmountOf(sdk.DefaultBondDenom).String())

// 				suite.app.ClaimKeeper.AfterSwap(ctx, addr1)
// 				coins = suite.app.BankKeeper.GetAllBalances(ctx, addr1)
// 				suite.Equal(claimRecords[0].InitialClaimableAmount.AmountOf(sdk.DefaultBondDenom).Quo(sdk.NewInt(4)).String(), coins.AmountOf(sdk.DefaultBondDenom).String())
// 			},
// 		},
// 		{
// 			fn: func() {
// 				ctx := suite.ctx.WithBlockTime(airdropStartTime.Add(durationUntilDecay).Add(durationOfDecay / 2))
// 				coins, err := suite.app.ClaimKeeper.GetClaimableAmountForAction(ctx, addr1, types.ActionSwap)
// 				suite.Require().NoError(err)
// 				suite.Require().Equal(claimRecords[0].InitialClaimableAmount.AmountOf(sdk.DefaultBondDenom).Quo(sdk.NewInt(8)).String(), coins.AmountOf(sdk.DefaultBondDenom).String())

// 				suite.app.ClaimKeeper.AfterSwap(ctx, addr1)
// 				coins = suite.app.BankKeeper.GetAllBalances(ctx, addr1)
// 				suite.Require().Equal(claimRecords[0].InitialClaimableAmount.AmountOf(sdk.DefaultBondDenom).Quo(sdk.NewInt(8)).String(), coins.AmountOf(sdk.DefaultBondDenom).String())
// 			},
// 		},
// 		{
// 			fn: func() {
// 				ctx := suite.ctx.WithBlockTime(airdropStartTime.Add(durationUntilDecay).Add(durationOfDecay))
// 				coins, err := suite.app.ClaimKeeper.GetClaimableAmountForAction(ctx, addr1, types.ActionSwap)
// 				suite.Require().NoError(err)
// 				suite.Require().True(coins.Empty())

// 				suite.app.ClaimKeeper.AfterSwap(ctx, addr1)
// 				coins = suite.app.BankKeeper.GetAllBalances(ctx, addr1)
// 				suite.Require().True(coins.Empty())
// 			},
// 		},
// 	}

// 	for _, test := range tests {
// 		suite.SetupTest()

// 		suite.app.ClaimKeeper.SetParams(suite.ctx, types.Params{
// 			AirdropStartTime:   airdropStartTime,
// 			DurationUntilDecay: durationUntilDecay,
// 			DurationOfDecay:    durationOfDecay,
// 		})

// 		suite.app.AccountKeeper.SetAccount(suite.ctx, authtypes.NewBaseAccount(addr1, nil, 0, 0))
// 		suite.app.ClaimKeeper.SetClaimRecords(suite.ctx, claimRecords)

// 		test.fn()
// 	}
// }

// func (suite *KeeperTestSuite) TestClawbackAirdrop() {
// 	suite.SetupTest()

// 	tests := []struct {
// 		name           string
// 		address        string
// 		sequence       uint64
// 		expectClawback bool
// 	}{
// 		{
// 			name:           "airdrop address active",
// 			address:        "osmo122fypjdzwscz998aytrrnmvavtaaarjjt6223p",
// 			sequence:       1,
// 			expectClawback: false,
// 		},
// 		{
// 			name:           "airdrop address inactive",
// 			address:        "osmo122g3jv9que3zkxy25a2wt0wlgh68mudwptyvzw",
// 			sequence:       0,
// 			expectClawback: true,
// 		},
// 		{
// 			name:           "non airdrop address active",
// 			address:        sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address()).String(),
// 			sequence:       1,
// 			expectClawback: false,
// 		},
// 		{
// 			name:           "non airdrop address inactive",
// 			address:        sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address()).String(),
// 			sequence:       0,
// 			expectClawback: false,
// 		},
// 	}

// 	for _, tc := range tests {
// 		addr, err := sdk.AccAddressFromBech32(tc.address)
// 		suite.Require().NoError(err, "err: %s test: %s", err, tc.name)
// 		acc := authtypes.NewBaseAccountWithAddress(addr)
// 		err = acc.SetSequence(tc.sequence)
// 		suite.Require().NoError(err, "err: %s test: %s", err, tc.name)
// 		suite.app.AccountKeeper.SetAccount(suite.ctx, acc)
// 		coins := sdk.NewCoins(
// 			sdk.NewInt64Coin("uosmo", 100), sdk.NewInt64Coin("uion", 100))
// 		simapp.FundAccount(suite.app.BankKeeper, suite.ctx, addr, coins)
// 	}

// 	err := suite.app.ClaimKeeper.EndAirdrop(suite.ctx, suite.app.ClaimKeeper.GetParams(suite.ctx))
// 	suite.Require().NoError(err, "err: %s", err)

// 	for _, tc := range tests {
// 		addr, err := sdk.AccAddressFromBech32(tc.address)
// 		suite.Require().NoError(err, "err: %s test: %s", err, tc.name)
// 		coins := suite.app.BankKeeper.GetAllBalances(suite.ctx, addr)
// 		if tc.expectClawback {
// 			suite.Require().True(coins.IsEqual(sdk.NewCoins()),
// 				"balance incorrect. test: %s", tc.name)
// 		} else {
// 			suite.Require().True(coins.IsEqual(sdk.NewCoins(
// 				sdk.NewInt64Coin("uosmo", 100), sdk.NewInt64Coin("uion", 100),
// 			)), "balance incorrect. test: %s", tc.name)
// 		}
// 	}
// }
