package types_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	tmtime "github.com/tendermint/tendermint/types/time"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	"github.com/cosmos/cosmos-sdk/simapp"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

var (
	stakeDenom = "stake"
	feeDenom   = "fee"
)

func contextAt(t time.Time) sdk.Context {
	return sdk.Context{}.WithBlockTime(t)
}

func TestGetVestedCoinsContVestingAcc(t *testing.T) {
	now := tmtime.Now()
	endTime := now.Add(24 * time.Hour)

	bacc, origCoins := initBaseAccount()
	cva := types.NewContinuousVestingAccount(bacc, origCoins, now.Unix(), endTime.Unix())

	// require no coins vested in the very beginning of the vesting schedule
	vestedCoins := cva.GetVestedCoins(now)
	require.Nil(t, vestedCoins)

	// require all coins vested at the end of the vesting schedule
	vestedCoins = cva.GetVestedCoins(endTime)
	require.Equal(t, origCoins, vestedCoins)

	// require 50% of coins vested
	vestedCoins = cva.GetVestedCoins(now.Add(12 * time.Hour))
	require.Equal(t, sdk.Coins{sdk.NewInt64Coin(feeDenom, 500), sdk.NewInt64Coin(stakeDenom, 50)}, vestedCoins)

	// require 100% of coins vested
	vestedCoins = cva.GetVestedCoins(now.Add(48 * time.Hour))
	require.Equal(t, origCoins, vestedCoins)
}

func TestGetVestingCoinsContVestingAcc(t *testing.T) {
	now := tmtime.Now()
	endTime := now.Add(24 * time.Hour)

	bacc, origCoins := initBaseAccount()
	cva := types.NewContinuousVestingAccount(bacc, origCoins, now.Unix(), endTime.Unix())

	// require all coins vesting in the beginning of the vesting schedule
	vestingCoins := cva.GetVestingCoins(now)
	require.Equal(t, origCoins, vestingCoins)

	// require no coins vesting at the end of the vesting schedule
	vestingCoins = cva.GetVestingCoins(endTime)
	require.Nil(t, vestingCoins)

	// require 50% of coins vesting
	vestingCoins = cva.GetVestingCoins(now.Add(12 * time.Hour))
	require.Equal(t, sdk.Coins{sdk.NewInt64Coin(feeDenom, 500), sdk.NewInt64Coin(stakeDenom, 50)}, vestingCoins)
}

func TestSpendableCoinsContVestingAcc(t *testing.T) {
	now := tmtime.Now()
	endTime := now.Add(24 * time.Hour)

	bacc, origCoins := initBaseAccount()
	cva := types.NewContinuousVestingAccount(bacc, origCoins, now.Unix(), endTime.Unix())

	// require that all original coins are locked at the end of the vesting
	// schedule
	lockedCoins := cva.LockedCoins(contextAt(now))
	require.Equal(t, origCoins, lockedCoins)

	// require that there exist no locked coins in the beginning of the
	lockedCoins = cva.LockedCoins(contextAt(endTime))
	require.Equal(t, sdk.NewCoins(), lockedCoins)

	// require that all vested coins (50%) are spendable
	lockedCoins = cva.LockedCoins(contextAt(now.Add(12 * time.Hour)))
	require.Equal(t, sdk.Coins{sdk.NewInt64Coin(feeDenom, 500), sdk.NewInt64Coin(stakeDenom, 50)}, lockedCoins)
}

func TestTrackDelegationContVestingAcc(t *testing.T) {
	now := tmtime.Now()
	endTime := now.Add(24 * time.Hour)

	bacc, origCoins := initBaseAccount()

	// require the ability to delegate all vesting coins
	cva := types.NewContinuousVestingAccount(bacc, origCoins, now.Unix(), endTime.Unix())
	cva.TrackDelegation(now, origCoins, origCoins)
	require.Equal(t, origCoins, cva.DelegatedVesting)
	require.Nil(t, cva.DelegatedFree)

	// require the ability to delegate all vested coins
	cva = types.NewContinuousVestingAccount(bacc, origCoins, now.Unix(), endTime.Unix())
	cva.TrackDelegation(endTime, origCoins, origCoins)
	require.Nil(t, cva.DelegatedVesting)
	require.Equal(t, origCoins, cva.DelegatedFree)

	// require the ability to delegate all vesting coins (50%) and all vested coins (50%)
	cva = types.NewContinuousVestingAccount(bacc, origCoins, now.Unix(), endTime.Unix())
	cva.TrackDelegation(now.Add(12*time.Hour), origCoins, sdk.Coins{sdk.NewInt64Coin(stakeDenom, 50)})
	require.Equal(t, sdk.Coins{sdk.NewInt64Coin(stakeDenom, 50)}, cva.DelegatedVesting)
	require.Nil(t, cva.DelegatedFree)

	cva.TrackDelegation(now.Add(12*time.Hour), origCoins, sdk.Coins{sdk.NewInt64Coin(stakeDenom, 50)})
	require.Equal(t, sdk.Coins{sdk.NewInt64Coin(stakeDenom, 50)}, cva.DelegatedVesting)
	require.Equal(t, sdk.Coins{sdk.NewInt64Coin(stakeDenom, 50)}, cva.DelegatedFree)

	// require no modifications when delegation amount is zero or not enough funds
	cva = types.NewContinuousVestingAccount(bacc, origCoins, now.Unix(), endTime.Unix())
	require.Panics(t, func() {
		cva.TrackDelegation(endTime, origCoins, sdk.Coins{sdk.NewInt64Coin(stakeDenom, 1000000)})
	})
	require.Nil(t, cva.DelegatedVesting)
	require.Nil(t, cva.DelegatedFree)
}

func TestTrackUndelegationContVestingAcc(t *testing.T) {
	now := tmtime.Now()
	endTime := now.Add(24 * time.Hour)

	bacc, origCoins := initBaseAccount()

	// require the ability to undelegate all vesting coins
	cva := types.NewContinuousVestingAccount(bacc, origCoins, now.Unix(), endTime.Unix())
	cva.TrackDelegation(now, origCoins, origCoins)
	cva.TrackUndelegation(origCoins)
	require.Nil(t, cva.DelegatedFree)
	require.Nil(t, cva.DelegatedVesting)

	// require the ability to undelegate all vested coins
	cva = types.NewContinuousVestingAccount(bacc, origCoins, now.Unix(), endTime.Unix())

	cva.TrackDelegation(endTime, origCoins, origCoins)
	cva.TrackUndelegation(origCoins)
	require.Nil(t, cva.DelegatedFree)
	require.Nil(t, cva.DelegatedVesting)

	// require no modifications when the undelegation amount is zero
	cva = types.NewContinuousVestingAccount(bacc, origCoins, now.Unix(), endTime.Unix())

	require.Panics(t, func() {
		cva.TrackUndelegation(sdk.Coins{sdk.NewInt64Coin(stakeDenom, 0)})
	})
	require.Nil(t, cva.DelegatedFree)
	require.Nil(t, cva.DelegatedVesting)

	// vest 50% and delegate to two validators
	cva = types.NewContinuousVestingAccount(bacc, origCoins, now.Unix(), endTime.Unix())
	cva.TrackDelegation(now.Add(12*time.Hour), origCoins, sdk.Coins{sdk.NewInt64Coin(stakeDenom, 50)})
	cva.TrackDelegation(now.Add(12*time.Hour), origCoins, sdk.Coins{sdk.NewInt64Coin(stakeDenom, 50)})

	// undelegate from one validator that got slashed 50%
	cva.TrackUndelegation(sdk.Coins{sdk.NewInt64Coin(stakeDenom, 25)})
	require.Equal(t, sdk.Coins{sdk.NewInt64Coin(stakeDenom, 25)}, cva.DelegatedFree)
	require.Equal(t, sdk.Coins{sdk.NewInt64Coin(stakeDenom, 50)}, cva.DelegatedVesting)

	// undelegate from the other validator that did not get slashed
	cva.TrackUndelegation(sdk.Coins{sdk.NewInt64Coin(stakeDenom, 50)})
	require.Nil(t, cva.DelegatedFree)
	require.Equal(t, sdk.Coins{sdk.NewInt64Coin(stakeDenom, 25)}, cva.DelegatedVesting)
}

func TestGetVestedCoinsDelVestingAcc(t *testing.T) {
	now := tmtime.Now()
	endTime := now.Add(24 * time.Hour)

	bacc, origCoins := initBaseAccount()

	// require no coins are vested until schedule maturation
	dva := types.NewDelayedVestingAccount(bacc, origCoins, endTime.Unix())
	vestedCoins := dva.GetVestedCoins(now)
	require.Nil(t, vestedCoins)

	// require all coins be vested at schedule maturation
	vestedCoins = dva.GetVestedCoins(endTime)
	require.Equal(t, origCoins, vestedCoins)
}

func TestGetVestingCoinsDelVestingAcc(t *testing.T) {
	now := tmtime.Now()
	endTime := now.Add(24 * time.Hour)

	bacc, origCoins := initBaseAccount()

	// require all coins vesting at the beginning of the schedule
	dva := types.NewDelayedVestingAccount(bacc, origCoins, endTime.Unix())
	vestingCoins := dva.GetVestingCoins(now)
	require.Equal(t, origCoins, vestingCoins)

	// require no coins vesting at schedule maturation
	vestingCoins = dva.GetVestingCoins(endTime)
	require.Nil(t, vestingCoins)
}

func TestSpendableCoinsDelVestingAcc(t *testing.T) {
	now := tmtime.Now()
	endTime := now.Add(24 * time.Hour)

	bacc, origCoins := initBaseAccount()

	// require that all coins are locked in the beginning of the vesting
	// schedule
	dva := types.NewDelayedVestingAccount(bacc, origCoins, endTime.Unix())
	lockedCoins := dva.LockedCoins(contextAt(now))
	require.True(t, lockedCoins.IsEqual(origCoins))

	// require that all coins are spendable after the maturation of the vesting
	// schedule
	lockedCoins = dva.LockedCoins(contextAt(endTime))
	require.Equal(t, sdk.NewCoins(), lockedCoins)

	// require that all coins are still vesting after some time
	lockedCoins = dva.LockedCoins(contextAt(now.Add(12 * time.Hour)))
	require.True(t, lockedCoins.IsEqual(origCoins))

	// delegate some locked coins
	// require that locked is reduced
	delegatedAmount := sdk.NewCoins(sdk.NewInt64Coin(stakeDenom, 50))
	dva.TrackDelegation(now.Add(12*time.Hour), origCoins, delegatedAmount)
	lockedCoins = dva.LockedCoins(contextAt(now.Add(12 * time.Hour)))
	require.True(t, lockedCoins.IsEqual(origCoins.Sub(delegatedAmount)))
}

