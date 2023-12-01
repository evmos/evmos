package types_test

import (
	"testing"
	"time"

	tmtime "github.com/cometbft/cometbft/types/time"
	"github.com/stretchr/testify/suite"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	sdkvesting "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"

	utiltx "github.com/evmos/evmos/v16/testutil/tx"
	"github.com/evmos/evmos/v16/x/vesting/types"
)

var (
	stakeDenom    = "stake"
	feeDenom      = "fee"
	lockupPeriods = sdkvesting.Periods{
		sdkvesting.Period{Length: int64(16 * 60 * 60), Amount: sdk.NewCoins(sdk.NewInt64Coin(feeDenom, 1000), sdk.NewInt64Coin(stakeDenom, 100))},
	}
	vestingPeriods = sdkvesting.Periods{
		sdkvesting.Period{Length: int64(12 * 60 * 60), Amount: sdk.Coins{sdk.NewInt64Coin(feeDenom, 500), sdk.NewInt64Coin(stakeDenom, 50)}},
		sdkvesting.Period{Length: int64(6 * 60 * 60), Amount: sdk.Coins{sdk.NewInt64Coin(feeDenom, 250), sdk.NewInt64Coin(stakeDenom, 25)}},
		sdkvesting.Period{Length: int64(6 * 60 * 60), Amount: sdk.Coins{sdk.NewInt64Coin(feeDenom, 250), sdk.NewInt64Coin(stakeDenom, 25)}},
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
	addr := sdk.AccAddress(utiltx.GenerateAddress().Bytes())
	baseAcc := authtypes.NewBaseAccountWithAddress(addr)
	initialVesting := sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 50))

	testCases := []struct {
		name   string
		acc    authtypes.GenesisAccount
		expErr bool
	}{
		{
			"Clawback vesting account - pass",
			types.NewClawbackVestingAccount(
				baseAcc,
				sdk.AccAddress([]byte("the funder")),
				initialVesting,
				time.Now(),
				sdkvesting.Periods{sdkvesting.Period{Length: 101, Amount: initialVesting}},
				sdkvesting.Periods{sdkvesting.Period{Length: 201, Amount: initialVesting}},
			),
			false,
		},
		{
			"Clawback vesting account - invalid vesting end",
			&types.ClawbackVestingAccount{
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
			true,
		},
		{
			"Clawback vesting account - lockup too long",
			&types.ClawbackVestingAccount{
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
			true,
		},
		{
			"Clawback vesting account - invalid lockup coins",
			&types.ClawbackVestingAccount{
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
			true,
		},
		{
			"Clawback vesting account - vesting too long",
			&types.ClawbackVestingAccount{
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
			true,
		},
		{
			"Clawback vesting account - invalid vesting coins",
			&types.ClawbackVestingAccount{
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
			true,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			suite.Require().Equal(tc.expErr, tc.acc.Validate() != nil)
		})
	}
}

func (suite *VestingAccountTestSuite) TestGetVestedVestingLockedCoins() {
	now := tmtime.Now()
	endTime := now.Add(24 * time.Hour)
	addr := sdk.AccAddress(utiltx.GenerateAddress().Bytes())
	bacc := authtypes.NewBaseAccountWithAddress(addr)
	va := types.NewClawbackVestingAccount(bacc, sdk.AccAddress([]byte("funder")), origCoins, now, lockupPeriods, vestingPeriods)

	testCases := []struct {
		name              string
		time              time.Time
		expVestedCoins    sdk.Coins
		expUnvestedCoins  sdk.Coins
		expSpendableCoins sdk.Coins
	}{
		{
			"no coins vested at the beginning of the vesting schedule",
			now,
			nil,
			origCoins,
			origCoins,
		},
		{
			"all coins vested at the end of the vesting schedule",
			endTime,
			origCoins,
			sdk.Coins{},
			sdk.NewCoins(),
		},
		{
			"no coins vested during first vesting period",
			now.Add(6 * time.Hour),
			nil,
			origCoins,
			origCoins,
		},
		{
			"no coins vested after period 1 before unlocking",
			now.Add(14 * time.Hour),
			nil,
			origCoins,
			origCoins,
		},
		{
			"50 percent of coins vested after period 1 at unlocking",
			now.Add(16 * time.Hour),
			sdk.Coins{sdk.NewInt64Coin(feeDenom, 500), sdk.NewInt64Coin(stakeDenom, 50)},
			sdk.Coins{sdk.NewInt64Coin(feeDenom, 500), sdk.NewInt64Coin(stakeDenom, 50)},
			sdk.Coins{sdk.NewInt64Coin(feeDenom, 500), sdk.NewInt64Coin(stakeDenom, 50)},
		},
		{
			"period 2 coins don't vest until period is over",
			now.Add(17 * time.Hour),
			sdk.Coins{sdk.NewInt64Coin(feeDenom, 500), sdk.NewInt64Coin(stakeDenom, 50)},
			sdk.Coins{sdk.NewInt64Coin(feeDenom, 500), sdk.NewInt64Coin(stakeDenom, 50)},
			sdk.Coins{sdk.NewInt64Coin(feeDenom, 500), sdk.NewInt64Coin(stakeDenom, 50)},
		},
		{
			"75 percent of coins vested after period 2",
			now.Add(18 * time.Hour),
			sdk.Coins{sdk.NewInt64Coin(feeDenom, 750), sdk.NewInt64Coin(stakeDenom, 75)},
			sdk.Coins{sdk.NewInt64Coin(feeDenom, 250), sdk.NewInt64Coin(stakeDenom, 25)},
			sdk.Coins{sdk.NewInt64Coin(feeDenom, 250), sdk.NewInt64Coin(stakeDenom, 25)},
		},
		{
			"100 percent of coins vested",
			now.Add(48 * time.Hour),
			origCoins,
			sdk.Coins{},
			sdk.NewCoins(),
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			vestedCoins := va.GetVestedCoins(tc.time)
			suite.Require().Equal(tc.expVestedCoins, vestedCoins)
			unvestedCoins := va.GetVestingCoins(tc.time)
			suite.Require().Equal(tc.expUnvestedCoins, unvestedCoins)
			spendableCoins := va.LockedCoins(tc.time)
			suite.Require().Equal(tc.expSpendableCoins, spendableCoins)
		})
	}
}

func (suite *VestingAccountTestSuite) TestGetVestedUnvestedLockedOnly() {
	now := tmtime.Now()
	endTime := now.Add(24 * time.Hour)
	addr := sdk.AccAddress(utiltx.GenerateAddress().Bytes())
	bacc := authtypes.NewBaseAccountWithAddress(addr)
	va := types.NewClawbackVestingAccount(bacc, sdk.AccAddress([]byte("funder")), origCoins, now, lockupPeriods, vestingPeriods)

	testCases := []struct {
		name             string
		time             time.Time
		expVestedCoins   sdk.Coins
		expUnvestedCoins sdk.Coins
		expLockedCoins   sdk.Coins
	}{
		{
			"no coins vested at the beginning of the vesting schedule",
			now,
			sdk.Coins{},
			origCoins,
			origCoins,
		},
		{
			"all coins vested at the end of the vesting schedule",
			endTime,
			origCoins,
			sdk.Coins{},
			sdk.Coins{},
		},
		{
			"no coins vested during first vesting period",
			now.Add(6 * time.Hour),
			sdk.Coins{},
			origCoins,
			origCoins,
		},
		{
			"50 percent of coins vested after period 1 before unlocking",
			now.Add(14 * time.Hour),
			sdk.Coins{sdk.NewInt64Coin(feeDenom, 500), sdk.NewInt64Coin(stakeDenom, 50)},
			sdk.Coins{sdk.NewInt64Coin(feeDenom, 500), sdk.NewInt64Coin(stakeDenom, 50)},
			origCoins,
		},
		{
			"50 percent of coins vested after period 1 at unlocking",
			now.Add(16 * time.Hour),
			sdk.Coins{sdk.NewInt64Coin(feeDenom, 500), sdk.NewInt64Coin(stakeDenom, 50)},
			sdk.Coins{sdk.NewInt64Coin(feeDenom, 500), sdk.NewInt64Coin(stakeDenom, 50)},
			sdk.Coins{},
		},
		{
			"period 2 coins don't vest until period is over",
			now.Add(17 * time.Hour),
			sdk.Coins{sdk.NewInt64Coin(feeDenom, 500), sdk.NewInt64Coin(stakeDenom, 50)},
			sdk.Coins{sdk.NewInt64Coin(feeDenom, 500), sdk.NewInt64Coin(stakeDenom, 50)},
			sdk.Coins{},
		},
		{
			"75 percent of coins vested after period 2",
			now.Add(18 * time.Hour),
			sdk.Coins{sdk.NewInt64Coin(feeDenom, 750), sdk.NewInt64Coin(stakeDenom, 75)},
			sdk.Coins{sdk.NewInt64Coin(feeDenom, 250), sdk.NewInt64Coin(stakeDenom, 25)},
			sdk.Coins{},
		},
		{
			"100 percent of coins vested",
			now.Add(48 * time.Hour),
			origCoins,
			sdk.Coins{},
			sdk.Coins{},
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			vestedCoins := va.GetVestedOnly(tc.time)
			suite.Require().Equal(tc.expVestedCoins, vestedCoins)
			unvestedCoins := va.GetUnvestedOnly(tc.time)
			suite.Require().Equal(tc.expUnvestedCoins, unvestedCoins)
			lockedCoins := va.GetLockedOnly(tc.time)
			suite.Require().Equal(tc.expLockedCoins, lockedCoins)
		})
	}
}

func (suite *VestingAccountTestSuite) TestTrackDelegationUndelegation() {
	now := tmtime.Now()
	endTime := now.Add(24 * time.Hour)

	testCases := []struct {
		name                   string
		delegate               func(*types.ClawbackVestingAccount)
		expDelegatedUnvested   sdk.Coins
		expDelegatedFree       sdk.Coins
		undelegate             func(*types.ClawbackVestingAccount)
		expUndelegatedUnvested sdk.Coins
		expUndelegatedFree     sdk.Coins
		expDelegationPanic     bool
		expUndelegationPanic   bool
	}{
		{
			"delegate and undelegate all unvested coins",
			func(va *types.ClawbackVestingAccount) {
				va.TrackDelegation(now, origCoins, origCoins)
			},
			origCoins,
			nil,
			func(va *types.ClawbackVestingAccount) {
				va.TrackUndelegation(origCoins)
			},
			sdk.Coins{},
			nil,
			false,
			false,
		},
		{
			"delegate and undelegated all vested coins",
			func(va *types.ClawbackVestingAccount) {
				va.TrackDelegation(endTime, origCoins, origCoins)
			},
			nil,
			origCoins,
			func(va *types.ClawbackVestingAccount) {
				va.TrackUndelegation(origCoins)
			},
			nil,
			sdk.Coins{},
			false,
			false,
		},
		{
			"delegate and undelegate half of unvested coins",
			func(va *types.ClawbackVestingAccount) {
				va.TrackDelegation(now, origCoins, vestingPeriods[0].Amount)
			},
			vestingPeriods[0].Amount,
			nil,
			func(va *types.ClawbackVestingAccount) {
				va.TrackUndelegation(vestingPeriods[0].Amount)
			},
			sdk.Coins{},
			nil,
			false,
			false,
		},
		{
			"no modifications when delegation amount is zero or not enough funds",
			func(va *types.ClawbackVestingAccount) {
				va.TrackDelegation(now, origCoins, sdk.Coins{sdk.NewInt64Coin(stakeDenom, 1000000)})
			},
			vestingPeriods[0].Amount,
			nil,
			func(va *types.ClawbackVestingAccount) {
				va.TrackUndelegation(vestingPeriods[0].Amount)
			},
			sdk.Coins{},
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
			nil,
			func(va *types.ClawbackVestingAccount) {
				va.TrackUndelegation(sdk.Coins{sdk.NewInt64Coin(stakeDenom, 0)})
			},
			sdk.Coins{},
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
			sdk.Coins{sdk.NewInt64Coin(stakeDenom, 50)},
			sdk.Coins{sdk.NewInt64Coin(stakeDenom, 50)},
			func(va *types.ClawbackVestingAccount) {
				va.TrackUndelegation(sdk.Coins{sdk.NewInt64Coin(stakeDenom, 25)})
			},
			sdk.Coins{sdk.NewInt64Coin(stakeDenom, 50)},
			sdk.Coins{sdk.NewInt64Coin(stakeDenom, 25)},
			false,
			false,
		},
		{
			"vest 50% and delegate to two validator and undelegate from one validator that got slashed 50% and undelegate from the other validator that did not get slashed",
			func(va *types.ClawbackVestingAccount) {
				va.TrackDelegation(now.Add(17*time.Hour), origCoins, sdk.Coins{sdk.NewInt64Coin(stakeDenom, 50)})
				va.TrackDelegation(now.Add(17*time.Hour), origCoins, sdk.Coins{sdk.NewInt64Coin(stakeDenom, 50)})
			},
			sdk.Coins{sdk.NewInt64Coin(stakeDenom, 50)},
			sdk.Coins{sdk.NewInt64Coin(stakeDenom, 50)},
			func(va *types.ClawbackVestingAccount) {
				va.TrackUndelegation(sdk.Coins{sdk.NewInt64Coin(stakeDenom, 25)})
				va.TrackUndelegation(sdk.Coins{sdk.NewInt64Coin(stakeDenom, 50)})
			},
			sdk.Coins{sdk.NewInt64Coin(stakeDenom, 25)},
			sdk.Coins{},
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
				// Track Delegation
				tc.delegate(va)
				suite.Require().Equal(tc.expDelegatedUnvested, va.DelegatedVesting)
				suite.Require().Equal(tc.expDelegatedFree, va.DelegatedFree)

				// Track Undelegation
				tc.undelegate(va)
				suite.Require().Equal(tc.expUndelegatedUnvested, va.DelegatedVesting)
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
