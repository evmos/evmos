package types_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	tmtime "github.com/tendermint/tendermint/types/time"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	sdkvesting "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"

	"github.com/tharsis/evmos/x/vesting/types"
)

var (
	stakeDenom = "stake"
	feeDenom   = "fee"
)

func TestGetVestedCoinsClawbackVestingAcc(t *testing.T) {
	now := tmtime.Now()
	endTime := now.Add(24 * time.Hour)
	lockupPeriods := sdkvesting.Periods{
		sdkvesting.Period{Length: int64(16 * 60 * 60), Amount: sdk.NewCoins(sdk.NewInt64Coin(feeDenom, 1000), sdk.NewInt64Coin(stakeDenom, 100))},
	}
	vestingPeriods := sdkvesting.Periods{
		sdkvesting.Period{Length: int64(12 * 60 * 60), Amount: sdk.Coins{sdk.NewInt64Coin(feeDenom, 500), sdk.NewInt64Coin(stakeDenom, 50)}},
		sdkvesting.Period{Length: int64(6 * 60 * 60), Amount: sdk.Coins{sdk.NewInt64Coin(feeDenom, 250), sdk.NewInt64Coin(stakeDenom, 25)}},
		sdkvesting.Period{Length: int64(6 * 60 * 60), Amount: sdk.Coins{sdk.NewInt64Coin(feeDenom, 250), sdk.NewInt64Coin(stakeDenom, 25)}},
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
	lockupPeriods := sdkvesting.Periods{
		sdkvesting.Period{Length: int64(16 * 60 * 60), Amount: sdk.NewCoins(sdk.NewInt64Coin(feeDenom, 1000), sdk.NewInt64Coin(stakeDenom, 100))},
	}
	vestingPeriods := sdkvesting.Periods{
		sdkvesting.Period{Length: int64(12 * 60 * 60), Amount: sdk.Coins{sdk.NewInt64Coin(feeDenom, 500), sdk.NewInt64Coin(stakeDenom, 50)}},
		sdkvesting.Period{Length: int64(6 * 60 * 60), Amount: sdk.Coins{sdk.NewInt64Coin(feeDenom, 250), sdk.NewInt64Coin(stakeDenom, 25)}},
		sdkvesting.Period{Length: int64(6 * 60 * 60), Amount: sdk.Coins{sdk.NewInt64Coin(feeDenom, 250), sdk.NewInt64Coin(stakeDenom, 25)}},
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
	lockupPeriods := sdkvesting.Periods{
		sdkvesting.Period{Length: int64(16 * 60 * 60), Amount: sdk.NewCoins(sdk.NewInt64Coin(feeDenom, 1000), sdk.NewInt64Coin(stakeDenom, 100))},
	}
	vestingPeriods := sdkvesting.Periods{
		sdkvesting.Period{Length: int64(12 * 60 * 60), Amount: sdk.Coins{sdk.NewInt64Coin(feeDenom, 500), sdk.NewInt64Coin(stakeDenom, 50)}},
		sdkvesting.Period{Length: int64(6 * 60 * 60), Amount: sdk.Coins{sdk.NewInt64Coin(feeDenom, 250), sdk.NewInt64Coin(stakeDenom, 25)}},
		sdkvesting.Period{Length: int64(6 * 60 * 60), Amount: sdk.Coins{sdk.NewInt64Coin(feeDenom, 250), sdk.NewInt64Coin(stakeDenom, 25)}},
	}

	bacc, origCoins := initBaseAccount()
	va := types.NewClawbackVestingAccount(bacc, sdk.AccAddress([]byte("funder")), origCoins, now.Unix(), lockupPeriods, vestingPeriods)

	// require that there exist no spendable coins at the beginning of the
	// vesting schedule
	lockedCoins := va.LockedCoins(now)
	require.Equal(t, origCoins, lockedCoins)

	// require that all original coins are spendable at the end of the vesting
	// schedule
	lockedCoins = va.LockedCoins(endTime)
	require.Equal(t, sdk.NewCoins(), lockedCoins)

	// require that all still vesting coins (50%) are locked
	lockedCoins = va.LockedCoins(now.Add(17 * time.Hour))
	require.Equal(t, sdk.Coins{sdk.NewInt64Coin(feeDenom, 500), sdk.NewInt64Coin(stakeDenom, 50)}, lockedCoins)
}

func TestTrackDelegationClawbackVestingAcc(t *testing.T) {
	now := tmtime.Now()
	endTime := now.Add(24 * time.Hour)
	lockupPeriods := sdkvesting.Periods{
		sdkvesting.Period{Length: int64(16 * 60 * 60), Amount: sdk.NewCoins(sdk.NewInt64Coin(feeDenom, 1000), sdk.NewInt64Coin(stakeDenom, 100))},
	}
	vestingPeriods := sdkvesting.Periods{
		sdkvesting.Period{Length: int64(12 * 60 * 60), Amount: sdk.Coins{sdk.NewInt64Coin(feeDenom, 500), sdk.NewInt64Coin(stakeDenom, 50)}},
		sdkvesting.Period{Length: int64(6 * 60 * 60), Amount: sdk.Coins{sdk.NewInt64Coin(feeDenom, 250), sdk.NewInt64Coin(stakeDenom, 25)}},
		sdkvesting.Period{Length: int64(6 * 60 * 60), Amount: sdk.Coins{sdk.NewInt64Coin(feeDenom, 250), sdk.NewInt64Coin(stakeDenom, 25)}},
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
	lockupPeriods := sdkvesting.Periods{
		sdkvesting.Period{Length: int64(16 * 60 * 60), Amount: sdk.NewCoins(sdk.NewInt64Coin(feeDenom, 1000), sdk.NewInt64Coin(stakeDenom, 100))},
	}
	vestingPeriods := sdkvesting.Periods{
		sdkvesting.Period{Length: int64(12 * 60 * 60), Amount: sdk.Coins{sdk.NewInt64Coin(feeDenom, 500), sdk.NewInt64Coin(stakeDenom, 50)}},
		sdkvesting.Period{Length: int64(6 * 60 * 60), Amount: sdk.Coins{sdk.NewInt64Coin(feeDenom, 250), sdk.NewInt64Coin(stakeDenom, 25)}},
		sdkvesting.Period{Length: int64(6 * 60 * 60), Amount: sdk.Coins{sdk.NewInt64Coin(feeDenom, 250), sdk.NewInt64Coin(stakeDenom, 25)}},
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
	lockupPeriods := sdkvesting.Periods{
		{Length: int64(12 * 3600), Amount: c(fee(1000), stake(100))}, // noon
	}
	vestingPeriods := sdkvesting.Periods{
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
	require.Equal(t, []sdkvesting.Period{{Length: int64(12 * 3600), Amount: c(fee(400), stake(50))}}, va2.LockupPeriods)
	require.Equal(t, []sdkvesting.Period{
		{Length: int64(8 * 3600), Amount: c(fee(200))},            // 8am
		{Length: int64(1 * 3600), Amount: c(fee(200), stake(50))}, // 9am
	}, va2.VestingPeriods)

	va2, amt = va.ComputeClawback(now.Add(23 * time.Hour).Unix())
	require.Equal(t, c(), amt)
	require.Equal(t, *va, va2)
}

func TestGenesisAccountValidate(t *testing.T) {
	pubkey := secp256k1.GenPrivKey().PubKey()
	addr := sdk.AccAddress(pubkey.Address())
	baseAcc := authtypes.NewBaseAccount(addr, pubkey, 0, 0)
	initialVesting := sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 50))
	baseVestingWithCoins := sdkvesting.NewBaseVestingAccount(baseAcc, initialVesting, 100)
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
			"valid clawback vesting account",
			types.NewClawbackVestingAccount(baseAcc, sdk.AccAddress([]byte("the funder")), initialVesting, 0,
				sdkvesting.Periods{sdkvesting.Period{Length: 101, Amount: initialVesting}},
				sdkvesting.Periods{sdkvesting.Period{Length: 201, Amount: initialVesting}}),
			false,
		},
		{
			"invalid clawback vesting end",
			&types.ClawbackVestingAccount{
				BaseVestingAccount: &sdkvesting.BaseVestingAccount{
					BaseAccount:     baseAcc,
					OriginalVesting: initialVesting,
					EndTime:         50,
				},
				FunderAddress:  "funder",
				StartTime:      100,
				LockupPeriods:  sdkvesting.Periods{sdkvesting.Period{Length: 10, Amount: initialVesting}},
				VestingPeriods: sdkvesting.Periods{sdkvesting.Period{Length: 10, Amount: initialVesting}},
			},
			true,
		},
		{
			"invalid clawback long lockup",
			&types.ClawbackVestingAccount{
				BaseVestingAccount: &sdkvesting.BaseVestingAccount{
					BaseAccount:     baseAcc,
					OriginalVesting: initialVesting,
					EndTime:         60,
				},
				FunderAddress:  "funder",
				StartTime:      50,
				LockupPeriods:  sdkvesting.Periods{sdkvesting.Period{Length: 20, Amount: initialVesting}},
				VestingPeriods: sdkvesting.Periods{sdkvesting.Period{Length: 10, Amount: initialVesting}},
			},
			true,
		},
		{
			"invalid clawback lockup coins",
			&types.ClawbackVestingAccount{
				BaseVestingAccount: &sdkvesting.BaseVestingAccount{
					BaseAccount:     baseAcc,
					OriginalVesting: initialVesting,
					EndTime:         120,
				},
				FunderAddress:  "funder",
				StartTime:      100,
				LockupPeriods:  sdkvesting.Periods{sdkvesting.Period{Length: 10, Amount: initialVesting.Add(initialVesting...)}},
				VestingPeriods: sdkvesting.Periods{sdkvesting.Period{Length: 10, Amount: initialVesting}},
			},
			true,
		},
		{
			"invalid clawback long vesting",
			&types.ClawbackVestingAccount{
				BaseVestingAccount: &sdkvesting.BaseVestingAccount{
					BaseAccount:     baseAcc,
					OriginalVesting: initialVesting,
					EndTime:         110,
				},
				FunderAddress:  "funder",
				StartTime:      100,
				LockupPeriods:  sdkvesting.Periods{sdkvesting.Period{Length: 10, Amount: initialVesting}},
				VestingPeriods: sdkvesting.Periods{sdkvesting.Period{Length: 20, Amount: initialVesting}},
			},
			true,
		},
		{
			"invalid clawback vesting coins",
			&types.ClawbackVestingAccount{
				BaseVestingAccount: &sdkvesting.BaseVestingAccount{
					BaseAccount:     baseAcc,
					OriginalVesting: initialVesting,
					EndTime:         120,
				},
				FunderAddress:  "funder",
				StartTime:      100,
				LockupPeriods:  sdkvesting.Periods{sdkvesting.Period{Length: 10, Amount: initialVesting}},
				VestingPeriods: sdkvesting.Periods{sdkvesting.Period{Length: 10, Amount: initialVesting.Add(initialVesting...)}},
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

// TODO fix register interface error
// func TestClawbackVestingAccountMarshal(t *testing.T) {
// 	baseAcc, coins := initBaseAccount()
// 	addr := sdk.AccAddress([]byte("the funder"))
// 	acc := types.NewClawbackVestingAccount(baseAcc, addr, coins, time.Now().Unix(),
// 		sdkvesting.Periods{sdkvesting.Period{Length: 3600, Amount: coins}}, sdkvesting.Periods{sdkvesting.Period{Length: 3600, Amount: coins}})

// 	app := simapp.Setup(false)

// 	bz, err := app.AccountKeeper.MarshalAccount(acc)
// 	require.NoError(t, err)

// 	acc2, err := app.AccountKeeper.UnmarshalAccount(bz)
// 	require.NoError(t, err)
// 	require.IsType(t, &types.ClawbackVestingAccount{}, acc2)
// 	require.Equal(t, acc.String(), acc2.String())

// 	// error on bad bytes
// 	_, err = app.AccountKeeper.UnmarshalAccount(bz[:len(bz)/2])
// 	require.Error(t, err)
// }

// func TestClawbackVestingAccountStore(t *testing.T) {
// 	baseAcc, coins := initBaseAccount()
// 	addr := sdk.AccAddress([]byte("the funder"))
// 	acc := types.NewClawbackVestingAccount(baseAcc, addr, coins, time.Now().Unix(),
// 		sdkvesting.Periods{sdkvesting.Period{Length: 3600, Amount: coins}}, sdkvesting.Periods{sdkvesting.Period{Length: 3600, Amount: coins}})

// 	app := simapp.Setup(false)
// 	ctx := app.BaseApp.NewContext(false, tmproto.Header{})
// 	CreateValidator(t, ctx, app, 100)

// 	app.AccountKeeper.SetAccount(ctx, acc)
// 	acc2 := app.AccountKeeper.GetAccount(ctx, acc.GetAddress())
// 	require.IsType(t, &types.ClawbackVestingAccount{}, acc2)
// 	require.Equal(t, acc.String(), acc2.String())
// }

func initBaseAccount() (*authtypes.BaseAccount, sdk.Coins) {
	_, _, addr := testdata.KeyTestPubAddr()
	origCoins := sdk.Coins{sdk.NewInt64Coin(feeDenom, 1000), sdk.NewInt64Coin(stakeDenom, 100)}
	bacc := authtypes.NewBaseAccountWithAddress(addr)

	return bacc, origCoins
}
