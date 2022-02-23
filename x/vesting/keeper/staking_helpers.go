package keeper

import (
	"math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
)

// TODO replace methods once available in the sdk `x/stake` module
// Taken from https://github.com/agoric-labs/cosmos-sdk/blob/4085-true-vesting/x/staking/keeper/delegation.go

// GetDelegatorUnbonding returns the total amount a delegator has unbonding
func (k Keeper) GetDelegatorUnbonding(
	ctx sdk.Context,
	delegator sdk.AccAddress,
) sdk.Int {
	unbonding := sdk.ZeroInt()
	k.IterateDelegatorUnbondingDelegations(ctx, delegator, func(ubd types.UnbondingDelegation) bool {
		for _, entry := range ubd.Entries {
			unbonding = unbonding.Add(entry.Balance)
		}
		return false
	})
	return unbonding
}

// GetDelegatorBonded returs the total amount a delegator has bonded
func (k Keeper) GetDelegatorBonded(
	ctx sdk.Context,
	delegator sdk.AccAddress,
) sdk.Int {
	bonded := sdk.ZeroInt()

	k.IterateDelegatorDelegations(ctx, delegator, func(delegation types.Delegation) bool {
		validatorAddr, err := sdk.ValAddressFromBech32(delegation.ValidatorAddress)
		if err != nil {
			panic(err) // shouldn't happen
		}
		validator, found := k.stakingKeeper.GetValidator(ctx, validatorAddr)
		if found {
			shares := delegation.Shares
			tokens := validator.TokensFromSharesTruncated(shares).RoundInt()
			bonded = bonded.Add(tokens)
		}
		return false
	})
	return bonded
}

// TransferUnbonding changes the ownership of UnbondingDelegation entries
// until the desired number of tokens have changed hands. Returns the actual
// number of tokens transferred.
func (k Keeper) TransferUnbonding(
	ctx sdk.Context,
	fromAddr, toAddr sdk.AccAddress,
	valAddr sdk.ValAddress,
	wantAmt sdk.Int,
) sdk.Int {
	transferred := sdk.ZeroInt()
	ubdFrom, found := k.stakingKeeper.GetUnbondingDelegation(ctx, fromAddr, valAddr)
	if !found {
		return transferred
	}
	ubdFromModified := false

	for i := 0; i < len(ubdFrom.Entries) && wantAmt.IsPositive(); i++ {
		entry := ubdFrom.Entries[i]
		toXfer := entry.Balance
		if toXfer.GT(wantAmt) {
			toXfer = wantAmt
		}
		if !toXfer.IsPositive() {
			continue
		}

		if k.stakingKeeper.HasMaxUnbondingDelegationEntries(ctx, toAddr, valAddr) {
			// TODO pre-compute the maximum entries we can add rather than checking each time
			break
		}
		ubdTo := k.stakingKeeper.SetUnbondingDelegationEntry(ctx, toAddr, valAddr, entry.CreationHeight, entry.CompletionTime, toXfer)
		k.stakingKeeper.InsertUBDQueue(ctx, ubdTo, entry.CompletionTime)
		transferred = transferred.Add(toXfer)
		wantAmt = wantAmt.Sub(toXfer)

		ubdFromModified = true
		remaining := entry.Balance.Sub(toXfer)
		if remaining.IsZero() {
			ubdFrom.RemoveEntry(int64(i))
			i--
			continue
		}
		entry.Balance = remaining
		ubdFrom.Entries[i] = entry
	}

	if ubdFromModified {
		if len(ubdFrom.Entries) == 0 {
			k.stakingKeeper.RemoveUnbondingDelegation(ctx, ubdFrom)
		} else {
			k.stakingKeeper.SetUnbondingDelegation(ctx, ubdFrom)
		}
	}
	return transferred
}