func TestTrackDelegationDelVestingAcc(t *testing.T) {
	now := tmtime.Now()
	endTime := now.Add(24 * time.Hour)

	bacc, origCoins := initBaseAccount()

	// require the ability to delegate all vesting coins
	dva := types.NewDelayedVestingAccount(bacc, origCoins, endTime.Unix())
	dva.TrackDelegation(now, origCoins, origCoins)
	require.Equal(t, origCoins, dva.DelegatedVesting)
	require.Nil(t, dva.DelegatedFree)

	// require the ability to delegate all vested coins
	dva = types.NewDelayedVestingAccount(bacc, origCoins, endTime.Unix())
	dva.TrackDelegation(endTime, origCoins, origCoins)
	require.Nil(t, dva.DelegatedVesting)
	require.Equal(t, origCoins, dva.DelegatedFree)

	// require the ability to delegate all coins half way through the vesting
	// schedule
	dva = types.NewDelayedVestingAccount(bacc, origCoins, endTime.Unix())
	dva.TrackDelegation(now.Add(12*time.Hour), origCoins, origCoins)
	require.Equal(t, origCoins, dva.DelegatedVesting)
	require.Nil(t, dva.DelegatedFree)

	// require no modifications when delegation amount is zero or not enough funds
	dva = types.NewDelayedVestingAccount(bacc, origCoins, endTime.Unix())

	require.Panics(t, func() {
		dva.TrackDelegation(endTime, origCoins, sdk.Coins{sdk.NewInt64Coin(stakeDenom, 1000000)})
	})
	require.Nil(t, dva.DelegatedVesting)
	require.Nil(t, dva.DelegatedFree)
}

func TestTrackUndelegationDelVestingAcc(t *testing.T) {
	now := tmtime.Now()
	endTime := now.Add(24 * time.Hour)

	bacc, origCoins := initBaseAccount()

	// require the ability to undelegate all vesting coins
	dva := types.NewDelayedVestingAccount(bacc, origCoins, endTime.Unix())
	dva.TrackDelegation(now, origCoins, origCoins)
	dva.TrackUndelegation(origCoins)
	require.Nil(t, dva.DelegatedFree)
	require.Nil(t, dva.DelegatedVesting)

	// require the ability to undelegate all vested coins
	dva = types.NewDelayedVestingAccount(bacc, origCoins, endTime.Unix())
	dva.TrackDelegation(endTime, origCoins, origCoins)
	dva.TrackUndelegation(origCoins)
	require.Nil(t, dva.DelegatedFree)
	require.Nil(t, dva.DelegatedVesting)

	// require no modifications when the undelegation amount is zero
	dva = types.NewDelayedVestingAccount(bacc, origCoins, endTime.Unix())

	require.Panics(t, func() {
		dva.TrackUndelegation(sdk.Coins{sdk.NewInt64Coin(stakeDenom, 0)})
	})
	require.Nil(t, dva.DelegatedFree)
	require.Nil(t, dva.DelegatedVesting)

	// vest 50% and delegate to two validators
	dva = types.NewDelayedVestingAccount(bacc, origCoins, endTime.Unix())
	dva.TrackDelegation(now.Add(12*time.Hour), origCoins, sdk.Coins{sdk.NewInt64Coin(stakeDenom, 50)})
	dva.TrackDelegation(now.Add(12*time.Hour), origCoins, sdk.Coins{sdk.NewInt64Coin(stakeDenom, 50)})

	// undelegate from one validator that got slashed 50%
	dva.TrackUndelegation(sdk.Coins{sdk.NewInt64Coin(stakeDenom, 25)})

	require.Nil(t, dva.DelegatedFree)
	require.Equal(t, sdk.Coins{sdk.NewInt64Coin(stakeDenom, 75)}, dva.DelegatedVesting)

	// undelegate from the other validator that did not get slashed
	dva.TrackUndelegation(sdk.Coins{sdk.NewInt64Coin(stakeDenom, 50)})
	require.Nil(t, dva.DelegatedFree)
	require.Equal(t, sdk.Coins{sdk.NewInt64Coin(stakeDenom, 25)}, dva.DelegatedVesting)
}

func TestAddGrantPeriodicVestingAcc(t *testing.T) {
	c := sdk.NewCoins
	fee := func(amt int64) sdk.Coin { return sdk.NewInt64Coin(feeDenom, amt) }
	stake := func(amt int64) sdk.Coin { return sdk.NewInt64Coin(stakeDenom, amt) }
	now := tmtime.Now()

	// set up simapp
	app := simapp.Setup(false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{}).WithBlockTime((now))
	require.Equal(t, "stake", app.StakingKeeper.BondDenom(ctx))

	// create an account with an initial grant
	periods := types.Periods{
		{Length: 100, Amount: c(fee(650), stake(40))},
		{Length: 100, Amount: c(fee(350), stake(60))},
	}
	bacc, origCoins := initBaseAccount()
	pva := types.NewPeriodicVestingAccount(bacc, origCoins, now.Unix(), periods)
	addr := pva.GetAddress()

	// simulate 54stake (unvested) lost to slashing
	pva.DelegatedVesting = c(stake(54))

	// Only 6stake are locked at now+150, due to slashing
	require.Equal(t, int64(6), pva.LockedCoins(ctx.WithBlockTime(now.Add(150*time.Second))).AmountOf(stakeDenom).Int64())

	// Add a new grant of 50stake
	newGrant := c(stake(50))
	pva.AddGrant(ctx, app.StakingKeeper, now.Add(500*time.Second).Unix(), []types.Period{{Length: 50, Amount: newGrant}}, newGrant)
	app.AccountKeeper.SetAccount(ctx, pva)

	// Only 56stake locked at now+150, due to slashing
	require.Equal(t, int64(56), pva.LockedCoins(ctx.WithBlockTime(now.Add(150*time.Second))).AmountOf(stakeDenom).Int64())

	// fund the vesting account with new grant (old has vested and transferred out)
	ctx = ctx.WithBlockTime(now.Add(500 * time.Second))
	err := simapp.FundAccount(app.BankKeeper, ctx, addr, newGrant)
	require.NoError(t, err)
	require.Equal(t, int64(50), app.BankKeeper.GetBalance(ctx, addr, stakeDenom).Amount.Int64())

	// we should not be able to transfer the new grant out until it has vested
	_, _, dest := testdata.KeyTestPubAddr()
	err = app.BankKeeper.SendCoins(ctx, addr, dest, c(stake(1)))
	require.Error(t, err)
	ctx = ctx.WithBlockTime(now.Add(600 * time.Second))
	err = app.BankKeeper.SendCoins(ctx, addr, dest, newGrant)
	require.NoError(t, err)
}

func TestAddGrantPeriodicVestingAcc_FullSlash(t *testing.T) {
	c := sdk.NewCoins
	stake := func(amt int64) sdk.Coin { return sdk.NewInt64Coin(stakeDenom, amt) }
	now := tmtime.Now()

	// set up simapp
	app := simapp.Setup(false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{}).WithBlockTime((now))
	require.Equal(t, "stake", app.StakingKeeper.BondDenom(ctx))

	// create an account with an initial grant
	periods := types.Periods{
		{Length: 100, Amount: c(stake(40))},
		{Length: 100, Amount: c(stake(60))},
	}
	bacc, _ := initBaseAccount()
	origCoins := c(stake(100))
	pva := types.NewPeriodicVestingAccount(bacc, origCoins, now.Unix(), periods)
	addr := pva.GetAddress()

	// simulate all 100stake lost to slashing
	pva.DelegatedVesting = c(stake(100))

	// Nothing locked at now+150 since all unvested lost to slashign
	require.Equal(t, int64(0), pva.LockedCoins(ctx.WithBlockTime(now.Add(150*time.Second))).AmountOf(stakeDenom).Int64())

	// Add a new grant of 50stake
	newGrant := c(stake(50))
	pva.AddGrant(ctx, app.StakingKeeper, now.Add(500*time.Second).Unix(), []types.Period{{Length: 50, Amount: newGrant}}, newGrant)
	app.AccountKeeper.SetAccount(ctx, pva)

	// The new 50 are locked at now+150
	require.Equal(t, int64(50), pva.LockedCoins(ctx.WithBlockTime(now.Add(150*time.Second))).AmountOf(stakeDenom).Int64())

	// fund the vesting account with new grant (old has vested and transferred out)
	ctx = ctx.WithBlockTime(now.Add(500 * time.Second))
	err := simapp.FundAccount(app.BankKeeper, ctx, addr, newGrant)
	require.NoError(t, err)
	require.Equal(t, int64(50), app.BankKeeper.GetBalance(ctx, addr, stakeDenom).Amount.Int64())

	// we should not be able to transfer the new grant out until it has vested
	_, _, dest := testdata.KeyTestPubAddr()
	err = app.BankKeeper.SendCoins(ctx, addr, dest, c(stake(1)))
	require.Error(t, err)
	ctx = ctx.WithBlockTime(now.Add(600 * time.Second))
	err = app.BankKeeper.SendCoins(ctx, addr, dest, newGrant)
	require.NoError(t, err)
}

func TestGetVestedCoinsPeriodicVestingAcc(t *testing.T) {
	now := tmtime.Now()
	endTime := now.Add(24 * time.Hour)
	periods := types.Periods{
		types.Period{Length: int64(12 * 60 * 60), Amount: sdk.Coins{sdk.NewInt64Coin(feeDenom, 500), sdk.NewInt64Coin(stakeDenom, 50)}},
		types.Period{Length: int64(6 * 60 * 60), Amount: sdk.Coins{sdk.NewInt64Coin(feeDenom, 250), sdk.NewInt64Coin(stakeDenom, 25)}},
		types.Period{Length: int64(6 * 60 * 60), Amount: sdk.Coins{sdk.NewInt64Coin(feeDenom, 250), sdk.NewInt64Coin(stakeDenom, 25)}},
	}

	bacc, origCoins := initBaseAccount()
	pva := types.NewPeriodicVestingAccount(bacc, origCoins, now.Unix(), periods)

	// require no coins vested at the beginning of the vesting schedule
	vestedCoins := pva.GetVestedCoins(now)
	require.Nil(t, vestedCoins)

	// require all coins vested at the end of the vesting schedule
	vestedCoins = pva.GetVestedCoins(endTime)
	require.Equal(t, origCoins, vestedCoins)

	// require no coins vested during first vesting period
	vestedCoins = pva.GetVestedCoins(now.Add(6 * time.Hour))
	require.Nil(t, vestedCoins)

	// require 50% of coins vested after period 1
	vestedCoins = pva.GetVestedCoins(now.Add(12 * time.Hour))
	require.Equal(t, sdk.Coins{sdk.NewInt64Coin(feeDenom, 500), sdk.NewInt64Coin(stakeDenom, 50)}, vestedCoins)

	// require period 2 coins don't vest until period is over
	vestedCoins = pva.GetVestedCoins(now.Add(15 * time.Hour))
	require.Equal(t, sdk.Coins{sdk.NewInt64Coin(feeDenom, 500), sdk.NewInt64Coin(stakeDenom, 50)}, vestedCoins)

	// require 75% of coins vested after period 2
	vestedCoins = pva.GetVestedCoins(now.Add(18 * time.Hour))
	require.Equal(t,
		sdk.Coins{
			sdk.NewInt64Coin(feeDenom, 750), sdk.NewInt64Coin(stakeDenom, 75)}, vestedCoins)

	// require 100% of coins vested
	vestedCoins = pva.GetVestedCoins(now.Add(48 * time.Hour))
	require.Equal(t, origCoins, vestedCoins)
}

