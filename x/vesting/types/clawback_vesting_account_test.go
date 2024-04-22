package types_test

import (
	"testing"
	"time"

	tmtime "github.com/cometbft/cometbft/types/time"
	"github.com/stretchr/testify/suite"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	sdkvesting "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"

	utiltx "github.com/evmos/evmos/v18/testutil/tx"
	"github.com/evmos/evmos/v18/x/vesting/types"
)

var (
	stakeDenom    = "stake"
	feeDenom      = "fee"
	lockupPeriods = sdkvesting.Periods{
		sdkvesting.Period{
			Length: int64(16 * 60 * 60), // 16hs
			Amount: sdk.NewCoins(sdk.NewInt64Coin(feeDenom, 1000), sdk.NewInt64Coin(stakeDenom, 100)),
		},
	}
	vestingPeriods = sdkvesting.Periods{
		sdkvesting.Period{
			Length: int64(12 * 60 * 60), // 12hs
			Amount: getPercentOfVestingCoins(50),
		},
		sdkvesting.Period{
			Length: int64(6 * 60 * 60), // 6hs
			Amount: getPercentOfVestingCoins(25),
		},
		sdkvesting.Period{
			Length: int64(6 * 60 * 60), // 6hs
			Amount: getPercentOfVestingCoins(25),
		},
	}
	origCoins = sdk.Coins{sdk.NewInt64Coin(feeDenom, 1000), sdk.NewInt64Coin(stakeDenom, 100)}
)

type VestingAccountTestSuite struct {
	suite.Suite
}

func TestVestingAccountSuite(t *testing.T) {
	suite.Run(t, new(VestingAccountTestSuite))
}

func (suite *VestingAccountTestSuite) TestClawbackAccountNew() {
	addr := sdk.AccAddress("test_address")
	baseAcc := authtypes.NewBaseAccountWithAddress(addr)
	initialVesting := sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 50))

	testCases := []struct {
		name      string
		acc       authtypes.GenesisAccount
		expErr    bool
		expErrMsg string
	}{
		{
			name: "Clawback vesting account - pass",
			acc: types.NewClawbackVestingAccount(
				baseAcc,
				sdk.AccAddress("the funder"),
				initialVesting,
				time.Now(),
				sdkvesting.Periods{sdkvesting.Period{Length: 101, Amount: initialVesting}},
				sdkvesting.Periods{sdkvesting.Period{Length: 201, Amount: initialVesting}},
			),
			expErr: false,
		},
		{
			name: "Clawback vesting account - invalid vesting end",
			acc: &types.ClawbackVestingAccount{
				BaseVestingAccount: &sdkvesting.BaseVestingAccount{
					BaseAccount:     baseAcc,
					OriginalVesting: initialVesting,
					EndTime:         50,
				},
				FunderAddress:  "funder",
				StartTime:      time.Unix(100, 0),
				LockupPeriods:  sdkvesting.Periods{sdkvesting.Period{Length: 10, Amount: initialVesting}},
				VestingPeriods: sdkvesting.Periods{sdkvesting.Period{Length: 10, Amount: initialVesting}},
			},
			expErr:    true,
			expErrMsg: "vesting start-time must be before end-time",
		},
		{
			name: "Clawback vesting account - lockup too long",
			acc: &types.ClawbackVestingAccount{
				BaseVestingAccount: &sdkvesting.BaseVestingAccount{
					BaseAccount:     baseAcc,
					OriginalVesting: initialVesting,
					EndTime:         60,
				},
				FunderAddress:  "funder",
				StartTime:      time.Unix(50, 0),
				LockupPeriods:  sdkvesting.Periods{sdkvesting.Period{Length: 20, Amount: initialVesting}},
				VestingPeriods: sdkvesting.Periods{sdkvesting.Period{Length: 10, Amount: initialVesting}},
			},
			expErr:    true,
			expErrMsg: "lockup schedule extends beyond account end time",
		},
		{
			name: "Clawback vesting account - invalid lockup coins",
			acc: &types.ClawbackVestingAccount{
				BaseVestingAccount: &sdkvesting.BaseVestingAccount{
					BaseAccount:     baseAcc,
					OriginalVesting: initialVesting,
					EndTime:         120,
				},
				FunderAddress:  "funder",
				StartTime:      time.Unix(100, 0),
				LockupPeriods:  sdkvesting.Periods{sdkvesting.Period{Length: 10, Amount: initialVesting.Add(initialVesting...)}},
				VestingPeriods: sdkvesting.Periods{sdkvesting.Period{Length: 10, Amount: initialVesting}},
			},
			expErr:    true,
			expErrMsg: "original vesting coins does not match the sum of all coins in lockup periods",
		},
		{
			name: "Clawback vesting account - vesting too long",
			acc: &types.ClawbackVestingAccount{
				BaseVestingAccount: &sdkvesting.BaseVestingAccount{
					BaseAccount:     baseAcc,
					OriginalVesting: initialVesting,
					EndTime:         110,
				},
				FunderAddress:  "funder",
				StartTime:      time.Unix(100, 0),
				LockupPeriods:  sdkvesting.Periods{sdkvesting.Period{Length: 10, Amount: initialVesting}},
				VestingPeriods: sdkvesting.Periods{sdkvesting.Period{Length: 20, Amount: initialVesting}},
			},
			expErr:    true,
			expErrMsg: "vesting schedule extends beyond account end time",
		},
		{
			name: "Clawback vesting account - invalid vesting coins",
			acc: &types.ClawbackVestingAccount{
				BaseVestingAccount: &sdkvesting.BaseVestingAccount{
					BaseAccount:     baseAcc,
					OriginalVesting: initialVesting,
					EndTime:         120,
				},
				FunderAddress:  "funder",
				StartTime:      time.Unix(100, 0),
				LockupPeriods:  sdkvesting.Periods{sdkvesting.Period{Length: 10, Amount: initialVesting}},
				VestingPeriods: sdkvesting.Periods{sdkvesting.Period{Length: 10, Amount: initialVesting.Add(initialVesting...)}},
			},
			expErr:    true,
			expErrMsg: "original vesting coins does not match the sum of all coins in vesting periods",
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			err := tc.acc.Validate()
			if tc.expErr {
				suite.Require().Error(err)
				suite.Require().Contains(err.Error(), tc.expErrMsg)
				return
			}
			suite.Require().NoError(err)
		})
	}
}

