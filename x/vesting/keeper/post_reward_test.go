package keeper_test

// import (
// 	"time"

// 	tmtime "github.com/tendermint/tendermint/types/time"

// 	sdk "github.com/cosmos/cosmos-sdk/types"
// 	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
// 	sdkvesting "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
// 	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

// 	"github.com/tharsis/evmos/testutil"
// 	"github.com/tharsis/evmos/x/vesting/types"
// )

// var (
// 	stakeDenom = "stake"
// 	feeDenom   = "fee"
// )

// func (suite *KeeperTestSuite) TestRewards() {
// 	c := sdk.NewCoins
// 	stake := func(x int64) sdk.Coin { return sdk.NewInt64Coin(stakeDenom, x) }
// 	now := tmtime.Now()
// 	origCoins := sdk.Coins{sdk.NewInt64Coin(feeDenom, 1000), sdk.NewInt64Coin(stakeDenom, 100)}

// 	// Create Validator
// 	val := suite.validator
// 	suite.Require().Equal("stake", suite.app.StakingKeeper.BondDenom(suite.ctx))

// 	// Create clawback vesting account
// 	lockupPeriods := sdkvesting.Periods{
// 		{Length: 1, Amount: c(stake(4000))},
// 	}
// 	vestingPeriods := sdkvesting.Periods{
// 		{Length: int64(100), Amount: c(stake(500))},
// 		{Length: int64(100), Amount: c(stake(500))},
// 		{Length: int64(100), Amount: c(stake(500))},
// 		{Length: int64(100), Amount: c(stake(500))},
// 		{Length: int64(100), Amount: c(stake(500))},
// 		{Length: int64(100), Amount: c(stake(500))},
// 		{Length: int64(100), Amount: c(stake(500))},
// 		{Length: int64(100), Amount: c(stake(500))},
// 	}
// 	vestingStart := s.ctx.BlockTime().Unix()
// 	baseAccount := authtypes.NewBaseAccountWithAddress(addr)
// 	funder := sdk.AccAddress(types.ModuleName)
// 	va := types.NewClawbackVestingAccount(baseAccount, funder, origCoins, vestingStart, lockupPeriods, vestingPeriods)
// 	testutil.FundAccount(s.app.BankKeeper, s.ctx, addr2, c(stake(3700)))
// 	s.app.AccountKeeper.SetAccount(s.ctx, va)

// 	// fund the vesting account with 300 stake lost to transfer
// 	err := testutil.FundAccount(suite.app.BankKeeper, suite.ctx, addr, c(stake(3700)))

// 	ctx := suite.ctx.WithBlockTime(now.Add(350 * time.Second))

// 	// delegate 1600
// 	shares, err := suite.app.StakingKeeper.Delegate(ctx, addr, sdk.NewInt(1600), stakingtypes.Unbonded, val, true)
// 	suite.Require().NoError(err)
// 	suite.Require().Equal(sdk.NewInt(1600), shares.TruncateInt())
// 	suite.Require().Equal(int64(2100), suite.app.BankKeeper.GetBalance(ctx, addr, stakeDenom).Amount.Int64())
// 	va = suite.app.AccountKeeper.GetAccount(ctx, addr).(*types.ClawbackVestingAccount)

// 	// distribute a reward of 120stake
// 	err = testutil.FundAccount(suite.app.BankKeeper, ctx, addr, c(stake(120)))
// 	suite.Require().NoError(err)
// 	suite.app.VestingKeeper.PostReward(ctx, *va, c(stake(120)))
// 	va = suite.app.AccountKeeper.GetAccount(ctx, addr).(*types.ClawbackVestingAccount)
// 	suite.Require().Equal(int64(4030), va.OriginalVesting.AmountOf(stakeDenom).Int64())
// 	suite.Require().Equal(8, len(va.VestingPeriods))
// 	for i := 0; i < 3; i++ {
// 		suite.Require().Equal(int64(500), va.VestingPeriods[i].Amount.AmountOf(stakeDenom).Int64())
// 	}
// 	for i := 3; i < 8; i++ {
// 		suite.Require().Equal(int64(506), va.VestingPeriods[i].Amount.AmountOf(stakeDenom).Int64())
// 	}
// }

// func (suite *KeeperTestSuite) TestRewards_PostSlash() {
// 	c := sdk.NewCoins
// 	stake := func(x int64) sdk.Coin { return sdk.NewInt64Coin(stakeDenom, x) }
// 	now := tmtime.Now()

