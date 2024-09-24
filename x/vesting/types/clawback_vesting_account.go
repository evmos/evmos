// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package types

import (
	"errors"
	"time"

	errorsmod "cosmossdk.io/errors"
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

// GetLockedUpVestedCoins returns the total number of vested coins that are locked.
func (va ClawbackVestingAccount) GetLockedUpVestedCoins(blockTime time.Time) sdk.Coins {
	return va.GetVestedCoins(blockTime).Sub(va.GetUnlockedVestedCoins(blockTime)...)
}

// GetUnlockedVestedCoins returns the total number of vested coins that are unlocked.
// If no coins are vested and unlocked, nil is returned.
func (va ClawbackVestingAccount) GetUnlockedVestedCoins(blockTime time.Time) sdk.Coins {
	coins := va.GetUnlockedCoins(blockTime).Min(va.GetVestedCoins(blockTime))
	if coins.IsZero() {
		return sdk.Coins{}
	}
	return coins
}

// GetVestingCoins returns the total number of vesting coins (unvested coins).
// If no coins are vesting, nil is returned.
func (va ClawbackVestingAccount) GetVestingCoins(blockTime time.Time) sdk.Coins {
	return va.OriginalVesting.Sub(va.GetVestedCoins(blockTime)...)
}

// LockedCoins returns the set of coins that are not spendable (i.e. locked or unvested),
// defined as the vesting coins (unvested) plus locked vested coins.
//
// totalAmt = vesting(un/locked) + lockedVested + unlockedVested
//
//	(all)   =   (cannot spend)    (cannot spend)   (CAN spend)
//
// lockedCoins = totalAmt - unlockedVested
func (va ClawbackVestingAccount) LockedCoins(blockTime time.Time) sdk.Coins {
	// Can delegate lockedUpVested coins and this will reduce the bank balance
	// of the account. As long as there're lockedUpVested coins, we'll consider
	// the delegated tokens as lockedUpVested tokens
	// min(lockedUpVested, DelegatedFree)
	//
	// Consider that the "DelegatedFree" coins tracked on delegations refer to vested tokens.
	// These "free" (vested) tokens can be locked up or unlocked
	lockedUpVestedDelegatedCoins := va.DelegatedFree.Min(va.GetLockedUpVestedCoins(blockTime))

	res, isNeg := va.OriginalVesting.SafeSub(va.GetUnlockedVestedCoins(blockTime).Add(lockedUpVestedDelegatedCoins...)...)

	// safety check
	if isNeg {
		return sdk.Coins{}
	}

	return res
}

// TrackDelegation tracks a desired delegation amount by setting the appropriate
// values for the amount of delegated free coins.
// The 'balance' input parameter is the delegator account balance.
// The 'amount' input parameter are the delegated coins
// Note that unvested coins cannot be delegated
func (va *ClawbackVestingAccount) TrackDelegation(_ time.Time, balance, amount sdk.Coins) {
	// Can only delegate vested (free) coins
	for _, coin := range amount {
		baseAmt := balance.AmountOf(coin.Denom)
		// Panic if the delegation amount is zero or if the base coins does not
		// exceed the desired delegation amount.
		if coin.Amount.IsZero() || baseAmt.LT(coin.Amount) {
			panic("delegation attempt with zero coins or insufficient funds")
		}
		va.DelegatedFree = va.DelegatedFree.Add(coin)
	}
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

	// use CoinEq to prevent panic
	if !CoinEq(lockupCoins, va.OriginalVesting) {
		return errors.New("original vesting coins does not match the sum of all coins in lockup periods")
	}

	vestingEnd := va.GetStartTime()
	vestingCoins := sdk.NewCoins()

	for _, p := range va.VestingPeriods {
		vestingEnd += p.Length
		vestingCoins = vestingCoins.Add(p.Amount...)
	}

	if vestingEnd > va.EndTime {
		return errors.New("vesting schedule extends beyond account end time")
	}

	if !CoinEq(vestingCoins, va.OriginalVesting) {
		return errors.New("original vesting coins does not match the sum of all coins in vesting periods")
	}

	return va.BaseVestingAccount.Validate()
}

// GetUnlockedCoins returns the unlocked coins at blockTime.
// Note that these unlocked coins can be vested or unvested
// and is determined by the lockup periods
func (va ClawbackVestingAccount) GetUnlockedCoins(blockTime time.Time) sdk.Coins {
	return ReadSchedule(va.GetStartTime(), va.EndTime, va.LockupPeriods, va.OriginalVesting, blockTime.Unix())
}

// GetLockedUpCoins returns the locked coins at blockTime.
// Note that these locked up coins can be vested or unvested,
// and is determined by the lockup periods
func (va ClawbackVestingAccount) GetLockedUpCoins(blockTime time.Time) sdk.Coins {
	return va.OriginalVesting.Sub(va.GetUnlockedCoins(blockTime)...)
}

// GetVestedCoins returns the vested coins at blockTime.
func (va ClawbackVestingAccount) GetVestedCoins(blockTime time.Time) sdk.Coins {
	return ReadSchedule(va.GetStartTime(), va.EndTime, va.VestingPeriods, va.OriginalVesting, blockTime.Unix())
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
	totalVested := va.GetVestedCoins(time.Unix(clawbackTime, 0))
	totalUnvested := va.GetVestingCoins(time.Unix(clawbackTime, 0))

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

// HasLockedCoins returns true if the block time has not passed all clawback
// account's lockup periods
func (va ClawbackVestingAccount) HasLockedCoins(blockTime time.Time) bool {
	return !va.GetLockedUpCoins(blockTime).IsZero()
}

// AddGrant merges a new clawback vesting grant into an existing
// ClawbackVestingAccount.
func (va *ClawbackVestingAccount) AddGrant(
	grantStartTime int64,
	grantLockupPeriods, grantVestingPeriods sdkvesting.Periods,
	grantCoins sdk.Coins,
) error {
	// check if the clawback vesting account has only been initialized and not yet funded --
	// in that case it's necessary to update the vesting account with the given start time because this is set to zero in the initialization
	if len(va.LockupPeriods) == 0 && len(va.VestingPeriods) == 0 {
		va.StartTime = time.Unix(grantStartTime, 0).UTC()
	}

	// modify schedules for the new grant
	accStartTime := va.GetStartTime()
	newLockupStart, newLockupEnd, newLockupPeriods := DisjunctPeriods(accStartTime, grantStartTime, va.LockupPeriods, grantLockupPeriods)
	newVestingStart, newVestingEnd, newVestingPeriods := DisjunctPeriods(
		accStartTime,
		grantStartTime,
		va.GetVestingPeriods(),
		grantVestingPeriods,
	)

	if newLockupStart != newVestingStart {
		return errorsmod.Wrapf(
			ErrVestingLockup,
			"vesting start time calculation should match lockup start (%d ≠ %d)",
			newVestingStart, newLockupStart,
		)
	}

	va.StartTime = time.Unix(newLockupStart, 0).UTC()
	va.EndTime = Max64(newLockupEnd, newVestingEnd)
	va.LockupPeriods = newLockupPeriods
	va.VestingPeriods = newVestingPeriods
	va.OriginalVesting = va.OriginalVesting.Add(grantCoins...)

	return nil
}
