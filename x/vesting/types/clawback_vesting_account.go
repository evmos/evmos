package types

import (
	"errors"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	vestexported "github.com/cosmos/cosmos-sdk/x/auth/vesting/exported"
	sdkvesting "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
)

var (
	_ vestexported.VestingAccount = (*ClawbackVestingAccount)(nil)
	_ authtypes.GenesisAccount    = (*ClawbackVestingAccount)(nil)
)

// NewClawbackVestingAccount returns a new ClawbackVestingAccount
func NewClawbackVestingAccount(
	baseAcc *authtypes.BaseAccount,
	funder sdk.AccAddress,
	originalVesting sdk.Coins,
	startTime time.Time,
	lockupPeriods,
	vestingPeriods sdkvesting.Periods,
) *ClawbackVestingAccount {
	// copy and align schedules to the same start time to
	// avoid mutating inputs
	lp := make(sdkvesting.Periods, len(lockupPeriods))
	copy(lp, lockupPeriods)
	vp := make(sdkvesting.Periods, len(vestingPeriods))
	copy(vp, vestingPeriods)

	_, endTime := AlignSchedules(startTime.Unix(), startTime.Unix(), lp, vp)

	baseVestingAcc := &sdkvesting.BaseVestingAccount{
		BaseAccount:     baseAcc,
		OriginalVesting: originalVesting,
		EndTime:         endTime,
	}

	return &ClawbackVestingAccount{
		BaseVestingAccount: baseVestingAcc,
		FunderAddress:      funder.String(),
		StartTime:          startTime,
		LockupPeriods:      lp,
		VestingPeriods:     vp,
	}
}

// GetVestedCoins returns the total number of vested coins that are still in lockup. If no coins are
// vested, nil is returned.
func (va ClawbackVestingAccount) GetVestedCoins(blockTime time.Time) sdk.Coins {
	// It's likely that one or the other schedule will be nearly trivial,
	// so there should be little overhead in recomputing the conjunction each time.
	coins := va.GetUnlockedOnly(blockTime).Min(va.GetVestedOnly(blockTime))
	if coins.IsZero() {
		return nil
	}
	return coins
}

// GetVestingCoins returns the total number of vesting coins. If no coins are
// vesting, nil is returned.
func (va ClawbackVestingAccount) GetVestingCoins(blockTime time.Time) sdk.Coins {
	return va.OriginalVesting.Sub(va.GetVestedCoins(blockTime)...)
}

// LockedCoins returns the set of coins that are not spendable (i.e. locked),
// defined as the vesting coins that are not delegated.
func (va ClawbackVestingAccount) LockedCoins(blockTime time.Time) sdk.Coins {
	return va.BaseVestingAccount.LockedCoinsFromVesting(va.GetVestingCoins(blockTime))
}

// TrackDelegation tracks a desired delegation amount by setting the appropriate
// values for the amount of delegated vesting, delegated free, and reducing the
// overall amount of base coins.
func (va *ClawbackVestingAccount) TrackDelegation(blockTime time.Time, balance, amount sdk.Coins) {
	va.BaseVestingAccount.TrackDelegation(balance, va.GetVestingCoins(blockTime), amount)
}

// GetStartTime returns the time when vesting starts for a periodic vesting
// account.
func (va ClawbackVestingAccount) GetStartTime() int64 {
	return va.StartTime.Unix()
}

// GetVestingPeriods returns vesting periods associated with periodic vesting account.
func (va ClawbackVestingAccount) GetVestingPeriods() sdkvesting.Periods {
	return va.VestingPeriods
}

// Validate checks for errors on the account fields
func (va ClawbackVestingAccount) Validate() error {
	if va.GetStartTime() >= va.GetEndTime() {
		return errors.New("vesting start-time must be before end-time")
	}

	lockupEnd := va.GetStartTime()
	lockupCoins := sdk.NewCoins()

	for _, p := range va.LockupPeriods {
		lockupEnd += p.Length
		lockupCoins = lockupCoins.Add(p.Amount...)
	}

	if lockupEnd > va.EndTime {
		return errors.New("lockup schedule extends beyond account end time")
	}

	// use coinEq to prevent panic
	if !coinEq(lockupCoins, va.OriginalVesting) {
		return errors.New("original vesting coins does not match the sum of all coins in lockup periods")
	}

	vestingEnd := va.GetStartTime()
	vestingCoins := sdk.NewCoins()

	for _, p := range va.VestingPeriods {
		vestingEnd += p.Length
		vestingCoins = vestingCoins.Add(p.Amount...)
	}

	if vestingEnd > va.EndTime {
		return errors.New("vesting schedule exteds beyond account end time")
	}

	if !coinEq(vestingCoins, va.OriginalVesting) {
		return errors.New("original vesting coins does not match the sum of all coins in vesting periods")
	}

	return va.BaseVestingAccount.Validate()
}