// 	// set up simapp and validators
// 	app := simsuite.app.Setup(false)
// 	ctx := suite.app.BaseApp.NewContext(false, tmproto.Header{}).WithBlockTime((now))
// 	_, val := createValidator(ctx, app, 100)
// 	suite.Require().Equal("stake", suite.app.StakingKeeper.BondDenom(ctx))

// 	// create vesting account with a simulated 350stake lost to slashing
// 	lockupPeriods := sdkvesting.Periods{
// 		{Length: 1, Amount: c(stake(4000))},
// 	}
// 	vestingPeriods := sdkvesting.Periods{
// 		{Length: int64(100), Amount: c(stake(500))},
// 		{Length: int64(100), Amount: c(stake(500))},
// 		{Length: int64(100), Amount: c(stake(500))},
// 		{Length: int64(100), Amount: c(stake(500))},
// 		{Length: int64(100), Amount: c(stake(500))},
// 		{Length: int64(100), Amount: c(stake(500))},
// 		{Length: int64(100), Amount: c(stake(500))},
// 		{Length: int64(100), Amount: c(stake(500))},
// 	}
// 	bacc, _ := initBaseAccount()
// 	origCoins := c(stake(4000))
// 	_, _, funder := testdata.KeyTestPubAddr()
// 	va := types.NewClawbackVestingAccount(bacc, funder, origCoins, now.Unix(), lockupPeriods, vestingPeriods)
// 	addr := va.GetAddress()
// 	va.DelegatedVesting = c(stake(350))
// 	suite.app.AccountKeeper.SetAccount(ctx, va)

// 	// fund the vesting account with 350 stake lost to slashing
// 	err := simsuite.app.FundAccount(suite.app.BankKeeper, ctx, addr, c(stake(3650)))
// 	suite.Require().NoError(err)
// 	suite.Require().Equal(int64(3650), app.BankKeeper.GetBalance(ctx, addr, stakeDenom).Amount.Int64())

// 	// delegate all 3650stake
// 	shares, err := app.StakingKeeper.Delegate(ctx, addr, sdk.NewInt(3650), stakingtypes.Unbonded, val, true)
// 	suite.Require().NoError(err)
// 	suite.Require().Equal(sdk.NewInt(3650), shares.TruncateInt())
// 	suite.Require().Equal(int64(0), app.BankKeeper.GetBalance(ctx, addr, stakeDenom).Amount.Int64())
// 	va = app.AccountKeeper.GetAccount(ctx, addr).(*types.ClawbackVestingAccount)

// 	// distribute a reward of 160stake - should all be unvested
// 	err = simapp.FundAccount(app.BankKeeper, ctx, addr, c(stake(160)))
// 	suite.Require().NoError(err)
// 	va.PostReward(ctx, c(stake(160)), app.AccountKeeper, app.BankKeeper, app.StakingKeeper)
// 	va = app.AccountKeeper.GetAccount(ctx, addr).(*types.ClawbackVestingAccount)
// 	suite.Require().Equal(int64(4160), va.OriginalVesting.AmountOf(stakeDenom).Int64())
// 	suite.Require().Equal(8, len(va.VestingPeriods))
// 	for i := 0; i < 8; i++ {
// 		suite.Require().Equal(int64(520), va.VestingPeriods[i].Amount.AmountOf(stakeDenom).Int64())
// 	}

// 	// must not be able to transfer reward until it vests
// 	_, _, dest := testdata.KeyTestPubAddr()
// 	err = app.BankKeeper.SendCoins(ctx, addr, desc(stake(1)))
// 	suite.Require().Error(err)
// 	ctx = ctx.WithBlockTime(now.Add(600 * time.Second))
// 	err = app.BankKeeper.SendCoins(ctx, addr, desc(stake(160)))
// 	suite.Require().NoError(err)

// 	// distribute another reward once everything has vested
// 	ctx = ctx.WithBlockTime(now.Add(1200 * time.Second))
// 	err = simapp.FundAccount(app.BankKeeper, ctx, addr, c(stake(160)))
// 	suite.Require().NoError(err)
// 	va.PostReward(ctx, c(stake(160)), app.AccountKeeper, app.BankKeeper, app.StakingKeeper)
// 	va = app.AccountKeeper.GetAccount(ctx, addr).(*types.ClawbackVestingAccount)
// 	// shouldn't be added to vesting schedule
// 	suite.Require().Equal(int64(4160), va.OriginalVesting.AmountOf(stakeDenom).Int64())
// }