func (suite *VestingAccountTestSuite) TestGetCoinsFunctions() {
	now := tmtime.Now()
	endTime := now.Add(24 * time.Hour)
	addr := sdk.AccAddress(utiltx.GenerateAddress().Bytes())
	bacc := authtypes.NewBaseAccountWithAddress(addr)
	va := types.NewClawbackVestingAccount(bacc, sdk.AccAddress([]byte("funder")), origCoins, now, lockupPeriods, vestingPeriods)

	testCases := []struct {
		name                   string
		time                   time.Time
		expVestedCoins         sdk.Coins
		expLockedUpVestedCoins sdk.Coins
		expUnlockedVestedCoins sdk.Coins
		expUnvestedCoins       sdk.Coins
		expLockedUpCoins       sdk.Coins
		expUnlockedCoins       sdk.Coins
		expNotSpendable        sdk.Coins
	}{
		{
			name:                   "no coins vested at the beginning of the vesting schedule, all locked",
			time:                   now,
			expVestedCoins:         sdk.Coins{},
			expLockedUpVestedCoins: sdk.Coins{},
			expUnlockedVestedCoins: sdk.Coins{},
			expUnvestedCoins:       origCoins,
			expLockedUpCoins:       origCoins,
			expUnlockedCoins:       sdk.Coins{},
			expNotSpendable:        origCoins,
		},
		{
			name:                   "all coins vested and unlocked at the end of the vesting schedule",
			time:                   endTime,
			expVestedCoins:         origCoins,
			expLockedUpVestedCoins: sdk.Coins{},
			expUnlockedVestedCoins: origCoins,
			expUnvestedCoins:       sdk.Coins{},
			expLockedUpCoins:       sdk.Coins{},
			expUnlockedCoins:       origCoins,
			expNotSpendable:        sdk.Coins{},
		},
		{
			name:                   "no coins vested during first vesting period, all still locked",
			time:                   now.Add(6 * time.Hour),
			expVestedCoins:         sdk.Coins{},
			expLockedUpVestedCoins: sdk.Coins{},
			expUnlockedVestedCoins: sdk.Coins{},
			expUnvestedCoins:       origCoins,
			expLockedUpCoins:       origCoins,
			expUnlockedCoins:       sdk.Coins{},
			expNotSpendable:        origCoins,
		},
		{
			name:                   "50 percent of coins are vested after 1st vesting period, but before unlocking (all locked coins)",
			time:                   now.Add(12 * time.Hour),
			expVestedCoins:         getPercentOfVestingCoins(50),
			expLockedUpVestedCoins: getPercentOfVestingCoins(50),
			expUnlockedVestedCoins: sdk.Coins{},
			expUnvestedCoins:       getPercentOfVestingCoins(50),
			expLockedUpCoins:       origCoins,
			expUnlockedCoins:       sdk.Coins{},
			expNotSpendable:        origCoins,
		},
		{
			name:                   "after lockup period (all coins unlocked) - 50 percent of coins already vested",
			time:                   now.Add(16 * time.Hour),
			expVestedCoins:         getPercentOfVestingCoins(50),
			expLockedUpVestedCoins: sdk.Coins{},
			expUnlockedVestedCoins: getPercentOfVestingCoins(50),
			expUnvestedCoins:       getPercentOfVestingCoins(50),
			expLockedUpCoins:       sdk.Coins{},
			expUnlockedCoins:       origCoins,
			expNotSpendable:        getPercentOfVestingCoins(50),
		},
		{
			name:                   "in between vesting periods 1 and 2 - no new coins don't vested",
			time:                   now.Add(17 * time.Hour),
			expVestedCoins:         getPercentOfVestingCoins(50),
			expLockedUpVestedCoins: sdk.Coins{},
			expUnlockedVestedCoins: getPercentOfVestingCoins(50),
			expUnvestedCoins:       getPercentOfVestingCoins(50),
			expLockedUpCoins:       sdk.Coins{},
			expUnlockedCoins:       origCoins,
			expNotSpendable:        getPercentOfVestingCoins(50),
		},
		{
			name:                   "75 percent of coins vested after period 2",
			time:                   now.Add(18 * time.Hour),
			expVestedCoins:         getPercentOfVestingCoins(75),
			expLockedUpVestedCoins: sdk.Coins{},
			expUnlockedVestedCoins: getPercentOfVestingCoins(75),
			expUnvestedCoins:       getPercentOfVestingCoins(25),
			expLockedUpCoins:       sdk.Coins{},
			expUnlockedCoins:       origCoins,
			expNotSpendable:        getPercentOfVestingCoins(25),
		},
		{
			name:                   "100 percent of coins vested",
			time:                   now.Add(48 * time.Hour),
			expVestedCoins:         origCoins,
			expLockedUpVestedCoins: sdk.Coins{},
			expUnlockedVestedCoins: origCoins,
			expUnvestedCoins:       sdk.Coins{},
			expLockedUpCoins:       sdk.Coins{},
			expUnlockedCoins:       origCoins,
			expNotSpendable:        sdk.Coins{},
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			vestedCoins := va.GetVestedCoins(tc.time)
			suite.Require().Equal(tc.expVestedCoins, vestedCoins)
			lockedUpVested := va.GetLockedUpVestedCoins(tc.time)
			suite.Require().Equal(tc.expLockedUpVestedCoins, lockedUpVested)
			unlockedVestedCoins := va.GetUnlockedVestedCoins(tc.time)
			suite.Require().Equal(tc.expUnlockedVestedCoins, unlockedVestedCoins)
			unvestedCoins := va.GetVestingCoins(tc.time)
			suite.Require().Equal(tc.expUnvestedCoins, unvestedCoins)
			lockedUpCoins := va.GetLockedUpCoins(tc.time)
			suite.Require().Equal(tc.expLockedUpCoins, lockedUpCoins)
			unlockedCoins := va.GetUnlockedCoins(tc.time)
			suite.Require().Equal(tc.expUnlockedCoins, unlockedCoins)
			suite.Require().Equal(tc.expNotSpendable, va.LockedCoins(tc.time))
		})
	}
}