// GetUnlockedOnly returns the unlocking schedule at blockTIme.
func (va ClawbackVestingAccount) GetUnlockedOnly(blockTime time.Time) sdk.Coins {
	return ReadSchedule(va.GetStartTime(), va.EndTime, va.LockupPeriods, va.OriginalVesting, blockTime.Unix())
}

// GetLockedOnly returns the locking schedule at blockTIme.
func (va ClawbackVestingAccount) GetLockedOnly(blockTime time.Time) sdk.Coins {
	return va.OriginalVesting.Sub(va.GetUnlockedOnly(blockTime)...)
}

// GetVestedOnly returns the vesting schedule at blockTime.
func (va ClawbackVestingAccount) GetVestedOnly(blockTime time.Time) sdk.Coins {
	return ReadSchedule(va.GetStartTime(), va.EndTime, va.VestingPeriods, va.OriginalVesting, blockTime.Unix())
}

// GetUnvestedOnly returns the unvesting schedule at blockTime.
func (va ClawbackVestingAccount) GetUnvestedOnly(blockTime time.Time) sdk.Coins {
	totalUnvested := va.OriginalVesting.Sub(va.GetVestedOnly(blockTime)...)
	if totalUnvested == nil {
		totalUnvested = sdk.Coins{}
	}
	return totalUnvested
}

// GetPassedPeriodCount returns the amount of passed periods at blockTime.
func (va ClawbackVestingAccount) GetPassedPeriodCount(blockTime time.Time) int {
	return ReadPastPeriodCount(va.GetStartTime(), va.EndTime, va.VestingPeriods, blockTime.Unix())
}

// ComputeClawback returns an account with all future vesting events removed and
// the clawback amount (total sum of these events). Future unlocking events are
// preserved and update in case unlocked vested coins remain after clawback.
func (va ClawbackVestingAccount) ComputeClawback(
	clawbackTime int64,
) (ClawbackVestingAccount, sdk.Coins) {
	// if the clawback time is before the vesting start time, perform a no-op
	// as there is nothing to clawback
	// NOTE: error must be checked during message execution
	if clawbackTime < va.GetStartTime() {
		return va, sdk.Coins{}
	}

	totalVested := va.GetVestedOnly(time.Unix(clawbackTime, 0))
	totalUnvested := va.GetUnvestedOnly(time.Unix(clawbackTime, 0))

	// Remove all unvested periods from the schedule
	passedPeriodID := va.GetPassedPeriodCount(time.Unix(clawbackTime, 0))
	newVestingPeriods := va.VestingPeriods[:passedPeriodID]
	newVestingEnd := va.GetStartTime() + newVestingPeriods.TotalLength()

	// Cap the unlocking schedule to the new total vested.
	//  - If lockup has already passed, all vested coins are unlocked.
	//  - If lockup has not passed, the vested coins, are still locked.
	capPeriods := sdkvesting.Periods{
		{
			Length: 0,
			Amount: totalVested,
		},
	}

	// minimum of the 2 periods
	_, newLockingEnd, newLockupPeriods := ConjunctPeriods(va.GetStartTime(), va.GetStartTime(), va.LockupPeriods, capPeriods)

	// Now construct the new account state
	va.OriginalVesting = totalVested
	va.EndTime = Max64(newVestingEnd, newLockingEnd)
	va.LockupPeriods = newLockupPeriods
	va.VestingPeriods = newVestingPeriods

	return va, totalUnvested
}

// HasLockedCoins returns true if the blocktime has not passed all clawback
// account's lockup periods
func (va ClawbackVestingAccount) HasLockedCoins(blockTime time.Time) bool {
	return !va.GetLockedOnly(blockTime).IsZero()
}
