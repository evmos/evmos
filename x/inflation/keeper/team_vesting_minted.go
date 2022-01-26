package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/tharsis/evmos/x/inflation/types"
)

// IsTeamVestingMinted returns true if the team vesting amount has already been
// minted
func (k Keeper) IsTeamVestingMinted(ctx sdk.Context) bool {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.KeyPrefixTeamVestingMinted)
	if len(bz) == 0 {
		return false
	}

	return sdk.BigEndianToUint64(bz) == 1
}

// SetPeriod stores the current period
func (k Keeper) SetTeamVestingMinted(ctx sdk.Context, teamVestingMinted bool) {
	store := ctx.KVStore(k.storeKey)
	var bz []byte
	if teamVestingMinted {
		bz = sdk.Uint64ToBigEndian(1)
	} else {
		bz = sdk.Uint64ToBigEndian(0)
	}

	store.Set(types.KeyPrefixTeamVestingMinted, bz)
}