func TestGetVestingCoinsPeriodicVestingAcc(t *testing.T) {
	now := tmtime.Now()
	endTime := now.Add(24 * time.Hour)
	periods := types.Periods{
		types.Period{Length: int64(12 * 60 * 60), Amount: sdk.Coins{sdk.NewInt64Coin(feeDenom, 500), sdk.NewInt64Coin(stakeDenom, 50)}},
		types.Period{Length: int64(6 * 60 * 60), Amount: sdk.Coins{sdk.NewInt64Coin(feeDenom, 250), sdk.NewInt64Coin(stakeDenom, 25)}},
		types.Period{Length: int64(6 * 60 * 60), Amount: sdk.Coins{sdk.NewInt64Coin(feeDenom, 250), sdk.NewInt64Coin(stakeDenom, 25)}},
	}

	bacc, origCoins := initBaseAccount()
	pva := types.NewPeriodicVestingAccount(bacc, origCoins, now.Unix(), periods)

	// require all coins vesting at the beginning of the vesting schedule
	vestingCoins := pva.GetVestingCoins(now)
	require.Equal(t, origCoins, vestingCoins)

	// require no coins vesting at the end of the vesting schedule
	vestingCoins = pva.GetVestingCoins(endTime)
	require.Nil(t, vestingCoins)

	// require 50% of coins vesting
	vestingCoins = pva.GetVestingCoins(now.Add(12 * time.Hour))
	require.Equal(t, sdk.Coins{sdk.NewInt64Coin(feeDenom, 500), sdk.NewInt64Coin(stakeDenom, 50)}, vestingCoins)

	// require 50% of coins vesting after period 1, but before period 2 completes.
	vestingCoins = pva.GetVestingCoins(now.Add(15 * time.Hour))
	require.Equal(t, sdk.Coins{sdk.NewInt64Coin(feeDenom, 500), sdk.NewInt64Coin(stakeDenom, 50)}, vestingCoins)

	// require 25% of coins vesting after period 2
	vestingCoins = pva.GetVestingCoins(now.Add(18 * time.Hour))
	require.Equal(t, sdk.Coins{sdk.NewInt64Coin(feeDenom, 250), sdk.NewInt64Coin(stakeDenom, 25)}, vestingCoins)

	// require 0% of coins vesting after vesting complete
	vestingCoins = pva.GetVestingCoins(now.Add(48 * time.Hour))
	require.Nil(t, vestingCoins)
}

func TestSpendableCoinsPeriodicVestingAcc(t *testing.T) {
	now := tmtime.Now()
	endTime := now.Add(24 * time.Hour)
	periods := types.Periods{
		types.Period{Length: int64(12 * 60 * 60), Amount: sdk.Coins{sdk.NewInt64Coin(feeDenom, 500), sdk.NewInt64Coin(stakeDenom, 50)}},
		types.Period{Length: int64(6 * 60 * 60), Amount: sdk.Coins{sdk.NewInt64Coin(feeDenom, 250), sdk.NewInt64Coin(stakeDenom, 25)}},
		types.Period{Length: int64(6 * 60 * 60), Amount: sdk.Coins{sdk.NewInt64Coin(feeDenom, 250), sdk.NewInt64Coin(stakeDenom, 25)}},
	}

	bacc, origCoins := initBaseAccount()
	pva := types.NewPeriodicVestingAccount(bacc, origCoins, now.Unix(), periods)

	// require that there exist no spendable coins at the beginning of the
	// vesting schedule
	lockedCoins := pva.LockedCoins(contextAt(now))
	require.Equal(t, origCoins, lockedCoins)

	// require that all original coins are spendable at the end of the vesting
	// schedule
	lockedCoins = pva.LockedCoins(contextAt(endTime))
	require.Equal(t, sdk.NewCoins(), lockedCoins)

	// require that all still vesting coins (50%) are locked
	lockedCoins = pva.LockedCoins(contextAt(now.Add(12 * time.Hour)))
	require.Equal(t, sdk.Coins{sdk.NewInt64Coin(feeDenom, 500), sdk.NewInt64Coin(stakeDenom, 50)}, lockedCoins)
}

func TestTrackDelegationPeriodicVestingAcc(t *testing.T) {
	now := tmtime.Now()
	endTime := now.Add(24 * time.Hour)
	periods := types.Periods{
		types.Period{Length: int64(12 * 60 * 60), Amount: sdk.Coins{sdk.NewInt64Coin(feeDenom, 500), sdk.NewInt64Coin(stakeDenom, 50)}},
		types.Period{Length: int64(6 * 60 * 60), Amount: sdk.Coins{sdk.NewInt64Coin(feeDenom, 250), sdk.NewInt64Coin(stakeDenom, 25)}},
		types.Period{Length: int64(6 * 60 * 60), Amount: sdk.Coins{sdk.NewInt64Coin(feeDenom, 250), sdk.NewInt64Coin(stakeDenom, 25)}},
	}

	bacc, origCoins := initBaseAccount()

	// require the ability to delegate all vesting coins
	pva := types.NewPeriodicVestingAccount(bacc, origCoins, now.Unix(), periods)
	pva.TrackDelegation(now, origCoins, origCoins)
	require.Equal(t, origCoins, pva.DelegatedVesting)
	require.Nil(t, pva.DelegatedFree)

	// require the ability to delegate all vested coins
	pva = types.NewPeriodicVestingAccount(bacc, origCoins, now.Unix(), periods)
	pva.TrackDelegation(endTime, origCoins, origCoins)
	require.Nil(t, pva.DelegatedVesting)
	require.Equal(t, origCoins, pva.DelegatedFree)

	// delegate half of vesting coins
	pva = types.NewPeriodicVestingAccount(bacc, origCoins, now.Unix(), periods)
	pva.TrackDelegation(now, origCoins, periods[0].Amount)
	// require that all delegated coins are delegated vesting
	require.Equal(t, pva.DelegatedVesting, periods[0].Amount)
	require.Nil(t, pva.DelegatedFree)

	// delegate 75% of coins, split between vested and vesting
	pva = types.NewPeriodicVestingAccount(bacc, origCoins, now.Unix(), periods)
	pva.TrackDelegation(now.Add(12*time.Hour), origCoins, periods[0].Amount.Add(periods[1].Amount...))
	// require that the maximum possible amount of vesting coins are chosen for delegation.
	require.Equal(t, pva.DelegatedFree, periods[1].Amount)
	require.Equal(t, pva.DelegatedVesting, periods[0].Amount)

	// require the ability to delegate all vesting coins (50%) and all vested coins (50%)
	pva = types.NewPeriodicVestingAccount(bacc, origCoins, now.Unix(), periods)
	pva.TrackDelegation(now.Add(12*time.Hour), origCoins, sdk.Coins{sdk.NewInt64Coin(stakeDenom, 50)})
	require.Equal(t, sdk.Coins{sdk.NewInt64Coin(stakeDenom, 50)}, pva.DelegatedVesting)
	require.Nil(t, pva.DelegatedFree)

	pva.TrackDelegation(now.Add(12*time.Hour), origCoins, sdk.Coins{sdk.NewInt64Coin(stakeDenom, 50)})
	require.Equal(t, sdk.Coins{sdk.NewInt64Coin(stakeDenom, 50)}, pva.DelegatedVesting)
	require.Equal(t, sdk.Coins{sdk.NewInt64Coin(stakeDenom, 50)}, pva.DelegatedFree)

	// require no modifications when delegation amount is zero or not enough funds
	pva = types.NewPeriodicVestingAccount(bacc, origCoins, now.Unix(), periods)
	require.Panics(t, func() {
		pva.TrackDelegation(endTime, origCoins, sdk.Coins{sdk.NewInt64Coin(stakeDenom, 1000000)})
	})
	require.Nil(t, pva.DelegatedVesting)
	require.Nil(t, pva.DelegatedFree)
}

func TestTrackUndelegationPeriodicVestingAcc(t *testing.T) {
	now := tmtime.Now()
	endTime := now.Add(24 * time.Hour)
	periods := types.Periods{
		types.Period{Length: int64(12 * 60 * 60), Amount: sdk.Coins{sdk.NewInt64Coin(feeDenom, 500), sdk.NewInt64Coin(stakeDenom, 50)}},
		types.Period{Length: int64(6 * 60 * 60), Amount: sdk.Coins{sdk.NewInt64Coin(feeDenom, 250), sdk.NewInt64Coin(stakeDenom, 25)}},
		types.Period{Length: int64(6 * 60 * 60), Amount: sdk.Coins{sdk.NewInt64Coin(feeDenom, 250), sdk.NewInt64Coin(stakeDenom, 25)}},
	}

	bacc, origCoins := initBaseAccount()

	// require the ability to undelegate all vesting coins at the beginning of vesting
	pva := types.NewPeriodicVestingAccount(bacc, origCoins, now.Unix(), periods)
	pva.TrackDelegation(now, origCoins, origCoins)
	pva.TrackUndelegation(origCoins)
	require.Nil(t, pva.DelegatedFree)
	require.Nil(t, pva.DelegatedVesting)

	// require the ability to undelegate all vested coins at the end of vesting
	pva = types.NewPeriodicVestingAccount(bacc, origCoins, now.Unix(), periods)

	pva.TrackDelegation(endTime, origCoins, origCoins)
	pva.TrackUndelegation(origCoins)
	require.Nil(t, pva.DelegatedFree)
	require.Nil(t, pva.DelegatedVesting)

	// require the ability to undelegate half of coins
	pva = types.NewPeriodicVestingAccount(bacc, origCoins, now.Unix(), periods)
	pva.TrackDelegation(endTime, origCoins, periods[0].Amount)
	pva.TrackUndelegation(periods[0].Amount)
	require.Nil(t, pva.DelegatedFree)
	require.Nil(t, pva.DelegatedVesting)

	// require no modifications when the undelegation amount is zero
	pva = types.NewPeriodicVestingAccount(bacc, origCoins, now.Unix(), periods)

	require.Panics(t, func() {
		pva.TrackUndelegation(sdk.Coins{sdk.NewInt64Coin(stakeDenom, 0)})
	})
	require.Nil(t, pva.DelegatedFree)
	require.Nil(t, pva.DelegatedVesting)

	// vest 50% and delegate to two validators
	pva = types.NewPeriodicVestingAccount(bacc, origCoins, now.Unix(), periods)
	pva.TrackDelegation(now.Add(12*time.Hour), origCoins, sdk.Coins{sdk.NewInt64Coin(stakeDenom, 50)})
	pva.TrackDelegation(now.Add(12*time.Hour), origCoins, sdk.Coins{sdk.NewInt64Coin(stakeDenom, 50)})

	// undelegate from one validator that got slashed 50%
	pva.TrackUndelegation(sdk.Coins{sdk.NewInt64Coin(stakeDenom, 25)})
	require.Equal(t, sdk.Coins{sdk.NewInt64Coin(stakeDenom, 25)}, pva.DelegatedFree)
	require.Equal(t, sdk.Coins{sdk.NewInt64Coin(stakeDenom, 50)}, pva.DelegatedVesting)

	// undelegate from the other validator that did not get slashed
	pva.TrackUndelegation(sdk.Coins{sdk.NewInt64Coin(stakeDenom, 50)})
	require.Nil(t, pva.DelegatedFree)
	require.Equal(t, sdk.Coins{sdk.NewInt64Coin(stakeDenom, 25)}, pva.DelegatedVesting)
}

