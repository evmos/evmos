package types

import (
	"errors"
	"time"

	yaml "gopkg.in/yaml.v2"

	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	sdkvesting "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"

	vestexported "github.com/cosmos/cosmos-sdk/x/auth/vesting/exported"
)

// Clawback Vesting Account
var (
	_ vestexported.VestingAccount = (*ClawbackVestingAccount)(nil)
	_ authtypes.GenesisAccount    = (*ClawbackVestingAccount)(nil)
)

// NewClawbackVestingAccount returns a new ClawbackVestingAccount
func NewClawbackVestingAccount(baseAcc *authtypes.BaseAccount, funder sdk.AccAddress, originalVesting sdk.Coins, startTime int64, lockupPeriods, vestingPeriods sdkvesting.Periods) *ClawbackVestingAccount {
	// copy and align schedules to avoid mutating inputs
	lp := make(sdkvesting.Periods, len(lockupPeriods))
	copy(lp, lockupPeriods)
	vp := make(sdkvesting.Periods, len(vestingPeriods))
	copy(vp, vestingPeriods)
	_, endTime := AlignSchedules(startTime, startTime, lp, vp)
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

// GetVestedCoins returns the total number of vested coins. If no coins are vested,
// nil is returned.
func (va ClawbackVestingAccount) GetVestedCoins(blockTime time.Time) sdk.Coins {
	// It's likely that one or the other schedule will be nearly trivial,
	// so there should be little overhead in recomputing the conjunction each time.
	coins := CoinsMin(va.GetUnlockedOnly(blockTime), va.GetVestedOnly(blockTime))
	if coins.IsZero() {
		return nil
	}
	return coins
}

// GetVestingCoins returns the total number of vesting coins. If no coins are
// vesting, nil is returned.
func (va ClawbackVestingAccount) GetVestingCoins(blockTime time.Time) sdk.Coins {
	return va.OriginalVesting.Sub(va.GetVestedCoins(blockTime))
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
	return va.StartTime
}

// GetVestingPeriods returns vesting periods associated with periodic vesting account.
func (va ClawbackVestingAccount) GetVestingPeriods() sdkvesting.Periods {
	return va.VestingPeriods
}

// coinEq returns whether two Coins are equal.
// The IsEqual() method can panic.
func coinEq(a, b sdk.Coins) bool {
	return a.IsAllLTE(b) && b.IsAllLTE(a)
}

// Validate checks for errors on the account fields
func (va ClawbackVestingAccount) Validate() error {
	if va.GetStartTime() >= va.GetEndTime() {
		return errors.New("vesting start-time must be before end-time")
	}

	lockupEnd := va.StartTime
	lockupCoins := sdk.NewCoins()
	for _, p := range va.LockupPeriods {
		lockupEnd += p.Length
		lockupCoins = lockupCoins.Add(p.Amount...)
	}
	if lockupEnd > va.EndTime {
		return errors.New("lockup schedule extends beyond account end time")
	}
	if !coinEq(lockupCoins, va.OriginalVesting) {
		return errors.New("original vesting coins does not match the sum of all coins in lockup periods")
	}

	vestingEnd := va.StartTime
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

type vestingAccountYAML struct {
	Address          sdk.AccAddress `json:"address"`
	PubKey           string         `json:"public_key"`
	AccountNumber    uint64         `json:"account_number"`
	Sequence         uint64         `json:"sequence"`
	OriginalVesting  sdk.Coins      `json:"original_vesting"`
	DelegatedFree    sdk.Coins      `json:"delegated_free"`
	DelegatedVesting sdk.Coins      `json:"delegated_vesting"`
	EndTime          int64          `json:"end_time"`

	// custom fields based on concrete vesting type which can be omitted
	StartTime      int64              `json:"start_time,omitempty"`
	VestingPeriods sdkvesting.Periods `json:"vesting_periods,omitempty"`
}

type getPK interface {
	GetPubKey() cryptotypes.PubKey
}

func getPKString(g getPK) string {
	if pk := g.GetPubKey(); pk != nil {
		return pk.String()
	}
	return ""
}

func marshalYaml(i interface{}) (interface{}, error) {
	bz, err := yaml.Marshal(i)
	if err != nil {
		return nil, err
	}
	return string(bz), nil
}

func (va ClawbackVestingAccount) String() string {
	out, _ := va.MarshalYAML()
	return out.(string)
}

// MarshalYAML returns the YAML representation of a ClawbackVestingAccount.
func (va ClawbackVestingAccount) MarshalYAML() (interface{}, error) {
	accAddr, err := sdk.AccAddressFromBech32(va.Address)
	if err != nil {
		return nil, err
	}

	out := vestingAccountYAML{
		Address:          accAddr,
		AccountNumber:    va.AccountNumber,
		PubKey:           getPKString(va),
		Sequence:         va.Sequence,
		OriginalVesting:  va.OriginalVesting,
		DelegatedFree:    va.DelegatedFree,
		DelegatedVesting: va.DelegatedVesting,
		EndTime:          va.EndTime,
		StartTime:        va.StartTime,
		VestingPeriods:   va.VestingPeriods,
	}
	return marshalYaml(out)
}

// GetUnlockedOnly returns the unlocking schedule at blockTIme.
// Like GetVestedCoins, but only for the lockup component.
func (va ClawbackVestingAccount) GetUnlockedOnly(blockTime time.Time) sdk.Coins {
	return ReadSchedule(va.StartTime, va.EndTime, va.LockupPeriods, va.OriginalVesting, blockTime.Unix())
}

// GetVestedOnly returns the vesting schedule and blockTime.
// Like GetVestedCoins, but only for the vesting (in the clawback sense) component.
func (va ClawbackVestingAccount) GetVestedOnly(blockTime time.Time) sdk.Coins {
	return ReadSchedule(va.StartTime, va.EndTime, va.VestingPeriods, va.OriginalVesting, blockTime.Unix())
}

// ComputeClawback returns an account with all future vesting events removed,
// plus the total sum of these events. When removing the future vesting events,
// the lockup schedule will also have to be capped to keep the total sums the same.
// (But future unlocking events might be preserved if they unlock currently vested coins.)
// If the amount returned is zero, then the returned account should be unchanged.
// Does not adjust DelegatedVesting
func (va ClawbackVestingAccount) ComputeClawback(clawbackTime int64) (ClawbackVestingAccount, sdk.Coins) {
	// Compute the truncated vesting schedule and amounts.
	// Work with the schedule as the primary data and recompute derived fields, e.g. OriginalVesting.
	t := va.StartTime
	totalVested := sdk.NewCoins()
	totalUnvested := sdk.NewCoins()
	unvestedIdx := 0
	for i, period := range va.VestingPeriods {
		t += period.Length
		// tie in time goes to clawback
		if t < clawbackTime {
			totalVested = totalVested.Add(period.Amount...)
			unvestedIdx = i + 1
		} else {
			totalUnvested = totalUnvested.Add(period.Amount...)
		}
	}
	newVestingPeriods := va.VestingPeriods[:unvestedIdx]

	// To cap the unlocking schedule to the new total vested, conjunct with a limiting schedule
	capPeriods := []sdkvesting.Period{
		{
			Length: 0,
			Amount: totalVested,
		},
	}
	_, _, newLockupPeriods := ConjunctPeriods(va.StartTime, va.StartTime, va.LockupPeriods, capPeriods)

	// Now construct the new account state
	va.OriginalVesting = totalVested
	va.EndTime = t
	va.LockupPeriods = newLockupPeriods
	va.VestingPeriods = newVestingPeriods
	// DelegatedVesting and DelegatedFree will be adjusted elsewhere

	return va, totalUnvested
}

// updateDelegation returns an account with its delegation bookkeeping modified for clawback,
// given the current disposition of the account's bank and staking state. Also returns
// the modified amount to claw back.
//
// Computation steps:
// - first, compute the total amount in bonded and unbonding states, used for BaseAccount bookkeeping;
// - based on the old bookkeeping, determine the amount lost to slashing since origin;
// - clip the amount to claw back to be at most the full funds in the account;
// - first claw back the unbonded funds, then go after what's delegated;
// - to the remaining delegated amount, add what's slashed;
// - the "encumbered" (locked up and/or vesting) amount of this goes in DV;
// - the remainder of the new delegated amount goes in DF.
func (va ClawbackVestingAccount) UpdateDelegation(encumbered, toClawBack, bonded, unbonding, unbonded sdk.Coins) (ClawbackVestingAccount, sdk.Coins) {
	delegated := bonded.Add(unbonding...)
	oldDelegated := va.DelegatedVesting.Add(va.DelegatedFree...)
	slashed := oldDelegated.Sub(CoinsMin(delegated, oldDelegated))
	total := delegated.Add(unbonded...)
	toClawBack = CoinsMin(toClawBack, total) // might have been slashed
	newDelegated := CoinsMin(delegated, total.Sub(toClawBack)).Add(slashed...)
	va.DelegatedVesting = CoinsMin(encumbered, newDelegated)
	va.DelegatedFree = newDelegated.Sub(va.DelegatedVesting)
	return va, toClawBack
}

// scaleCoins scales the given coins, rounding down.
func ScaleCoins(coins sdk.Coins, scale sdk.Dec) sdk.Coins {
	scaledCoins := sdk.NewCoins()
	for _, coin := range coins {
		amt := coin.Amount.ToDec().Mul(scale).TruncateInt() // round down
		scaledCoins = scaledCoins.Add(sdk.NewCoin(coin.Denom, amt))
	}
	return scaledCoins
}

// minInt returns the minimum of its arguments.
func MinInt(a, b sdk.Int) sdk.Int {
	if a.GT(b) {
		return b
	}
	return a
}