// func (suite *KeeperTestSuite) TestAddGrantClawbackVestingAcc_fullSlash() {
// 	c := sdk.NewCoins
// 	stake := func(amt int64) sdk.Coin { return sdk.NewInt64Coin(stakeDenom, amt) }
// 	now := tmtime.Now()

// 	// set up simapp
// 	app := simapp.Setup(false)
// 	ctx := app.BaseApp.NewContext(false, tmproto.Header{}).WithBlockTime((now))
// 	suite.Require().Equal("stake", app.StakingKeeper.BondDenom(ctx))

// 	// create an account with an initial grant
// 	_, _, funder := testdata.KeyTestPubAddr()
// 	lockupPeriods := sdkvesting.Periods{{Length: 1, Amount: c(stake(100))}}
// 	vestingPeriods := sdkvesting.Periods{
// 		{Length: 100, Amount: c(stake(40))},
// 		{Length: 100, Amount: c(stake(60))},
// 	}
// 	bacc, _ := initBaseAccount()
// 	origCoins := c(stake(100))
// 	va := types.NewClawbackVestingAccount(bacc, funder, origCoins, now.Unix(), lockupPeriods, vestingPeriods)
// 	addr := va.GetAddress()

// 	// simulate all 100stake lost to slashing
// 	va.DelegatedVesting = c(stake(100))

// 	// Nothing locked at now+150, due to slashing
// 	suite.Require().Equal(int64(0), va.LockedCoins(ctx.WithBlockTime(now.Add(150*time.Second))).AmountOf(stakeDenom).Int64())

// 	// Add a new grant of 50stake
// 	newGrant := c(stake(50))
// 	va.AddGrant(ctx, app.StakingKeeper, now.Add(500*time.Second).Unix(),
// 		[]sdkvesting.Period{{Length: 1, Amount: newGrant}},
// 		[]sdkvesting.Period{{Length: 50, Amount: newGrant}}, newGrant)
// 	app.AccountKeeper.SetAccount(ctx, va)

// 	// The new 50stake are locked at now+150
// 	suite.Require().Equal(int64(50), va.LockedCoins(ctx.WithBlockTime(now.Add(150*time.Second))).AmountOf(stakeDenom).Int64())

// 	// fund the vesting account with new grant (old has vested and transferred out)
// 	ctx = ctx.WithBlockTime(now.Add(500 * time.Second))
// 	err := simapp.FundAccount(app.BankKeeper, ctx, addr, newGrant)
// 	suite.Require().NoError(err)
// 	suite.Require().Equal(int64(50), app.BankKeeper.GetBalance(ctx, addr, stakeDenom).Amount.Int64())

// 	// we should not be able to transfer the new grant out until it has vested
// 	_, _, dest := testdata.KeyTestPubAddr()
// 	err = app.BankKeeper.SendCoins(ctx, addr, desc(stake(1)))
// 	suite.Require().Error(err)
// 	ctx = ctx.WithBlockTime(now.Add(600 * time.Second))
// 	err = app.BankKeeper.SendCoins(ctx, addr, desnewGrant)
// 	suite.Require().NoError(err)
// }

// func (suite *KeeperTestSuite) TestAddGrantClawbackVestingAcc() {
// 	c := sdk.NewCoins
// 	fee := func(amt int64) sdk.Coin { return sdk.NewInt64Coin(feeDenom, amt) }
// 	stake := func(amt int64) sdk.Coin { return sdk.NewInt64Coin(stakeDenom, amt) }
// 	now := tmtime.Now()

// 	// set up simapp
// 	app := simapp.Setup(false)
// 	ctx := app.BaseApp.NewContext(false, tmproto.Header{}).WithBlockTime((now))
// 	suite.Require().Equal("stake", app.StakingKeeper.BondDenom(ctx))

// 	// create an account with an initial grant
// 	_, _, funder := testdata.KeyTestPubAddr()
// 	lockupPeriods := sdkvesting.Periods{{Length: 1, Amount: c(fee(1000), stake(100))}}
// 	vestingPeriods := sdkvesting.Periods{
// 		{Length: 100, Amount: c(fee(650), stake(40))},
// 		{Length: 100, Amount: c(fee(350), stake(60))},
// 	}
// 	bacc, origCoins := initBaseAccount()
// 	va := types.NewClawbackVestingAccount(bacc, funder, origCoins, now.Unix(), lockupPeriods, vestingPeriods)
// 	addr := va.GetAddress()

// 	// simulate 54stake (unvested) lost to slashing
// 	va.DelegatedVesting = c(stake(54))