func TestGetVestedCoinsPermLockedVestingAcc(t *testing.T) {
	now := tmtime.Now()
	endTime := now.Add(1000 * 24 * time.Hour)

	bacc, origCoins := initBaseAccount()

	// require no coins are vested
	plva := types.NewPermanentLockedAccount(bacc, origCoins)
	vestedCoins := plva.GetVestedCoins(now)
	require.Nil(t, vestedCoins)

	// require no coins be vested at end time
	vestedCoins = plva.GetVestedCoins(endTime)
	require.Nil(t, vestedCoins)
}

func TestGetVestingCoinsPermLockedVestingAcc(t *testing.T) {
	now := tmtime.Now()
	endTime := now.Add(1000 * 24 * time.Hour)

	bacc, origCoins := initBaseAccount()

	// require all coins vesting at the beginning of the schedule
	plva := types.NewPermanentLockedAccount(bacc, origCoins)
	vestingCoins := plva.GetVestingCoins(now)
	require.Equal(t, origCoins, vestingCoins)

	// require all coins vesting at the end time
	vestingCoins = plva.GetVestingCoins(endTime)
	require.Equal(t, origCoins, vestingCoins)
}

func TestSpendableCoinsPermLockedVestingAcc(t *testing.T) {
	now := tmtime.Now()
	endTime := now.Add(1000 * 24 * time.Hour)

	bacc, origCoins := initBaseAccount()

	// require that all coins are locked in the beginning of the vesting
	// schedule
	plva := types.NewPermanentLockedAccount(bacc, origCoins)
	lockedCoins := plva.LockedCoins(contextAt(now))
	require.True(t, lockedCoins.IsEqual(origCoins))

	// require that all coins are still locked at end time
	lockedCoins = plva.LockedCoins(contextAt(endTime))
	require.True(t, lockedCoins.IsEqual(origCoins))

	// delegate some locked coins
	// require that locked is reduced
	delegatedAmount := sdk.NewCoins(sdk.NewInt64Coin(stakeDenom, 50))
	plva.TrackDelegation(now.Add(12*time.Hour), origCoins, delegatedAmount)
	lockedCoins = plva.LockedCoins(contextAt(now.Add(12 * time.Hour)))
	require.True(t, lockedCoins.IsEqual(origCoins.Sub(delegatedAmount)))
}

func TestTrackDelegationPermLockedVestingAcc(t *testing.T) {
	now := tmtime.Now()
	endTime := now.Add(1000 * 24 * time.Hour)

	bacc, origCoins := initBaseAccount()

	// require the ability to delegate all vesting coins
	plva := types.NewPermanentLockedAccount(bacc, origCoins)
	plva.TrackDelegation(now, origCoins, origCoins)
	require.Equal(t, origCoins, plva.DelegatedVesting)
	require.Nil(t, plva.DelegatedFree)

	// require the ability to delegate all vested coins at endTime
	plva = types.NewPermanentLockedAccount(bacc, origCoins)
	plva.TrackDelegation(endTime, origCoins, origCoins)
	require.Equal(t, origCoins, plva.DelegatedVesting)
	require.Nil(t, plva.DelegatedFree)

	// require no modifications when delegation amount is zero or not enough funds
	plva = types.NewPermanentLockedAccount(bacc, origCoins)

	require.Panics(t, func() {
		plva.TrackDelegation(endTime, origCoins, sdk.Coins{sdk.NewInt64Coin(stakeDenom, 1000000)})
	})
	require.Nil(t, plva.DelegatedVesting)
	require.Nil(t, plva.DelegatedFree)
}

func TestTrackUndelegationPermLockedVestingAcc(t *testing.T) {
	now := tmtime.Now()
	endTime := now.Add(1000 * 24 * time.Hour)

	bacc, origCoins := initBaseAccount()

	// require the ability to undelegate all vesting coins
	plva := types.NewPermanentLockedAccount(bacc, origCoins)
	plva.TrackDelegation(now, origCoins, origCoins)
	plva.TrackUndelegation(origCoins)
	require.Nil(t, plva.DelegatedFree)
	require.Nil(t, plva.DelegatedVesting)

	// require the ability to undelegate all vesting coins at endTime
	plva = types.NewPermanentLockedAccount(bacc, origCoins)
	plva.TrackDelegation(endTime, origCoins, origCoins)
	plva.TrackUndelegation(origCoins)
	require.Nil(t, plva.DelegatedFree)
	require.Nil(t, plva.DelegatedVesting)

	// require no modifications when the undelegation amount is zero
	plva = types.NewPermanentLockedAccount(bacc, origCoins)
	require.Panics(t, func() {
		plva.TrackUndelegation(sdk.Coins{sdk.NewInt64Coin(stakeDenom, 0)})
	})
	require.Nil(t, plva.DelegatedFree)
	require.Nil(t, plva.DelegatedVesting)

	// delegate to two validators
	plva = types.NewPermanentLockedAccount(bacc, origCoins)
	plva.TrackDelegation(now, origCoins, sdk.Coins{sdk.NewInt64Coin(stakeDenom, 50)})
	plva.TrackDelegation(now, origCoins, sdk.Coins{sdk.NewInt64Coin(stakeDenom, 50)})

	// undelegate from one validator that got slashed 50%
	plva.TrackUndelegation(sdk.Coins{sdk.NewInt64Coin(stakeDenom, 25)})

	require.Nil(t, plva.DelegatedFree)
	require.Equal(t, sdk.Coins{sdk.NewInt64Coin(stakeDenom, 75)}, plva.DelegatedVesting)

	// undelegate from the other validator that did not get slashed
	plva.TrackUndelegation(sdk.Coins{sdk.NewInt64Coin(stakeDenom, 50)})
	require.Nil(t, plva.DelegatedFree)
	require.Equal(t, sdk.Coins{sdk.NewInt64Coin(stakeDenom, 25)}, plva.DelegatedVesting)
}

func TestGetVestedCoinsClawbackVestingAcc(t *testing.T) {
	now := tmtime.Now()
	endTime := now.Add(24 * time.Hour)
	lockupPeriods := types.Periods{
		types.Period{Length: int64(16 * 60 * 60), Amount: sdk.NewCoins(sdk.NewInt64Coin(feeDenom, 1000), sdk.NewInt64Coin(stakeDenom, 100))},
	}
	vestingPeriods := types.Periods{
		types.Period{Length: int64(12 * 60 * 60), Amount: sdk.Coins{sdk.NewInt64Coin(feeDenom, 500), sdk.NewInt64Coin(stakeDenom, 50)}},
		types.Period{Length: int64(6 * 60 * 60), Amount: sdk.Coins{sdk.NewInt64Coin(feeDenom, 250), sdk.NewInt64Coin(stakeDenom, 25)}},
		types.Period{Length: int64(6 * 60 * 60), Amount: sdk.Coins{sdk.NewInt64Coin(feeDenom, 250), sdk.NewInt64Coin(stakeDenom, 25)}},
	}

	bacc, origCoins := initBaseAccount()
	va := types.NewClawbackVestingAccount(bacc, sdk.AccAddress([]byte("funder")), origCoins, now.Unix(), lockupPeriods, vestingPeriods)

	// require no coins vested at the beginning of the vesting schedule
	vestedCoins := va.GetVestedCoins(now)
	require.Nil(t, vestedCoins)

	// require all coins vested at the end of the vesting schedule
	vestedCoins = va.GetVestedCoins(endTime)
	require.Equal(t, origCoins, vestedCoins)

	// require no coins vested during first vesting period
	vestedCoins = va.GetVestedCoins(now.Add(6 * time.Hour))
	require.Nil(t, vestedCoins)

	// require no coins vested after period1 before unlocking
	vestedCoins = va.GetVestedCoins(now.Add(14 * time.Hour))
	require.Nil(t, vestedCoins)

	// require 50% of coins vested after period 1 at unlocking
	vestedCoins = va.GetVestedCoins(now.Add(16 * time.Hour))
	require.Equal(t, sdk.Coins{sdk.NewInt64Coin(feeDenom, 500), sdk.NewInt64Coin(stakeDenom, 50)}, vestedCoins)

	// require period 2 coins don't vest until period is over
	vestedCoins = va.GetVestedCoins(now.Add(17 * time.Hour))
	require.Equal(t, sdk.Coins{sdk.NewInt64Coin(feeDenom, 500), sdk.NewInt64Coin(stakeDenom, 50)}, vestedCoins)

	// require 75% of coins vested after period 2
	vestedCoins = va.GetVestedCoins(now.Add(18 * time.Hour))
	require.Equal(t,
		sdk.Coins{
			sdk.NewInt64Coin(feeDenom, 750), sdk.NewInt64Coin(stakeDenom, 75)}, vestedCoins)

	// require 100% of coins vested
	vestedCoins = va.GetVestedCoins(now.Add(48 * time.Hour))
	require.Equal(t, origCoins, vestedCoins)
}

