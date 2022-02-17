package keeper

import (
	"math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkvesting "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"

	"github.com/tharsis/evmos/x/vesting/types"
)

// -------------------------------------------------------------------------- //
// MOVED FROM AGORIC TYPES BECAUSE THEY ACCESS KEEPERS

// AddGrant merges a new clawback vesting grant into an existing ClawbackVestingAccount.
func (k Keeper) AddGrantToClawbackVestingAccount(ctx sdk.Context, va *types.ClawbackVestingAccount, grantStartTime int64, grantLockupPeriods, grantVestingPeriods []sdkvesting.Period, grantCoins sdk.Coins) {
	// how much is really delegated?
	bondedAmt := k.GetDelegatorBonded(ctx, va.GetAddress())
	unbondingAmt := k.GetDelegatorUnbonding(ctx, va.GetAddress())
	delegatedAmt := bondedAmt.Add(unbondingAmt)
	delegated := sdk.NewCoins(sdk.NewCoin(k.stakingKeeper.BondDenom(ctx), delegatedAmt))

	// discover what has been slashed
	oldDelegated := va.DelegatedVesting.Add(va.DelegatedFree...)
	slashed := oldDelegated.Sub(types.CoinsMin(oldDelegated, delegated))

	// Absorb the slashed amount by eliminating the tail of the vesting and lockup schedules
	unvestedSlashed := types.CoinsMin(slashed, va.OriginalVesting)
	if !unvestedSlashed.IsZero() {
		newOrigVesting := va.OriginalVesting.Sub(unvestedSlashed)
		cutoffPeriods := []sdkvesting.Period{{Length: 1, Amount: newOrigVesting}}
		start := va.GetStartTime()
		_, newLockupEnd, newLockupPeriods := types.ConjunctPeriods(start, start, va.LockupPeriods, cutoffPeriods)
		_, newVestingEnd, newVestingPeriods := types.ConjunctPeriods(start, start, va.VestingPeriods, cutoffPeriods)
		va.OriginalVesting = newOrigVesting
		va.EndTime = types.Max64(newLockupEnd, newVestingEnd)
		va.LockupPeriods = newLockupPeriods
		va.VestingPeriods = newVestingPeriods
	}

	// modify schedules for the new grant
	newLockupStart, newLockupEnd, newLockupPeriods := types.DisjunctPeriods(va.StartTime, grantStartTime, va.LockupPeriods, grantLockupPeriods)
	newVestingStart, newVestingEnd, newVestingPeriods := types.DisjunctPeriods(va.StartTime, grantStartTime,
		va.GetVestingPeriods(), grantVestingPeriods)
	if newLockupStart != newVestingStart {
		panic("bad start time calculation")
	}
	va.StartTime = newLockupStart
	va.EndTime = types.Max64(newLockupEnd, newVestingEnd)
	va.LockupPeriods = newLockupPeriods
	va.VestingPeriods = newVestingPeriods
	va.OriginalVesting = va.OriginalVesting.Add(grantCoins...)

	// cap DV at the current unvested amount, DF rounds out to current delegated
	unvested := va.GetVestingCoins(ctx.BlockTime())
	va.DelegatedVesting = types.CoinsMin(delegated, unvested)
	va.DelegatedFree = delegated.Sub(va.DelegatedVesting)
}

// Clawback transfers unvested tokens in a ClawbackVestingAccount to dest.
// Future vesting events are removed. Unstaked tokens are simply sent.
// Unbonding and staked tokens are transferred with their staking state
// intact.  Account state is updated to reflect the removals.
func (k Keeper) ClawbackFromClawbackVestingAccount(ctx sdk.Context, va types.ClawbackVestingAccount, dest sdk.AccAddress) error {
	// Compute the clawback based on the account state only, and update account
	updatedAcc, toClawBack := va.ComputeClawback(ctx.BlockTime().Unix())
	if toClawBack.IsZero() {
		return nil
	}
	addr := updatedAcc.GetAddress()
	bondDenom := k.stakingKeeper.BondDenom(ctx)

	// Compute the clawback based on bank balance and delegation, and update account
	encumbered := updatedAcc.GetVestingCoins(ctx.BlockTime())
	bondedAmt := k.GetDelegatorBonded(ctx, addr)
	unbondingAmt := k.GetDelegatorUnbonding(ctx, addr)
	bonded := sdk.NewCoins(sdk.NewCoin(bondDenom, bondedAmt))
	unbonding := sdk.NewCoins(sdk.NewCoin(bondDenom, unbondingAmt))
	unbonded := k.bankKeeper.GetAllBalances(ctx, addr)
	updatedAcc, toClawBack = updatedAcc.UpdateDelegation(encumbered, toClawBack, bonded, unbonding, unbonded)

	// Write now now so that the bank module sees unvested tokens are unlocked.
	// Note that all store writes are aborted if there is a panic, so there is
	// no danger in writing incomplete results.
	k.accountKeeper.SetAccount(ctx, &updatedAcc)

	// Now that future vesting events (and associated lockup) are removed,
	// the balance of the account is unlocked and can be freely transferred.
	spendable := k.bankKeeper.SpendableCoins(ctx, addr)
	toXfer := types.CoinsMin(toClawBack, spendable)
	err := k.bankKeeper.SendCoins(ctx, addr, dest, toXfer)
	if err != nil {
		return err // shouldn't happen, given spendable check
	}
	toClawBack = toClawBack.Sub(toXfer)

	// We need to traverse the staking data structures to update the
	// vesting account bookkeeping, and to recover more funds if necessary.
	// Staking is the only way unvested tokens should be missing from the bank balance.

	// If we need more, transfer UnbondingDelegations.
	want := toClawBack.AmountOf(bondDenom)
	unbondings := k.stakingKeeper.GetUnbondingDelegations(ctx, addr, math.MaxUint16)
	for _, unbonding := range unbondings {
		valAddr, err := sdk.ValAddressFromBech32(unbonding.ValidatorAddress)
		if err != nil {
			panic(err)
		}
		transferred := k.TransferUnbonding(ctx, addr, dest, valAddr, want)
		want = want.Sub(transferred)
		if !want.IsPositive() {
			break
		}
	}

	// If we need more, transfer Delegations.
	if want.IsPositive() {
		delegations := k.stakingKeeper.GetDelegatorDelegations(ctx, addr, math.MaxUint16)
		for _, delegation := range delegations {
			validatorAddr, err := sdk.ValAddressFromBech32(delegation.ValidatorAddress)
			if err != nil {
				panic(err) // shouldn't happen
			}
			validator, found := k.stakingKeeper.GetValidator(ctx, validatorAddr)
			if !found {
				// validator has been removed
				continue
			}
			wantShares, err := validator.SharesFromTokensTruncated(want)
			if err != nil {
				// validator has no tokens
				continue
			}
			transferredShares := k.TransferDelegation(ctx, addr, dest, delegation.GetValidatorAddr(), wantShares)
			// to be conservative in what we're clawing back, round transferred shares up
			transferred := validator.TokensFromSharesRoundUp(transferredShares).RoundInt()
			want = want.Sub(transferred)
			if !want.IsPositive() {
				// Could be slightly negative, due to rounding?
				// Don't think so, due to the precautions above.
				break
			}
		}
	}

	// If we've transferred everything and still haven't transferred the desired clawback amount,
	// then the account must have most some unvested tokens from slashing.
	return nil
}

