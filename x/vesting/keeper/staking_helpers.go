package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
)

// TODO replace methods once available in the sdk `x/staking` module
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
