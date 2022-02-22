package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/tharsis/evmos/x/vesting/types"
)

// TODO Can we remove, if delegating unvested coins is perhibited?
// PostReward encumbers a previously-deposited reward according to the current
// vesting apportionment of staking. Note that rewards might be unvested, but
// still locked.
func (k Keeper) PostReward(
	ctx sdk.Context,
	va types.ClawbackVestingAccount,
	reward sdk.Coins,
) {
	// Find the scheduled amount of vested and unvested staking tokens
	bondDenom := k.stakingKeeper.BondDenom(ctx)
	vested := types.ReadSchedule(va.StartTime, va.EndTime, va.VestingPeriods, va.OriginalVesting, ctx.BlockTime().Unix()).AmountOf(bondDenom)
	unvested := va.OriginalVesting.AmountOf(bondDenom).Sub(vested)

	// ! CODE ANALYSIS
	if unvested.IsZero() {
		// no need to adjust the vesting schedule
		return
	}

	if vested.IsZero() { // ! never reached
		// all staked tokens must be unvested
		k.distributeReward(ctx, va, bondDenom, reward)
		return
	}

	// Determine pre-reward total
	bonded := k.GetDelegatorBonded(ctx, va.GetAddress())
	unbonding := k.GetDelegatorUnbonding(ctx, va.GetAddress())
	unbonded := k.bankKeeper.GetBalance(ctx, va.GetAddress(), bondDenom).Amount
	total := bonded.Add(unbonding).Add(unbonded)
	total = total.Sub(types.MinInt(total, reward.AmountOf(bondDenom)))

	// Adjust vested/unvested for the actual amount in the account, preferring
	// them to be unvested. Amount may have been reduced by slashing/transfers
	unvested = types.MinInt(unvested, total)
	vested = total.Sub(unvested)

	// ! In Evmos case vested >= bonded because unvested coins can't be bonded
	// Restrict logic to just the bonded tokens, preferring them to be vested
	vested = types.MinInt(vested, bonded) // ! => vested == bonded
	unvested = bonded.Sub(vested)         // ! => unvested == 0

	// Compute the unvested amount of reward and add to vesting schedule
	if unvested.IsZero() { // ! always returns
		return
	}
	// ! Unreached code
	if vested.IsZero() {
		k.distributeReward(ctx, va, bondDenom, reward)
		return
	}
	unvestedRatio := unvested.ToDec().QuoTruncate(bonded.ToDec()) // round down
	unvestedReward := types.ScaleCoins(reward, unvestedRatio)
	k.distributeReward(ctx, va, bondDenom, unvestedReward)
}

// distributeReward distributes the reward amongst all upcoming periods in the
// vesting schedule. The amount of the reward for each period is scaled
// proportionally the upcoming vesting schedule.
func (k Keeper) distributeReward(
	ctx sdk.Context,
	va types.ClawbackVestingAccount,
	bondDenom string,
	reward sdk.Coins,
) {
	// Get total unvested tokens from all upcoming periods
	now := ctx.BlockTime().Unix()
	t := va.StartTime
	firstUnvestedPeriod := 0
	unvestedTokens := sdk.ZeroInt()
	for i, period := range va.VestingPeriods {
		t += period.Length
		if t <= now {
			firstUnvestedPeriod = i + 1
			continue
		}
		unvestedTokens = unvestedTokens.Add(period.Amount.AmountOf(bondDenom))
	}

	// Add reward to each upcoming period, in proportion of ratio of running total
	// vesting amount to the total unvested tokens
	runningTotalReward := sdk.NewCoins()
	runningTotalStaking := sdk.ZeroInt()
	for i := firstUnvestedPeriod; i < len(va.VestingPeriods); i++ {
		period := va.VestingPeriods[i]
		runningTotalStaking = runningTotalStaking.Add(period.Amount.AmountOf(bondDenom))
		runningTotRatio := runningTotalStaking.ToDec().Quo(unvestedTokens.ToDec())
		targetCoins := types.ScaleCoins(reward, runningTotRatio)
		thisReward := targetCoins.Sub(runningTotalReward)
		runningTotalReward = targetCoins
		period.Amount = period.Amount.Add(thisReward...)
		va.VestingPeriods[i] = period
	}

	// Increase original vesting amount to track added reward
	va.OriginalVesting = va.OriginalVesting.Add(reward...)
	k.accountKeeper.SetAccount(ctx, &va)
}