// PostReward encumbers a previously-deposited reward according to the current vesting apportionment of staking.
// Note that rewards might be unvested, but are unlocked.
func (k Keeper) PostReward(ctx sdk.Context, va types.ClawbackVestingAccount, reward sdk.Coins) {
	// Find the scheduled amount of vested and unvested staking tokens
	bondDenom := k.stakingKeeper.BondDenom(ctx)
	vested := types.ReadSchedule(va.StartTime, va.EndTime, va.VestingPeriods, va.OriginalVesting, ctx.BlockTime().Unix()).AmountOf(bondDenom)
	unvested := va.OriginalVesting.AmountOf(bondDenom).Sub(vested)

	if unvested.IsZero() {
		// no need to adjust the vesting schedule
		return
	}

	if vested.IsZero() {
		// all staked tokens must be unvested
		k.DistributeReward(ctx, va, bondDenom, reward)
		return
	}

	// Find current split of account balance on staking axis
	bonded := k.GetDelegatorBonded(ctx, va.GetAddress())
	unbonding := k.GetDelegatorUnbonding(ctx, va.GetAddress())
	unbonded := k.bankKeeper.GetBalance(ctx, va.GetAddress(), bondDenom).Amount
	total := bonded.Add(unbonding).Add(unbonded)
	total = total.Sub(types.MinInt(total, reward.AmountOf(bondDenom))) // look at pre-reward total

	// Adjust vested/unvested for the actual amount in the account (transfers, slashing)
	// preferring them to be unvested
	unvested = types.MinInt(unvested, total) // may have been reduced by slashing
	vested = total.Sub(unvested)

	// Now restrict to just the bonded tokens, preferring them to be vested
	vested = types.MinInt(vested, bonded)
	unvested = bonded.Sub(vested)

	// Compute the unvested amount of reward and add to vesting schedule
	if unvested.IsZero() {
		return
	}
	if vested.IsZero() {
		k.DistributeReward(ctx, va, bondDenom, reward)
		return
	}
	unvestedRatio := unvested.ToDec().QuoTruncate(bonded.ToDec()) // round down
	unvestedReward := types.ScaleCoins(reward, unvestedRatio)
	k.DistributeReward(ctx, va, bondDenom, unvestedReward)
}

// distributeReward adds the reward to the future vesting schedule in proportion to the future vesting
// staking tokens.
func (k Keeper) DistributeReward(ctx sdk.Context, va types.ClawbackVestingAccount, bondDenom string, reward sdk.Coins) {
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

	runningTotReward := sdk.NewCoins()
	runningTotStaking := sdk.ZeroInt()
	for i := firstUnvestedPeriod; i < len(va.VestingPeriods); i++ {
		period := va.VestingPeriods[i]
		runningTotStaking = runningTotStaking.Add(period.Amount.AmountOf(bondDenom))
		runningTotRatio := runningTotStaking.ToDec().Quo(unvestedTokens.ToDec())
		targetCoins := types.ScaleCoins(reward, runningTotRatio)
		thisReward := targetCoins.Sub(runningTotReward)
		runningTotReward = targetCoins
		period.Amount = period.Amount.Add(thisReward...)
		va.VestingPeriods[i] = period
	}

	va.OriginalVesting = va.OriginalVesting.Add(reward...)
	k.accountKeeper.SetAccount(ctx, &va)
}