// TransferDelegation changes the ownership of at most the desired number of shares.
// Returns the actual number of shares transferred. Will also transfer redelegation
// entries to ensure that all redelegations are matched by sufficient shares.
// Note that no tokens are transferred to or from any pool or account, since no
// delegation is actually changing state.
func (k Keeper) TransferDelegation(
	ctx sdk.Context,
	fromAddr, toAddr sdk.AccAddress,
	valAddr sdk.ValAddress,
	wantShares sdk.Dec,
) sdk.Dec {
	transferred := sdk.ZeroDec()

	// sanity checks
	if !wantShares.IsPositive() {
		return transferred
	}
	validator, found := k.stakingKeeper.GetValidator(ctx, valAddr)
	if !found {
		return transferred
	}
	delFrom, found := k.stakingKeeper.GetDelegation(ctx, fromAddr, valAddr)
	if !found {
		return transferred
	}

	// Check redelegation entry limits while we can still return early.
	// Assume the worst case that we need to transfer all redelegation entries
	mightExceedLimit := false
	k.IterateDelegatorRedelegations(ctx, fromAddr, func(toRedelegation types.Redelegation) (stop bool) {
		// There's no redelegation index by delegator and dstVal or vice-versa.
		// The minimum cardinality is to look up by delegator, so scan and skip.
		if toRedelegation.ValidatorDstAddress != valAddr.String() {
			return false
		}
		fromRedelegation, found := k.stakingKeeper.GetRedelegation(ctx, fromAddr, sdk.ValAddress(toRedelegation.ValidatorSrcAddress), sdk.ValAddress(toRedelegation.ValidatorDstAddress))
		if found && len(toRedelegation.Entries)+len(fromRedelegation.Entries) >= int(k.stakingKeeper.MaxEntries(ctx)) {
			mightExceedLimit = true
			return true
		}
		return false
	})
	if mightExceedLimit {
		// avoid types.ErrMaxRedelegationEntries
		return transferred
	}

	// compute shares to transfer, amount left behind
	transferred = delFrom.Shares
	if transferred.GT(wantShares) {
		transferred = wantShares
	}
	remaining := delFrom.Shares.Sub(transferred)

	// Update or create the delTo object, calling appropriate hooks
	delTo, found := k.stakingKeeper.GetDelegation(ctx, toAddr, validator.GetOperator())
	if !found {
		delTo = types.NewDelegation(toAddr, validator.GetOperator(), sdk.ZeroDec())
	}

	if found {
		k.Hooks().BeforeDelegationSharesModified(ctx, toAddr, validator.GetOperator())
	} else {
		k.Hooks().BeforeDelegationCreated(ctx, toAddr, validator.GetOperator())
	}
	delTo.Shares = delTo.Shares.Add(transferred)
	k.stakingKeeper.SetDelegation(ctx, delTo)
	k.Hooks().AfterDelegationModified(ctx, toAddr, valAddr)

	// Update source delegation
	if remaining.IsZero() {
		k.Hooks().BeforeDelegationRemoved(ctx, fromAddr, valAddr)
		k.stakingKeeper.RemoveDelegation(ctx, delFrom)
	} else {
		k.Hooks().BeforeDelegationSharesModified(ctx, fromAddr, valAddr)
		delFrom.Shares = remaining
		k.stakingKeeper.SetDelegation(ctx, delFrom)
		k.Hooks().AfterDelegationModified(ctx, fromAddr, valAddr)
	}

	// If there are not enough remaining shares to be responsible for
	// the redelegations, transfer some redelegations.
	// For instance, if the original delegation of 300 shares to validator A
	// had redelegations for 100 shares each from validators B, C, and D,
	// and if we're transferring 175 shares, then we might keep the redelegation
	// from B, transfer the one from D, and split the redelegation from C
	// keeping a liability for 25 shares and transferring one for 75 shares.
	// Of course, the redelegations themselves can have multiple entries for
	// different timestamps, so we're actually working at a finer granularity.
	redelegations := k.stakingKeeper.GetRedelegations(ctx, fromAddr, math.MaxUint16)
	for j, redelegation := range redelegations {
		// There's no redelegation index by delegator and dstVal or vice-versa.
		// The minimum cardinality is to look up by delegator, so scan and skip.
		if redelegation.ValidatorDstAddress != valAddr.String() {
			continue
		}
		redelegationModified := false
		entriesRemaining := false
		for i := 0; i < len(redelegation.Entries); i++ {
			entry := redelegation.Entries[i]

			// Partition SharesDst between keeping and sending
			sharesToKeep := entry.SharesDst
			sharesToSend := sdk.ZeroDec()
			if entry.SharesDst.GT(remaining) {
				sharesToKeep = remaining
				sharesToSend = entry.SharesDst.Sub(sharesToKeep)
			}
			remaining = remaining.Sub(sharesToKeep) // fewer local shares available to cover liability

			if sharesToSend.IsZero() {
				// Leave the entry here
				entriesRemaining = true
				continue
			}
			if sharesToKeep.IsZero() {
				// Transfer the whole entry, delete locally
				toRed := k.stakingKeeper.SetRedelegationEntry(
					ctx, toAddr, sdk.ValAddress(redelegation.ValidatorSrcAddress),
					sdk.ValAddress(redelegation.ValidatorDstAddress),
					entry.CreationHeight, entry.CompletionTime, entry.InitialBalance, sdk.ZeroDec(), sharesToSend,
				)
				k.stakingKeeper.InsertRedelegationQueue(ctx, toRed, entry.CompletionTime)
				(&redelegations[j]).RemoveEntry(int64(i))
				i--
				// okay to leave an obsolete entry in the queue for the removed entry
				redelegationModified = true
			} else {
				// Proportionally divide the entry
				fracSending := sharesToSend.Quo(entry.SharesDst)
				balanceToSend := fracSending.MulInt(entry.InitialBalance).TruncateInt()
				balanceToKeep := entry.InitialBalance.Sub(balanceToSend)
				toRed := k.stakingKeeper.SetRedelegationEntry(
					ctx, toAddr, sdk.ValAddress(redelegation.ValidatorSrcAddress),
					sdk.ValAddress(redelegation.ValidatorDstAddress),
					entry.CreationHeight, entry.CompletionTime, balanceToSend, sdk.ZeroDec(), sharesToSend,
				)
				k.stakingKeeper.InsertRedelegationQueue(ctx, toRed, entry.CompletionTime)
				entry.InitialBalance = balanceToKeep
				entry.SharesDst = sharesToKeep
				redelegation.Entries[i] = entry
				// not modifying the completion time, so no need to change the queue
				redelegationModified = true
				entriesRemaining = true
			}
		}
		if redelegationModified {
			if entriesRemaining {
				k.stakingKeeper.SetRedelegation(ctx, redelegation)
			} else {
				k.stakingKeeper.RemoveRedelegation(ctx, redelegation)
			}
		}
	}
	return transferred
}