func TestGetVestingCoinsClawbackVestingAcc(t *testing.T) {
	now := tmtime.Now()
	endTime := now.Add(24 * time.Hour)
	lockupPeriods := types.Periods{
		types.Period{Length: int64(16 * 60 * 60), Amount: sdk.NewCoins(sdk.NewInt64Coin(feeDenom, 1000), sdk.NewInt64Coin(stakeDenom, 100))},
	}
	vestingPeriods := types.Periods{
		types.Period{Length: int64(12 * 60 * 60), Amount: sdk.Coins{sdk.NewInt64Coin(feeDenom, 500), sdk.NewInt64Coin(stakeDenom, 50)}},
		types.Period{Length: int64(6 * 60 * 60), Amount: sdk.Coins{sdk.NewInt64Coin(feeDenom, 250), sdk.NewInt64Coin(stakeDenom, 25)}},
		types.Period{Length: int64(6 * 60 * 60), Amount: sdk.Coins{sdk.NewInt64Coin(feeDenom, 250), sdk.NewInt64Coin(stakeDenom, 25)}},
	}

	bacc, origCoins := initBaseAccount()
	va := types.NewClawbackVestingAccount(bacc, sdk.AccAddress([]byte("funder")), origCoins, now.Unix(), lockupPeriods, vestingPeriods)

	// require all coins vesting at the beginning of the vesting schedule
	vestingCoins := va.GetVestingCoins(now)
	require.Equal(t, origCoins, vestingCoins)

	// require no coins vesting at the end of the vesting schedule
	vestingCoins = va.GetVestingCoins(endTime)
	require.Nil(t, vestingCoins)

	// require all coins vesting at first vesting event
	vestingCoins = va.GetVestingCoins(now.Add(12 * time.Hour))
	require.Equal(t, origCoins, vestingCoins)

	// require all coins vesting after period 1 before unlocking
	vestingCoins = va.GetVestingCoins(now.Add(15 * time.Hour))
	require.Equal(t, origCoins, vestingCoins)

	// require 50% of coins vesting after period 1 at unlocking
	vestingCoins = va.GetVestingCoins(now.Add(16 * time.Hour))
	require.Equal(t, sdk.Coins{sdk.NewInt64Coin(feeDenom, 500), sdk.NewInt64Coin(stakeDenom, 50)}, vestingCoins)

	// require 50% of coins vesting after period 1, after unlocking
	vestingCoins = va.GetVestingCoins(now.Add(17 * time.Hour))
	require.Equal(t, sdk.Coins{sdk.NewInt64Coin(feeDenom, 500), sdk.NewInt64Coin(stakeDenom, 50)}, vestingCoins)

	// require 25% of coins vesting after period 2
	vestingCoins = va.GetVestingCoins(now.Add(18 * time.Hour))
	require.Equal(t, sdk.Coins{sdk.NewInt64Coin(feeDenom, 250), sdk.NewInt64Coin(stakeDenom, 25)}, vestingCoins)

	// require 0% of coins vesting after vesting complete
	vestingCoins = va.GetVestingCoins(now.Add(48 * time.Hour))
	require.Nil(t, vestingCoins)
}

func TestSpendableCoinsClawbackVestingAcc(t *testing.T) {
	now := tmtime.Now()
	endTime := now.Add(24 * time.Hour)
	lockupPeriods := types.Periods{
		types.Period{Length: int64(16 * 60 * 60), Amount: sdk.NewCoins(sdk.NewInt64Coin(feeDenom, 1000), sdk.NewInt64Coin(stakeDenom, 100))},
	}
	vestingPeriods := types.Periods{
		types.Period{Length: int64(12 * 60 * 60), Amount: sdk.Coins{sdk.NewInt64Coin(feeDenom, 500), sdk.NewInt64Coin(stakeDenom, 50)}},
		types.Period{Length: int64(6 * 60 * 60), Amount: sdk.Coins{sdk.NewInt64Coin(feeDenom, 250), sdk.NewInt64Coin(stakeDenom, 25)}},
		types.Period{Length: int64(6 * 60 * 60), Amount: sdk.Coins{sdk.NewInt64Coin(feeDenom, 250), sdk.NewInt64Coin(stakeDenom, 25)}},
	}

	bacc, origCoins := initBaseAccount()
	va := types.NewClawbackVestingAccount(bacc, sdk.AccAddress([]byte("funder")), origCoins, now.Unix(), lockupPeriods, vestingPeriods)

	// require that there exist no spendable coins at the beginning of the
	// vesting schedule
	lockedCoins := va.LockedCoins(contextAt(now))
	require.Equal(t, origCoins, lockedCoins)

	// require that all original coins are spendable at the end of the vesting
	// schedule
	lockedCoins = va.LockedCoins(contextAt(endTime))
	require.Equal(t, sdk.NewCoins(), lockedCoins)

	// require that all still vesting coins (50%) are locked
	lockedCoins = va.LockedCoins(contextAt(now.Add(17 * time.Hour)))
	require.Equal(t, sdk.Coins{sdk.NewInt64Coin(feeDenom, 500), sdk.NewInt64Coin(stakeDenom, 50)}, lockedCoins)
}

func TestTrackDelegationClawbackVestingAcc(t *testing.T) {
	now := tmtime.Now()
	endTime := now.Add(24 * time.Hour)
	lockupPeriods := types.Periods{
		types.Period{Length: int64(16 * 60 * 60), Amount: sdk.NewCoins(sdk.NewInt64Coin(feeDenom, 1000), sdk.NewInt64Coin(stakeDenom, 100))},
	}
	vestingPeriods := types.Periods{
		types.Period{Length: int64(12 * 60 * 60), Amount: sdk.Coins{sdk.NewInt64Coin(feeDenom, 500), sdk.NewInt64Coin(stakeDenom, 50)}},
		types.Period{Length: int64(6 * 60 * 60), Amount: sdk.Coins{sdk.NewInt64Coin(feeDenom, 250), sdk.NewInt64Coin(stakeDenom, 25)}},
		types.Period{Length: int64(6 * 60 * 60), Amount: sdk.Coins{sdk.NewInt64Coin(feeDenom, 250), sdk.NewInt64Coin(stakeDenom, 25)}},
	}

	bacc, origCoins := initBaseAccount()

	// require the ability to delegate all vesting coins
	va := types.NewClawbackVestingAccount(bacc, sdk.AccAddress([]byte("funder")), origCoins, now.Unix(), lockupPeriods, vestingPeriods)
	va.TrackDelegation(now, origCoins, origCoins)
	require.Equal(t, origCoins, va.DelegatedVesting)
	require.Nil(t, va.DelegatedFree)

	// require the ability to delegate all vested coins
	va = types.NewClawbackVestingAccount(bacc, sdk.AccAddress([]byte("funder")), origCoins, now.Unix(), lockupPeriods, vestingPeriods)
	va.TrackDelegation(endTime, origCoins, origCoins)
	require.Nil(t, va.DelegatedVesting)
	require.Equal(t, origCoins, va.DelegatedFree)

	// delegate half of vesting coins
	va = types.NewClawbackVestingAccount(bacc, sdk.AccAddress([]byte("funder")), origCoins, now.Unix(), lockupPeriods, vestingPeriods)
	va.TrackDelegation(now, origCoins, vestingPeriods[0].Amount)
	// require that all delegated coins are delegated vesting
	require.Equal(t, va.DelegatedVesting, vestingPeriods[0].Amount)
	require.Nil(t, va.DelegatedFree)

	// delegate 75% of coins, split between vested and vesting
	va = types.NewClawbackVestingAccount(bacc, sdk.AccAddress([]byte("funder")), origCoins, now.Unix(), lockupPeriods, vestingPeriods)
	va.TrackDelegation(now.Add(17*time.Hour), origCoins, vestingPeriods[0].Amount.Add(vestingPeriods[1].Amount...))
	// require that the maximum possible amount of vesting coins are chosen for delegation.
	require.Equal(t, va.DelegatedFree, vestingPeriods[1].Amount)
	require.Equal(t, va.DelegatedVesting, vestingPeriods[0].Amount)

	// require the ability to delegate all vesting coins (50%) and all vested coins (50%)
	va = types.NewClawbackVestingAccount(bacc, sdk.AccAddress([]byte("funder")), origCoins, now.Unix(), lockupPeriods, vestingPeriods)
	va.TrackDelegation(now.Add(17*time.Hour), origCoins, sdk.Coins{sdk.NewInt64Coin(stakeDenom, 50)})
	require.Equal(t, sdk.Coins{sdk.NewInt64Coin(stakeDenom, 50)}, va.DelegatedVesting)
	require.Nil(t, va.DelegatedFree)

	va.TrackDelegation(now.Add(17*time.Hour), origCoins, sdk.Coins{sdk.NewInt64Coin(stakeDenom, 50)})
	require.Equal(t, sdk.Coins{sdk.NewInt64Coin(stakeDenom, 50)}, va.DelegatedVesting)
	require.Equal(t, sdk.Coins{sdk.NewInt64Coin(stakeDenom, 50)}, va.DelegatedFree)

	// require no modifications when delegation amount is zero or not enough funds
	va = types.NewClawbackVestingAccount(bacc, sdk.AccAddress([]byte("funder")), origCoins, now.Unix(), lockupPeriods, vestingPeriods)
	require.Panics(t, func() {
		va.TrackDelegation(endTime, origCoins, sdk.Coins{sdk.NewInt64Coin(stakeDenom, 1000000)})
	})
	require.Nil(t, va.DelegatedVesting)
	require.Nil(t, va.DelegatedFree)
}

