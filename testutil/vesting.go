// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)
package testutil

import (
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkvesting "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	"github.com/evmos/evmos/v20/utils"
)

type vestingSchedule struct {
	CliffMonths int64
	// CliffPeriodLength in seconds
	CliffPeriodLength int64
	NumLockupPeriods  int64
	LockupMonths      int64
	// LockupPeriodLength in seconds
	LockupPeriodLength int64
	NumVestingPeriods  int64
	// VestingPeriodLength in seconds
	VestingPeriodLength    int64
	TotalVestingCoins      sdk.Coins
	VestedCoinsPerPeriod   sdk.Coins
	UnlockedCoinsPerLockup sdk.Coins
	VestingPeriods         []sdkvesting.Period
	LockupPeriods          []sdkvesting.Period
}

// Vesting schedule for tests that use vesting account
var (
	TestVestingSchedule vestingSchedule
	// Monthly vesting period
	stakeDenom    = utils.BaseDenom
	amt           = math.NewInt(1e17)
	vestingLength = int64(60 * 60 * 24 * 30) // in seconds
	vestingAmt    = sdk.NewCoins(sdk.NewCoin(stakeDenom, amt))
	vestingPeriod = sdkvesting.Period{Length: vestingLength, Amount: vestingAmt}

	// 4 years vesting total
	periodsTotal    = int64(48)
	vestingAmtTotal = sdk.NewCoins(sdk.NewCoin(stakeDenom, amt.Mul(math.NewInt(periodsTotal))))

	// 6 month cliff
	cliff       = int64(6)
	cliffLength = vestingLength * cliff
	cliffAmt    = sdk.NewCoins(sdk.NewCoin(stakeDenom, amt.Mul(math.NewInt(cliff))))
	cliffPeriod = sdkvesting.Period{Length: cliffLength, Amount: cliffAmt}

	// 12 month lockup
	lockup       = int64(12) // 12 months
	lockupLength = vestingLength * lockup
	// Unlock at 12 and 24 months
	numLockupPeriods = int64(2)
	// Unlock half of the total vest in each unlock event. By default, all tokens are
	// unlocked after surpassing the final period.
	unlockedPerLockup = vestingAmtTotal.QuoInt(math.NewInt(numLockupPeriods))
	lockupPeriod      = sdkvesting.Period{Length: lockupLength, Amount: unlockedPerLockup}
	lockupPeriods     = make(sdkvesting.Periods, numLockupPeriods)
	// add initial cliff to vesting periods
	vestingPeriods = sdkvesting.Periods{cliffPeriod}
)

func init() {
	for i := range lockupPeriods {
		lockupPeriods[i] = lockupPeriod
	}

	// Create vesting periods with initial cliff
	for p := int64(1); p <= periodsTotal-cliff; p++ {
		vestingPeriods = append(vestingPeriods, vestingPeriod)
	}

	TestVestingSchedule = vestingSchedule{
		CliffMonths:            cliff,
		CliffPeriodLength:      cliffLength,
		NumLockupPeriods:       numLockupPeriods,
		NumVestingPeriods:      periodsTotal,
		LockupMonths:           lockup,
		LockupPeriodLength:     lockupLength,
		VestingPeriodLength:    vestingLength,
		TotalVestingCoins:      vestingAmtTotal,
		VestedCoinsPerPeriod:   vestingAmt,
		UnlockedCoinsPerLockup: unlockedPerLockup,
		VestingPeriods:         vestingPeriods,
		LockupPeriods:          lockupPeriods,
	}
}