// 	// Only 6stake are locked at now+150, due to slashing
// 	suite.Require().Equal(int64(6), va.LockedCoins(ctx.WithBlockTime(now.Add(150*time.Second))).AmountOf(stakeDenom).Int64())

// 	// Add a new grant of 50stake
// 	newGrant := c(stake(50))
// 	va.AddGrant(ctx, app.StakingKeeper, now.Add(500*time.Second).Unix(),
// 		[]sdkvesting.Period{{Length: 1, Amount: newGrant}},
// 		[]sdkvesting.Period{{Length: 50, Amount: newGrant}}, newGrant)
// 	app.AccountKeeper.SetAccount(ctx, va)

// 	// Only 56stake locked at now+150, due to slashing
// 	suite.Require().Equal(int64(56), va.LockedCoins(ctx.WithBlockTime(now.Add(150*time.Second))).AmountOf(stakeDenom).Int64())

// 	// fund the vesting account with new grant (old has vested and transferred out)
// 	ctx = ctx.WithBlockTime(now.Add(500 * time.Second))
// 	err := simapp.FundAccount(app.BankKeeper, ctx, addr, newGrant)
// 	suite.Require().NoError(err)
// 	suite.Require().Equal(int64(50), app.BankKeeper.GetBalance(ctx, addr, stakeDenom).Amount.Int64())

// 	// we should not be able to transfer the new grant out until it has vested
// 	_, _, dest := testdata.KeyTestPubAddr()
// 	err = app.BankKeeper.SendCoins(ctx, addr, desc(stake(1)))
// 	suite.Require().Error(err)
// 	ctx = ctx.WithBlockTime(now.Add(600 * time.Second))
// 	err = app.BankKeeper.SendCoins(ctx, addr, desnewGrant)
// 	suite.Require().NoError(err)
// }

////////////////////////////////////////////////////////

// func TestClawback() {
// 	c := sdk.NewCoins
// 	fee := func(x int64) sdk.Coin { return sdk.NewInt64Coin(feeDenom, x) }
// 	stake := func(x int64) sdk.Coin { return sdk.NewInt64Coin(stakeDenom, x) }
// 	now := tmtime.Now()

// 	// set up simapp and validators
// 	app := simapp.Setup(false)
// 	ctx := app.BaseApp.NewContext(false, tmproto.Header{}).WithBlockTime((now))
// 	valAddr, val := CreateValidator(ctx, app, 100)
// 	suite.Require().Equal("stake", app.StakingKeeper.BondDenom(ctx))

// 	lockupPeriods := sdkvesting.Periods{
// 		{Length: int64(12 * 3600), Amount: c(fee(1000), stake(100))}, // noon
// 	}
// 	vestingPeriods := sdkvesting.Periods{
// 		{Length: int64(8 * 3600), Amount: c(fee(200))},            // 8am
// 		{Length: int64(1 * 3600), Amount: c(fee(200), stake(50))}, // 9am
// 		{Length: int64(6 * 3600), Amount: c(fee(200), stake(50))}, // 3pm
// 		{Length: int64(2 * 3600), Amount: c(fee(200))},            // 5pm
// 		{Length: int64(1 * 3600), Amount: c(fee(200))},            // 6pm
// 	}

// 	bacc, origCoins := types.InitBaseAccount()
// 	_, _, funder := testdata.KeyTestPubAddr()
// 	va := types.NewClawbackVestingAccount(bacc, funder, origCoins, now.Unix(), lockupPeriods, vestingPeriods)
// 	// simulate 17stake lost to slashing
// 	va.DelegatedVesting = c(stake(17))
// 	addr := va.GetAddress()
// 	app.AccountKeeper.SetAccount(ctx, va)

// 	// fund the vesting account with 17 take lost to slashing
// 	err := simapp.FundAccount(app.BankKeeper, ctx, addr, c(fee(1000), stake(83)))
// 	suite.Require().NoError(err)
// 	suite.Require().Equal(int64(1000), app.BankKeeper.GetBalance(ctx, addr, feeDenom).Amount.Int64())
// 	suite.Require().Equal(int64(83), app.BankKeeper.GetBalance(ctx, addr, stakeDenom).Amount.Int64())
// 	ctx = ctx.WithBlockTime(now.Add(11 * time.Hour))