func TestTrackUndelegationClawbackVestingAcc(t *testing.T) {
	now := tmtime.Now()
	endTime := now.Add(24 * time.Hour)
	lockupPeriods := types.Periods{
		types.Period{Length: int64(16 * 60 * 60), Amount: sdk.NewCoins(sdk.NewInt64Coin(feeDenom, 1000), sdk.NewInt64Coin(stakeDenom, 100))},
	}
	vestingPeriods := types.Periods{
		types.Period{Length: int64(12 * 60 * 60), Amount: sdk.Coins{sdk.NewInt64Coin(feeDenom, 500), sdk.NewInt64Coin(stakeDenom, 50)}},
		types.Period{Length: int64(6 * 60 * 60), Amount: sdk.Coins{sdk.NewInt64Coin(feeDenom, 250), sdk.NewInt64Coin(stakeDenom, 25)}},
		types.Period{Length: int64(6 * 60 * 60), Amount: sdk.Coins{sdk.NewInt64Coin(feeDenom, 250), sdk.NewInt64Coin(stakeDenom, 25)}},
	}

	bacc, origCoins := initBaseAccount()

	// require the ability to undelegate all vesting coins at the beginning of vesting
	va := types.NewClawbackVestingAccount(bacc, sdk.AccAddress([]byte("funder")), origCoins, now.Unix(), lockupPeriods, vestingPeriods)
	va.TrackDelegation(now, origCoins, origCoins)
	va.TrackUndelegation(origCoins)
	require.Nil(t, va.DelegatedFree)
	require.Nil(t, va.DelegatedVesting)

	// require the ability to undelegate all vested coins at the end of vesting
	va = types.NewClawbackVestingAccount(bacc, sdk.AccAddress([]byte("funder")), origCoins, now.Unix(), lockupPeriods, vestingPeriods)

	va.TrackDelegation(endTime, origCoins, origCoins)
	va.TrackUndelegation(origCoins)
	require.Nil(t, va.DelegatedFree)
	require.Nil(t, va.DelegatedVesting)

	// require the ability to undelegate half of coins
	va = types.NewClawbackVestingAccount(bacc, sdk.AccAddress([]byte("funder")), origCoins, now.Unix(), lockupPeriods, vestingPeriods)
	va.TrackDelegation(endTime, origCoins, vestingPeriods[0].Amount)
	va.TrackUndelegation(vestingPeriods[0].Amount)
	require.Nil(t, va.DelegatedFree)
	require.Nil(t, va.DelegatedVesting)

	// require no modifications when the undelegation amount is zero
	va = types.NewClawbackVestingAccount(bacc, sdk.AccAddress([]byte("funder")), origCoins, now.Unix(), lockupPeriods, vestingPeriods)

	require.Panics(t, func() {
		va.TrackUndelegation(sdk.Coins{sdk.NewInt64Coin(stakeDenom, 0)})
	})
	require.Nil(t, va.DelegatedFree)
	require.Nil(t, va.DelegatedVesting)

	// vest 50% and delegate to two validators
	va = types.NewClawbackVestingAccount(bacc, sdk.AccAddress([]byte("funder")), origCoins, now.Unix(), lockupPeriods, vestingPeriods)
	va.TrackDelegation(now.Add(17*time.Hour), origCoins, sdk.Coins{sdk.NewInt64Coin(stakeDenom, 50)})
	va.TrackDelegation(now.Add(17*time.Hour), origCoins, sdk.Coins{sdk.NewInt64Coin(stakeDenom, 50)})

	// undelegate from one validator that got slashed 50%
	va.TrackUndelegation(sdk.Coins{sdk.NewInt64Coin(stakeDenom, 25)})
	require.Equal(t, sdk.Coins{sdk.NewInt64Coin(stakeDenom, 25)}, va.DelegatedFree)
	require.Equal(t, sdk.Coins{sdk.NewInt64Coin(stakeDenom, 50)}, va.DelegatedVesting)

	// undelegate from the other validator that did not get slashed
	va.TrackUndelegation(sdk.Coins{sdk.NewInt64Coin(stakeDenom, 50)})
	require.Nil(t, va.DelegatedFree)
	require.Equal(t, sdk.Coins{sdk.NewInt64Coin(stakeDenom, 25)}, va.DelegatedVesting)
}

func TestComputeClawback(t *testing.T) {
	c := sdk.NewCoins
	fee := func(x int64) sdk.Coin { return sdk.NewInt64Coin(feeDenom, x) }
	stake := func(x int64) sdk.Coin { return sdk.NewInt64Coin(stakeDenom, x) }
	now := tmtime.Now()
	lockupPeriods := types.Periods{
		{Length: int64(12 * 3600), Amount: c(fee(1000), stake(100))}, // noon
	}
	vestingPeriods := types.Periods{
		{Length: int64(8 * 3600), Amount: c(fee(200))},            // 8am
		{Length: int64(1 * 3600), Amount: c(fee(200), stake(50))}, // 9am
		{Length: int64(6 * 3600), Amount: c(fee(200), stake(50))}, // 3pm
		{Length: int64(2 * 3600), Amount: c(fee(200))},            // 5pm
		{Length: int64(1 * 3600), Amount: c(fee(200))},            // 6pm
	}

	bacc, origCoins := initBaseAccount()
	va := types.NewClawbackVestingAccount(bacc, sdk.AccAddress([]byte("funder")), origCoins, now.Unix(), lockupPeriods, vestingPeriods)

	va2, amt := va.ComputeClawback(now.Unix())
	require.Equal(t, c(fee(1000), stake(100)), amt)
	require.Equal(t, c(), va2.OriginalVesting)
	require.Equal(t, 0, len(va2.LockupPeriods))
	require.Equal(t, 0, len(va2.VestingPeriods))

	va2, amt = va.ComputeClawback(now.Add(11 * time.Hour).Unix())
	require.Equal(t, c(fee(600), stake(50)), amt)
	require.Equal(t, c(fee(400), stake(50)), va2.OriginalVesting)
	require.Equal(t, []types.Period{{Length: int64(12 * 3600), Amount: c(fee(400), stake(50))}}, va2.LockupPeriods)
	require.Equal(t, []types.Period{
		{Length: int64(8 * 3600), Amount: c(fee(200))},            // 8am
		{Length: int64(1 * 3600), Amount: c(fee(200), stake(50))}, // 9am
	}, va2.VestingPeriods)

	va2, amt = va.ComputeClawback(now.Add(23 * time.Hour).Unix())
	require.Equal(t, c(), amt)
	require.Equal(t, *va, va2)
}

func TestClawback(t *testing.T) {
	c := sdk.NewCoins
	fee := func(x int64) sdk.Coin { return sdk.NewInt64Coin(feeDenom, x) }
	stake := func(x int64) sdk.Coin { return sdk.NewInt64Coin(stakeDenom, x) }
	now := tmtime.Now()

	// set up simapp and validators
	app := simapp.Setup(false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{}).WithBlockTime((now))
	valAddr, val := createValidator(t, ctx, app, 100)
	require.Equal(t, "stake", app.StakingKeeper.BondDenom(ctx))

	lockupPeriods := types.Periods{
		{Length: int64(12 * 3600), Amount: c(fee(1000), stake(100))}, // noon
	}
	vestingPeriods := types.Periods{
		{Length: int64(8 * 3600), Amount: c(fee(200))},            // 8am
		{Length: int64(1 * 3600), Amount: c(fee(200), stake(50))}, // 9am
		{Length: int64(6 * 3600), Amount: c(fee(200), stake(50))}, // 3pm
		{Length: int64(2 * 3600), Amount: c(fee(200))},            // 5pm
		{Length: int64(1 * 3600), Amount: c(fee(200))},            // 6pm
	}

	bacc, origCoins := initBaseAccount()
	_, _, funder := testdata.KeyTestPubAddr()
	va := types.NewClawbackVestingAccount(bacc, funder, origCoins, now.Unix(), lockupPeriods, vestingPeriods)
	// simulate 17stake lost to slashing
	va.DelegatedVesting = c(stake(17))
	addr := va.GetAddress()
	app.AccountKeeper.SetAccount(ctx, va)

	// fund the vesting account with 17 take lost to slashing
	err := simapp.FundAccount(app.BankKeeper, ctx, addr, c(fee(1000), stake(83)))
	require.NoError(t, err)
	require.Equal(t, int64(1000), app.BankKeeper.GetBalance(ctx, addr, feeDenom).Amount.Int64())
	require.Equal(t, int64(83), app.BankKeeper.GetBalance(ctx, addr, stakeDenom).Amount.Int64())
	ctx = ctx.WithBlockTime(now.Add(11 * time.Hour))

	// delegate 65
	shares, err := app.StakingKeeper.Delegate(ctx, addr, sdk.NewInt(65), stakingtypes.Unbonded, val, true)
	require.NoError(t, err)
	require.Equal(t, sdk.NewInt(65), shares.TruncateInt())
	require.Equal(t, int64(18), app.BankKeeper.GetBalance(ctx, addr, stakeDenom).Amount.Int64())

	// undelegate 5
	_, err = app.StakingKeeper.Undelegate(ctx, addr, valAddr, sdk.NewDec(5))
	require.NoError(t, err)

	// clawback the unvested funds (600fee, 50stake)
	_, _, dest := testdata.KeyTestPubAddr()
	va2 := app.AccountKeeper.GetAccount(ctx, addr).(*types.ClawbackVestingAccount)
	err = va2.Clawback(ctx, dest, app.AccountKeeper, app.BankKeeper, app.StakingKeeper)
	require.NoError(t, err)

	// check vesting account
	// want 400fee, 33stake (delegated), all vested
	feeAmt := app.BankKeeper.GetBalance(ctx, addr, feeDenom).Amount
	require.Equal(t, int64(400), feeAmt.Int64())
	stakeAmt := app.BankKeeper.GetBalance(ctx, addr, stakeDenom).Amount
	require.Equal(t, int64(0), stakeAmt.Int64())
	del, found := app.StakingKeeper.GetDelegation(ctx, addr, valAddr)
	require.True(t, found)
	shares = del.GetShares()
	val, found = app.StakingKeeper.GetValidator(ctx, valAddr)
	require.True(t, found)
	stakeAmt = val.TokensFromSharesTruncated(shares).RoundInt()
	require.Equal(t, sdk.NewInt(33), stakeAmt)

	// check destination account
	// want 600fee, 50stake (18 unbonded, 5 unboinding, 27 bonded)
	feeAmt = app.BankKeeper.GetBalance(ctx, dest, feeDenom).Amount
	require.Equal(t, int64(600), feeAmt.Int64())
	stakeAmt = app.BankKeeper.GetBalance(ctx, dest, stakeDenom).Amount
	require.Equal(t, int64(18), stakeAmt.Int64())
	del, found = app.StakingKeeper.GetDelegation(ctx, dest, valAddr)
	require.True(t, found)
	shares = del.GetShares()
	stakeAmt = val.TokensFromSharesTruncated(shares).RoundInt()
	require.Equal(t, sdk.NewInt(27), stakeAmt)
	ubd, found := app.StakingKeeper.GetUnbondingDelegation(ctx, dest, valAddr)
	require.True(t, found)
	require.Equal(t, sdk.NewInt(5), ubd.Entries[0].Balance)
}

func TestRewards(t *testing.T) {
	c := sdk.NewCoins
	stake := func(x int64) sdk.Coin { return sdk.NewInt64Coin(stakeDenom, x) }
	now := tmtime.Now()

	// set up simapp and validators
	app := simapp.Setup(false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{}).WithBlockTime((now))
	_, val := createValidator(t, ctx, app, 100)
	require.Equal(t, "stake", app.StakingKeeper.BondDenom(ctx))

	// create vesting account
	lockupPeriods := types.Periods{
		{Length: 1, Amount: c(stake(4000))},
	}
	vestingPeriods := types.Periods{
		{Length: int64(100), Amount: c(stake(500))},
		{Length: int64(100), Amount: c(stake(500))},
		{Length: int64(100), Amount: c(stake(500))},
		{Length: int64(100), Amount: c(stake(500))},
		{Length: int64(100), Amount: c(stake(500))},
		{Length: int64(100), Amount: c(stake(500))},
		{Length: int64(100), Amount: c(stake(500))},
		{Length: int64(100), Amount: c(stake(500))},
	}
	bacc, _ := initBaseAccount()
	origCoins := c(stake(4000))
	_, _, funder := testdata.KeyTestPubAddr()
	va := types.NewClawbackVestingAccount(bacc, funder, origCoins, now.Unix(), lockupPeriods, vestingPeriods)
	addr := va.GetAddress()
	app.AccountKeeper.SetAccount(ctx, va)

	// fund the vesting account with 300 stake lost to transfer
	err := simapp.FundAccount(app.BankKeeper, ctx, addr, c(stake(3700)))
	require.NoError(t, err)
	require.Equal(t, int64(3700), app.BankKeeper.GetBalance(ctx, addr, stakeDenom).Amount.Int64())
	ctx = ctx.WithBlockTime(now.Add(350 * time.Second))

	// delegate 1600
	shares, err := app.StakingKeeper.Delegate(ctx, addr, sdk.NewInt(1600), stakingtypes.Unbonded, val, true)
	require.NoError(t, err)
	require.Equal(t, sdk.NewInt(1600), shares.TruncateInt())
	require.Equal(t, int64(2100), app.BankKeeper.GetBalance(ctx, addr, stakeDenom).Amount.Int64())
	va = app.AccountKeeper.GetAccount(ctx, addr).(*types.ClawbackVestingAccount)

	// distribute a reward of 120stake
	err = simapp.FundAccount(app.BankKeeper, ctx, addr, c(stake(120)))
	require.NoError(t, err)
	va.PostReward(ctx, c(stake(120)), app.AccountKeeper, app.BankKeeper, app.StakingKeeper)
	va = app.AccountKeeper.GetAccount(ctx, addr).(*types.ClawbackVestingAccount)
	require.Equal(t, int64(4030), va.OriginalVesting.AmountOf(stakeDenom).Int64())
	require.Equal(t, 8, len(va.VestingPeriods))
	for i := 0; i < 3; i++ {
		require.Equal(t, int64(500), va.VestingPeriods[i].Amount.AmountOf(stakeDenom).Int64())
	}
	for i := 3; i < 8; i++ {
		require.Equal(t, int64(506), va.VestingPeriods[i].Amount.AmountOf(stakeDenom).Int64())
	}
}