// iterate through a delegator's unbonding delegations
func (k Keeper) IterateDelegatorUnbondingDelegations(
	ctx sdk.Context,
	delegator sdk.AccAddress,
	cb func(ubd types.UnbondingDelegation) (stop bool),
) {
	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, types.GetUBDsKey(delegator))
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		ubd := types.MustUnmarshalUBD(k.cdc, iterator.Value())
		if cb(ubd) {
			break
		}
	}
}

// IterateDelegatorDelegations iterates through one delegator's delegations
func (k Keeper) IterateDelegatorDelegations(
	ctx sdk.Context,
	delegator sdk.AccAddress,
	cb func(delegation types.Delegation) (stop bool),
) {
	store := ctx.KVStore(k.storeKey)
	delegatorPrefixKey := types.GetDelegationsKey(delegator)
	iterator := sdk.KVStorePrefixIterator(store, delegatorPrefixKey)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		delegation := types.MustUnmarshalDelegation(k.cdc, iterator.Value())
		if cb(delegation) {
			break
		}
	}
}

// iterate through one delegator's redelegations
func (k Keeper) IterateDelegatorRedelegations(
	ctx sdk.Context,
	delegator sdk.AccAddress,
	fn func(red types.Redelegation) (stop bool),
) {
	store := ctx.KVStore(k.storeKey)
	delegatorPrefixKey := types.GetREDsKey(delegator)

	iterator := sdk.KVStorePrefixIterator(store, delegatorPrefixKey)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		red := types.MustUnmarshalRED(k.cdc, iterator.Value())
		if stop := fn(red); stop {
			break
		}
	}
}