func (suite *VestingAccountTestSuite) TestTrackDelegationUndelegation() {
	now := tmtime.Now()
	endTime := now.Add(24 * time.Hour)

	testCases := []struct {
		name                 string
		delegate             func(*types.ClawbackVestingAccount)
		expDelegatedFree     sdk.Coins
		undelegate           func(*types.ClawbackVestingAccount)
		expUndelegatedFree   sdk.Coins
		expDelegationPanic   bool
		expUndelegationPanic bool
	}{
		{
			"delegate and undelegated all vested coins",
			func(va *types.ClawbackVestingAccount) {
				va.TrackDelegation(endTime, origCoins, origCoins)
			},
			origCoins,
			func(va *types.ClawbackVestingAccount) {
				va.TrackUndelegation(origCoins)
			},
			sdk.Coins{},
			false,
			false,
		},
		{
			"delegate and undelegate half of vested coins",
			func(va *types.ClawbackVestingAccount) {
				va.TrackDelegation(now, origCoins, vestingPeriods[0].Amount)
			},
			vestingPeriods[0].Amount,
			func(va *types.ClawbackVestingAccount) {
				va.TrackUndelegation(vestingPeriods[0].Amount)
			},
			sdk.Coins{},
			false,
			false,
		},
		{
			"no modifications when delegation amount is zero or not enough funds",
			func(va *types.ClawbackVestingAccount) {
				va.TrackDelegation(now, origCoins, sdk.Coins{sdk.NewInt64Coin(stakeDenom, 1000000)})
			},
			vestingPeriods[0].Amount,
			func(va *types.ClawbackVestingAccount) {
				va.TrackUndelegation(vestingPeriods[0].Amount)
			},
			sdk.Coins{},
			true,
			false,
		},
		{
			"no modifications when undelegation amount is zero or not enough funds",
			func(va *types.ClawbackVestingAccount) {
				va.TrackDelegation(now, origCoins, origCoins)
			},
			vestingPeriods[0].Amount,
			func(va *types.ClawbackVestingAccount) {
				va.TrackUndelegation(sdk.Coins{sdk.NewInt64Coin(stakeDenom, 0)})
			},
			sdk.Coins{},
			false,
			true,
		},
		{
			"vest 50% and delegate to two validator and undelegate from one validator that got slashed 50%",
			func(va *types.ClawbackVestingAccount) {
				va.TrackDelegation(now.Add(17*time.Hour), origCoins, sdk.Coins{sdk.NewInt64Coin(stakeDenom, 50)})
				va.TrackDelegation(now.Add(17*time.Hour), origCoins, sdk.Coins{sdk.NewInt64Coin(stakeDenom, 50)})
			},
			sdk.Coins{sdk.NewInt64Coin(stakeDenom, 100)},
			func(va *types.ClawbackVestingAccount) {
				va.TrackUndelegation(sdk.Coins{sdk.NewInt64Coin(stakeDenom, 25)})
			},
			sdk.Coins{sdk.NewInt64Coin(stakeDenom, 75)},
			false,
			false,
		},
		{
			"vest 50% and delegate to two validator and undelegate from one validator that got slashed 50% and undelegate from the other validator that did not get slashed",
			func(va *types.ClawbackVestingAccount) {
				va.TrackDelegation(now.Add(17*time.Hour), origCoins, sdk.Coins{sdk.NewInt64Coin(stakeDenom, 50)})
				va.TrackDelegation(now.Add(17*time.Hour), origCoins, sdk.Coins{sdk.NewInt64Coin(stakeDenom, 50)})
			},
			sdk.Coins{sdk.NewInt64Coin(stakeDenom, 100)},
			func(va *types.ClawbackVestingAccount) {
				va.TrackUndelegation(sdk.Coins{sdk.NewInt64Coin(stakeDenom, 25)})
				va.TrackUndelegation(sdk.Coins{sdk.NewInt64Coin(stakeDenom, 50)})
			},
			sdk.Coins{sdk.NewInt64Coin(stakeDenom, 25)},
			false,
			false,
		},
	}
	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			addr := sdk.AccAddress(utiltx.GenerateAddress().Bytes())
			bacc := authtypes.NewBaseAccountWithAddress(addr)
			va := types.NewClawbackVestingAccount(bacc, sdk.AccAddress([]byte("funder")), origCoins, now, lockupPeriods, vestingPeriods)
			if tc.expDelegationPanic { //nolint:gocritic
				suite.Require().Panics(func() {
					tc.delegate(va)
				})
			} else if tc.expUndelegationPanic {
				suite.Require().Panics(func() {
					tc.undelegate(va)
				})
			} else {
				var emptyCoins sdk.Coins
				// Track Delegation
				tc.delegate(va)
				suite.Require().Equal(emptyCoins, va.DelegatedVesting)
				suite.Require().Equal(tc.expDelegatedFree, va.DelegatedFree)

				// Track Undelegation
				tc.undelegate(va)
				suite.Require().Equal(emptyCoins, va.DelegatedVesting)
				suite.Require().Equal(tc.expUndelegatedFree, va.DelegatedFree)
			}
		})
	}
}