func TestRewards_PostSlash(t *testing.T) {
	c := sdk.NewCoins
	stake := func(x int64) sdk.Coin { return sdk.NewInt64Coin(stakeDenom, x) }
	now := tmtime.Now()

	// set up simapp and validators
	app := simapp.Setup(false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{}).WithBlockTime((now))
	_, val := createValidator(t, ctx, app, 100)
	require.Equal(t, "stake", app.StakingKeeper.BondDenom(ctx))

	// create vesting account with a simulated 350stake lost to slashing
	lockupPeriods := types.Periods{
		{Length: 1, Amount: c(stake(4000))},
	}
	vestingPeriods := types.Periods{
		{Length: int64(100), Amount: c(stake(500))},
		{Length: int64(100), Amount: c(stake(500))},
		{Length: int64(100), Amount: c(stake(500))},
		{Length: int64(100), Amount: c(stake(500))},
		{Length: int64(100), Amount: c(stake(500))},
		{Length: int64(100), Amount: c(stake(500))},
		{Length: int64(100), Amount: c(stake(500))},
		{Length: int64(100), Amount: c(stake(500))},
	}
	bacc, _ := initBaseAccount()
	origCoins := c(stake(4000))
	_, _, funder := testdata.KeyTestPubAddr()
	va := types.NewClawbackVestingAccount(bacc, funder, origCoins, now.Unix(), lockupPeriods, vestingPeriods)
	addr := va.GetAddress()
	va.DelegatedVesting = c(stake(350))
	app.AccountKeeper.SetAccount(ctx, va)

	// fund the vesting account with 350 stake lost to slashing
	err := simapp.FundAccount(app.BankKeeper, ctx, addr, c(stake(3650)))
	require.NoError(t, err)
	require.Equal(t, int64(3650), app.BankKeeper.GetBalance(ctx, addr, stakeDenom).Amount.Int64())

	// delegate all 3650stake
	shares, err := app.StakingKeeper.Delegate(ctx, addr, sdk.NewInt(3650), stakingtypes.Unbonded, val, true)
	require.NoError(t, err)
	require.Equal(t, sdk.NewInt(3650), shares.TruncateInt())
	require.Equal(t, int64(0), app.BankKeeper.GetBalance(ctx, addr, stakeDenom).Amount.Int64())
	va = app.AccountKeeper.GetAccount(ctx, addr).(*types.ClawbackVestingAccount)

	// distribute a reward of 160stake - should all be unvested
	err = simapp.FundAccount(app.BankKeeper, ctx, addr, c(stake(160)))
	require.NoError(t, err)
	va.PostReward(ctx, c(stake(160)), app.AccountKeeper, app.BankKeeper, app.StakingKeeper)
	va = app.AccountKeeper.GetAccount(ctx, addr).(*types.ClawbackVestingAccount)
	require.Equal(t, int64(4160), va.OriginalVesting.AmountOf(stakeDenom).Int64())
	require.Equal(t, 8, len(va.VestingPeriods))
	for i := 0; i < 8; i++ {
		require.Equal(t, int64(520), va.VestingPeriods[i].Amount.AmountOf(stakeDenom).Int64())
	}

	// must not be able to transfer reward until it vests
	_, _, dest := testdata.KeyTestPubAddr()
	err = app.BankKeeper.SendCoins(ctx, addr, dest, c(stake(1)))
	require.Error(t, err)
	ctx = ctx.WithBlockTime(now.Add(600 * time.Second))
	err = app.BankKeeper.SendCoins(ctx, addr, dest, c(stake(160)))
	require.NoError(t, err)

	// distribute another reward once everything has vested
	ctx = ctx.WithBlockTime(now.Add(1200 * time.Second))
	err = simapp.FundAccount(app.BankKeeper, ctx, addr, c(stake(160)))
	require.NoError(t, err)
	va.PostReward(ctx, c(stake(160)), app.AccountKeeper, app.BankKeeper, app.StakingKeeper)
	va = app.AccountKeeper.GetAccount(ctx, addr).(*types.ClawbackVestingAccount)
	// shouldn't be added to vesting schedule
	require.Equal(t, int64(4160), va.OriginalVesting.AmountOf(stakeDenom).Int64())
}

func TestAddGrantClawbackVestingAcc_fullSlash(t *testing.T) {
	c := sdk.NewCoins
	stake := func(amt int64) sdk.Coin { return sdk.NewInt64Coin(stakeDenom, amt) }
	now := tmtime.Now()

	// set up simapp
	app := simapp.Setup(false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{}).WithBlockTime((now))
	require.Equal(t, "stake", app.StakingKeeper.BondDenom(ctx))

	// create an account with an initial grant
	_, _, funder := testdata.KeyTestPubAddr()
	lockupPeriods := types.Periods{{Length: 1, Amount: c(stake(100))}}
	vestingPeriods := types.Periods{
		{Length: 100, Amount: c(stake(40))},
		{Length: 100, Amount: c(stake(60))},
	}
	bacc, _ := initBaseAccount()
	origCoins := c(stake(100))
	va := types.NewClawbackVestingAccount(bacc, funder, origCoins, now.Unix(), lockupPeriods, vestingPeriods)
	addr := va.GetAddress()

	// simulate all 100stake lost to slashing
	va.DelegatedVesting = c(stake(100))

	// Nothing locked at now+150, due to slashing
	require.Equal(t, int64(0), va.LockedCoins(ctx.WithBlockTime(now.Add(150*time.Second))).AmountOf(stakeDenom).Int64())

	// Add a new grant of 50stake
	newGrant := c(stake(50))
	va.AddGrant(ctx, app.StakingKeeper, now.Add(500*time.Second).Unix(),
		[]types.Period{{Length: 1, Amount: newGrant}},
		[]types.Period{{Length: 50, Amount: newGrant}}, newGrant)
	app.AccountKeeper.SetAccount(ctx, va)

	// The new 50stake are locked at now+150
	require.Equal(t, int64(50), va.LockedCoins(ctx.WithBlockTime(now.Add(150*time.Second))).AmountOf(stakeDenom).Int64())

	// fund the vesting account with new grant (old has vested and transferred out)
	ctx = ctx.WithBlockTime(now.Add(500 * time.Second))
	err := simapp.FundAccount(app.BankKeeper, ctx, addr, newGrant)
	require.NoError(t, err)
	require.Equal(t, int64(50), app.BankKeeper.GetBalance(ctx, addr, stakeDenom).Amount.Int64())

	// we should not be able to transfer the new grant out until it has vested
	_, _, dest := testdata.KeyTestPubAddr()
	err = app.BankKeeper.SendCoins(ctx, addr, dest, c(stake(1)))
	require.Error(t, err)
	ctx = ctx.WithBlockTime(now.Add(600 * time.Second))
	err = app.BankKeeper.SendCoins(ctx, addr, dest, newGrant)
	require.NoError(t, err)
}

func TestAddGrantClawbackVestingAcc(t *testing.T) {
	c := sdk.NewCoins
	fee := func(amt int64) sdk.Coin { return sdk.NewInt64Coin(feeDenom, amt) }
	stake := func(amt int64) sdk.Coin { return sdk.NewInt64Coin(stakeDenom, amt) }
	now := tmtime.Now()

	// set up simapp
	app := simapp.Setup(false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{}).WithBlockTime((now))
	require.Equal(t, "stake", app.StakingKeeper.BondDenom(ctx))

	// create an account with an initial grant
	_, _, funder := testdata.KeyTestPubAddr()
	lockupPeriods := types.Periods{{Length: 1, Amount: c(fee(1000), stake(100))}}
	vestingPeriods := types.Periods{
		{Length: 100, Amount: c(fee(650), stake(40))},
		{Length: 100, Amount: c(fee(350), stake(60))},
	}
	bacc, origCoins := initBaseAccount()
	va := types.NewClawbackVestingAccount(bacc, funder, origCoins, now.Unix(), lockupPeriods, vestingPeriods)
	addr := va.GetAddress()

	// simulate 54stake (unvested) lost to slashing
	va.DelegatedVesting = c(stake(54))

	// Only 6stake are locked at now+150, due to slashing
	require.Equal(t, int64(6), va.LockedCoins(ctx.WithBlockTime(now.Add(150*time.Second))).AmountOf(stakeDenom).Int64())

	// Add a new grant of 50stake
	newGrant := c(stake(50))
	va.AddGrant(ctx, app.StakingKeeper, now.Add(500*time.Second).Unix(),
		[]types.Period{{Length: 1, Amount: newGrant}},
		[]types.Period{{Length: 50, Amount: newGrant}}, newGrant)
	app.AccountKeeper.SetAccount(ctx, va)

	// Only 56stake locked at now+150, due to slashing
	require.Equal(t, int64(56), va.LockedCoins(ctx.WithBlockTime(now.Add(150*time.Second))).AmountOf(stakeDenom).Int64())

	// fund the vesting account with new grant (old has vested and transferred out)
	ctx = ctx.WithBlockTime(now.Add(500 * time.Second))
	err := simapp.FundAccount(app.BankKeeper, ctx, addr, newGrant)
	require.NoError(t, err)
	require.Equal(t, int64(50), app.BankKeeper.GetBalance(ctx, addr, stakeDenom).Amount.Int64())

	// we should not be able to transfer the new grant out until it has vested
	_, _, dest := testdata.KeyTestPubAddr()
	err = app.BankKeeper.SendCoins(ctx, addr, dest, c(stake(1)))
	require.Error(t, err)
	ctx = ctx.WithBlockTime(now.Add(600 * time.Second))
	err = app.BankKeeper.SendCoins(ctx, addr, dest, newGrant)
	require.NoError(t, err)
}