// 	// delegate 65
// 	shares, err := app.StakingKeeper.Delegate(ctx, addr, sdk.NewInt(65), stakingtypes.Unbonded, val, true)
// 	suite.Require().NoError(err)
// 	suite.Require().Equal(sdk.NewInt(65), shares.TruncateInt())
// 	suite.Require().Equal(int64(18), app.BankKeeper.GetBalance(ctx, addr, stakeDenom).Amount.Int64())

// 	// undelegate 5
// 	_, err = app.StakingKeeper.Undelegate(ctx, addr, valAddr, sdk.NewDec(5))
// 	suite.Require().NoError(err)

// 	// clawback the unvested funds (600fee, 50stake)
// 	_, _, dest := testdata.KeyTestPubAddr()
// 	va2 := app.AccountKeeper.GetAccount(ctx, addr).(*types.ClawbackVestingAccount)
// 	err = va2.Clawback(ctx, desapp.AccountKeeper, app.BankKeeper, app.StakingKeeper)
// 	suite.Require().NoError(err)

// 	// check vesting account
// 	// want 400fee, 33stake (delegated), all vested
// 	feeAmt := app.BankKeeper.GetBalance(ctx, addr, feeDenom).Amount
// 	suite.Require().Equal(int64(400), feeAmt.Int64())
// 	stakeAmt := app.BankKeeper.GetBalance(ctx, addr, stakeDenom).Amount
// 	suite.Require().Equal(int64(0), stakeAmt.Int64())
// 	del, found := app.StakingKeeper.GetDelegation(ctx, addr, valAddr)
// 	suite.Require().True(found)
// 	shares = del.GetShares()
// 	val, found = app.StakingKeeper.GetValidator(ctx, valAddr)
// 	suite.Require().True(found)
// 	stakeAmt = val.TokensFromSharesTruncated(shares).RoundInt()
// 	suite.Require().Equal(sdk.NewInt(33), stakeAmt)

// 	// check destination account
// 	// want 600fee, 50stake (18 unbonded, 5 unboinding, 27 bonded)
// 	feeAmt = app.BankKeeper.GetBalance(ctx, desfeeDenom).Amount
// 	suite.Require().Equal(int64(600), feeAmt.Int64())
// 	stakeAmt = app.BankKeeper.GetBalance(ctx, desstakeDenom).Amount
// 	suite.Require().Equal(int64(18), stakeAmt.Int64())
// 	del, found = app.StakingKeeper.GetDelegation(ctx, desvalAddr)
// 	suite.Require().True(found)
// 	shares = del.GetShares()
// 	stakeAmt = val.TokensFromSharesTruncated(shares).RoundInt()
// 	suite.Require().Equal(sdk.NewInt(27), stakeAmt)
// 	ubd, found := app.StakingKeeper.GetUnbondingDelegation(ctx, desvalAddr)
// 	suite.Require().True(found)
// 	suite.Require().Equal(sdk.NewInt(5), ubd.Entries[0].Balance)
// }

// // createValidator creates a validator in the given SimApp.
// func createValidator(, ctx sdk.Contexapp *simapp.SimApp, powers int64) (sdk.ValAddress, stakingtypes.Validator) {
// 	valTokens := sdk.TokensFromConsensusPower(powers, sdk.DefaultPowerReduction)
// 	addrs := simapp.AddTestAddrsIncremental(app, ctx, 1, valTokens)
// 	valAddrs := simapp.ConvertAddrsToValAddrs(addrs)
// 	pks := simapp.CreateTestPubKeys(1)
// 	cdc := app.AppCodec() //simapp.MakeTestEncodingConfig().Marshaler

// 	app.StakingKeeper = stakingkeeper.NewKeeper(
// 		cdc,
// 		app.GetKey(stakingtypes.StoreKey),
// 		app.AccountKeeper,
// 		app.BankKeeper,
// 		app.GetSubspace(stakingtypes.ModuleName),
// 	)

// 	val, err := stakingtypes.NewValidator(valAddrs[0], pks[0], stakingtypes.Description{})
// 	suite.Require().NoError(err)

// 	app.StakingKeeper.SetValidator(ctx, val)
// 	suite.Require().NoError(app.StakingKeeper.SetValidatorByConsAddr(ctx, val))
// 	app.StakingKeeper.SetNewValidatorByPowerIndex(ctx, val)

// 	_, err = app.StakingKeeper.Delegate(ctx, addrs[0], valTokens, stakingtypes.Unbonded, val, true)
// 	suite.Require().NoError(err)

// 	_ = staking.EndBlocker(ctx, app.StakingKeeper)

// 	return valAddrs[0], val
// }