func (suite *VestingAccountTestSuite) TestComputeClawback() {
	fee := func(x int64) sdk.Coin { return sdk.NewInt64Coin(feeDenom, x) }
	stake := func(x int64) sdk.Coin { return sdk.NewInt64Coin(stakeDenom, x) }
	now := tmtime.Now()
	lockupPeriods := sdkvesting.Periods{
		{Length: int64(12 * 3600), Amount: sdk.NewCoins(fee(1000), stake(100))}, // noon
	}
	vestingPeriods := sdkvesting.Periods{
		{Length: int64(8 * 3600), Amount: sdk.NewCoins(fee(200))},            // 8am
		{Length: int64(1 * 3600), Amount: sdk.NewCoins(fee(200), stake(50))}, // 9am
		{Length: int64(6 * 3600), Amount: sdk.NewCoins(fee(200), stake(50))}, // 3pm
		{Length: int64(2 * 3600), Amount: sdk.NewCoins(fee(200))},            // 5pm
		{Length: int64(1 * 3600), Amount: sdk.NewCoins(fee(200))},            // 6pm
	}

	testCases := []struct {
		name               string
		time               int64
		expClawedBack      sdk.Coins
		expOriginalVesting sdk.Coins
		expLockupPeriods   sdkvesting.Periods
		expVestingPeriods  sdkvesting.Periods
	}{
		{
			"should claw back everything if clawed back before start time",
			now.Add(-time.Hour).Unix(),
			origCoins,
			sdk.Coins{},
			sdkvesting.Periods{},
			sdkvesting.Periods{},
		},
		{
			"should clawback everything before any vesting or lockup period passes",
			now.Unix(),
			sdk.NewCoins(fee(1000), stake(100)),
			sdk.Coins{},
			sdkvesting.Periods{},
			sdkvesting.Periods{},
		},
		{
			"it should clawback after two vesting periods and before the first lock period",
			now.Add(11 * time.Hour).Unix(),
			sdk.Coins{fee(600), stake(50)}, // last 3 periods are still vesting
			sdk.Coins{fee(400), stake(50)}, // first 2 periods
			sdkvesting.Periods{{Length: int64(12 * 3600), Amount: sdk.NewCoins(fee(400), stake(50))}},
			vestingPeriods[:2],
		},
		{
			"should clawback zero coins after all vesting and locked periods",
			now.Add(23 * time.Hour).Unix(),
			sdk.Coins{},
			sdk.Coins{fee(1000), stake(100)},
			lockupPeriods,
			vestingPeriods,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			addr := sdk.AccAddress(utiltx.GenerateAddress().Bytes())
			bacc := authtypes.NewBaseAccountWithAddress(addr)
			va := types.NewClawbackVestingAccount(bacc, sdk.AccAddress([]byte("funder")), origCoins, now, lockupPeriods, vestingPeriods)

			va2, amt := va.ComputeClawback(tc.time)

			suite.Require().Equal(tc.expClawedBack, amt)
			suite.Require().Equal(tc.expOriginalVesting, va2.OriginalVesting)
			suite.Require().Equal(tc.expLockupPeriods, va2.LockupPeriods)
			suite.Require().Equal(tc.expVestingPeriods, va2.VestingPeriods)
		})
	}
}

// getPercentOfVestingCoins is a helper function to calculate
// the specified percentage of the coins in the vesting schedule
func getPercentOfVestingCoins(percentage int64) sdk.Coins {
	if percentage < 0 || percentage > 100 {
		panic("invalid percentage passed!")
	}
	var retCoins sdk.Coins
	for _, coin := range origCoins {
		retCoins = retCoins.Add(sdk.NewCoin(coin.Denom, coin.Amount.MulRaw(percentage).QuoRaw(100)))
	}
	return retCoins
}