func TestGenesisAccountValidate(t *testing.T) {
	pubkey := secp256k1.GenPrivKey().PubKey()
	addr := sdk.AccAddress(pubkey.Address())
	baseAcc := authtypes.NewBaseAccount(addr, pubkey, 0, 0)
	initialVesting := sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 50))
	baseVestingWithCoins := types.NewBaseVestingAccount(baseAcc, initialVesting, 100)
	tests := []struct {
		name   string
		acc    authtypes.GenesisAccount
		expErr bool
	}{
		{
			"valid base account",
			baseAcc,
			false,
		},
		{
			"invalid base valid account",
			authtypes.NewBaseAccount(addr, secp256k1.GenPrivKey().PubKey(), 0, 0),
			true,
		},
		{
			"valid base vesting account",
			baseVestingWithCoins,
			false,
		},
		{
			"valid continuous vesting account",
			types.NewContinuousVestingAccount(baseAcc, initialVesting, 100, 200),
			false,
		},
		{
			"invalid vesting times",
			types.NewContinuousVestingAccount(baseAcc, initialVesting, 1654668078, 1554668078),
			true,
		},
		{
			"valid periodic vesting account",
			types.NewPeriodicVestingAccount(baseAcc, initialVesting, 0, types.Periods{types.Period{Length: int64(100), Amount: sdk.Coins{sdk.NewInt64Coin(sdk.DefaultBondDenom, 50)}}}),
			false,
		},
		{
			"invalid vesting period lengths",
			types.NewPeriodicVestingAccountRaw(
				baseVestingWithCoins,
				0, types.Periods{types.Period{Length: int64(50), Amount: sdk.Coins{sdk.NewInt64Coin(sdk.DefaultBondDenom, 50)}}}),
			true,
		},
		{
			"invalid vesting period amounts",
			types.NewPeriodicVestingAccountRaw(
				baseVestingWithCoins,
				0, types.Periods{types.Period{Length: int64(100), Amount: sdk.Coins{sdk.NewInt64Coin(sdk.DefaultBondDenom, 25)}}}),
			true,
		},
		{
			"valid permanent locked vesting account",
			types.NewPermanentLockedAccount(baseAcc, initialVesting),
			false,
		},
		{
			"invalid positive end time for permanently locked vest account",
			&types.PermanentLockedAccount{BaseVestingAccount: baseVestingWithCoins},
			true,
		},
		{
			"valid clawback vesting account",
			types.NewClawbackVestingAccount(baseAcc, sdk.AccAddress([]byte("the funder")), initialVesting, 0,
				types.Periods{types.Period{Length: 101, Amount: initialVesting}},
				types.Periods{types.Period{Length: 201, Amount: initialVesting}}),
			false,
		},
		{
			"invalid clawback vesting end",
			&types.ClawbackVestingAccount{
				BaseVestingAccount: &types.BaseVestingAccount{
					BaseAccount:     baseAcc,
					OriginalVesting: initialVesting,
					EndTime:         50,
				},
				FunderAddress:  "funder",
				StartTime:      100,
				LockupPeriods:  types.Periods{types.Period{Length: 10, Amount: initialVesting}},
				VestingPeriods: types.Periods{types.Period{Length: 10, Amount: initialVesting}},
			},
			true,
		},
		{
			"invalid clawback long lockup",
			&types.ClawbackVestingAccount{
				BaseVestingAccount: &types.BaseVestingAccount{
					BaseAccount:     baseAcc,
					OriginalVesting: initialVesting,
					EndTime:         60,
				},
				FunderAddress:  "funder",
				StartTime:      50,
				LockupPeriods:  types.Periods{types.Period{Length: 20, Amount: initialVesting}},
				VestingPeriods: types.Periods{types.Period{Length: 10, Amount: initialVesting}},
			},
			true,
		},
		{
			"invalid clawback lockup coins",
			&types.ClawbackVestingAccount{
				BaseVestingAccount: &types.BaseVestingAccount{
					BaseAccount:     baseAcc,
					OriginalVesting: initialVesting,
					EndTime:         120,
				},
				FunderAddress:  "funder",
				StartTime:      100,
				LockupPeriods:  types.Periods{types.Period{Length: 10, Amount: initialVesting.Add(initialVesting...)}},
				VestingPeriods: types.Periods{types.Period{Length: 10, Amount: initialVesting}},
			},
			true,
		},
		{
			"invalid clawback long vesting",
			&types.ClawbackVestingAccount{
				BaseVestingAccount: &types.BaseVestingAccount{
					BaseAccount:     baseAcc,
					OriginalVesting: initialVesting,
					EndTime:         110,
				},
				FunderAddress:  "funder",
				StartTime:      100,
				LockupPeriods:  types.Periods{types.Period{Length: 10, Amount: initialVesting}},
				VestingPeriods: types.Periods{types.Period{Length: 20, Amount: initialVesting}},
			},
			true,
		},
		{
			"invalid clawback vesting coins",
			&types.ClawbackVestingAccount{
				BaseVestingAccount: &types.BaseVestingAccount{
					BaseAccount:     baseAcc,
					OriginalVesting: initialVesting,
					EndTime:         120,
				},
				FunderAddress:  "funder",
				StartTime:      100,
				LockupPeriods:  types.Periods{types.Period{Length: 10, Amount: initialVesting}},
				VestingPeriods: types.Periods{types.Period{Length: 10, Amount: initialVesting.Add(initialVesting...)}},
			},
			true,
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.expErr, tt.acc.Validate() != nil)
		})
	}
}

func TestContinuousVestingAccountMarshal(t *testing.T) {
	baseAcc, coins := initBaseAccount()
	baseVesting := types.NewBaseVestingAccount(baseAcc, coins, time.Now().Unix())
	acc := types.NewContinuousVestingAccountRaw(baseVesting, baseVesting.EndTime)

	app := simapp.Setup(false)

	bz, err := app.AccountKeeper.MarshalAccount(acc)
	require.NoError(t, err)

	acc2, err := app.AccountKeeper.UnmarshalAccount(bz)
	require.NoError(t, err)
	require.IsType(t, &types.ContinuousVestingAccount{}, acc2)
	require.Equal(t, acc.String(), acc2.String())

	// error on bad bytes
	_, err = app.AccountKeeper.UnmarshalAccount(bz[:len(bz)/2])
	require.Error(t, err)
}

func TestPeriodicVestingAccountMarshal(t *testing.T) {
	baseAcc, coins := initBaseAccount()
	acc := types.NewPeriodicVestingAccount(baseAcc, coins, time.Now().Unix(), types.Periods{types.Period{3600, coins}})

	app := simapp.Setup(false)

	bz, err := app.AccountKeeper.MarshalAccount(acc)
	require.NoError(t, err)

	acc2, err := app.AccountKeeper.UnmarshalAccount(bz)
	require.NoError(t, err)
	require.IsType(t, &types.PeriodicVestingAccount{}, acc2)
	require.Equal(t, acc.String(), acc2.String())

	// error on bad bytes
	_, err = app.AccountKeeper.UnmarshalAccount(bz[:len(bz)/2])
	require.Error(t, err)
}

func TestDelayedVestingAccountMarshal(t *testing.T) {
	baseAcc, coins := initBaseAccount()
	acc := types.NewDelayedVestingAccount(baseAcc, coins, time.Now().Unix())

	app := simapp.Setup(false)

	bz, err := app.AccountKeeper.MarshalAccount(acc)
	require.NoError(t, err)

	acc2, err := app.AccountKeeper.UnmarshalAccount(bz)
	require.NoError(t, err)
	require.IsType(t, &types.DelayedVestingAccount{}, acc2)
	require.Equal(t, acc.String(), acc2.String())

	// error on bad bytes
	_, err = app.AccountKeeper.UnmarshalAccount(bz[:len(bz)/2])
	require.Error(t, err)
}

func TestPermanentLockedAccountMarshal(t *testing.T) {
	baseAcc, coins := initBaseAccount()
	acc := types.NewPermanentLockedAccount(baseAcc, coins)

	app := simapp.Setup(false)

	bz, err := app.AccountKeeper.MarshalAccount(acc)
	require.NoError(t, err)

	acc2, err := app.AccountKeeper.UnmarshalAccount(bz)
	require.NoError(t, err)
	require.IsType(t, &types.PermanentLockedAccount{}, acc2)
	require.Equal(t, acc.String(), acc2.String())

	// error on bad bytes
	_, err = app.AccountKeeper.UnmarshalAccount(bz[:len(bz)/2])
	require.Error(t, err)
}

func TestClawbackVestingAccountMarshal(t *testing.T) {
	baseAcc, coins := initBaseAccount()
	addr := sdk.AccAddress([]byte("the funder"))
	acc := types.NewClawbackVestingAccount(baseAcc, addr, coins, time.Now().Unix(),
		types.Periods{types.Period{3600, coins}}, types.Periods{types.Period{3600, coins}})

	app := simapp.Setup(false)

	bz, err := app.AccountKeeper.MarshalAccount(acc)
	require.NoError(t, err)

	acc2, err := app.AccountKeeper.UnmarshalAccount(bz)
	require.NoError(t, err)
	require.IsType(t, &types.ClawbackVestingAccount{}, acc2)
	require.Equal(t, acc.String(), acc2.String())

	// error on bad bytes
	_, err = app.AccountKeeper.UnmarshalAccount(bz[:len(bz)/2])
	require.Error(t, err)
}

func TestClawbackVestingAccountStore(t *testing.T) {
	baseAcc, coins := initBaseAccount()
	addr := sdk.AccAddress([]byte("the funder"))
	acc := types.NewClawbackVestingAccount(baseAcc, addr, coins, time.Now().Unix(),
		types.Periods{types.Period{3600, coins}}, types.Periods{types.Period{3600, coins}})

	app := simapp.Setup(false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})
	createValidator(t, ctx, app, 100)

	app.AccountKeeper.SetAccount(ctx, acc)
	acc2 := app.AccountKeeper.GetAccount(ctx, acc.GetAddress())
	require.IsType(t, &types.ClawbackVestingAccount{}, acc2)
	require.Equal(t, acc.String(), acc2.String())
}

func initBaseAccount() (*authtypes.BaseAccount, sdk.Coins) {
	_, _, addr := testdata.KeyTestPubAddr()
	origCoins := sdk.Coins{sdk.NewInt64Coin(feeDenom, 1000), sdk.NewInt64Coin(stakeDenom, 100)}
	bacc := authtypes.NewBaseAccountWithAddress(addr)

	return bacc, origCoins
}
