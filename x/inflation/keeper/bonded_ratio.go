package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/tharsis/evmos/x/inflation/types"
)

// GetBondedRatio gets the current BondedRatio
func (k Keeper) GetBondedRatio(ctx sdk.Context) (sdk.Dec, bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.KeyPrefixBondedRatio)
	if len(bz) == 0 {
		return sdk.OneDec(), false
	}

	var bondedRatio sdk.Dec
	err := bondedRatio.Unmarshal(bz)
	if err != nil {
		panic(fmt.Errorf("unable to unmarshal bondedRatio value: %w", err))
	}

	return bondedRatio, true
}

// SetBondedRatio sets the current BondedRatio
func (k Keeper) SetBondedRatio(ctx sdk.Context, bondedRatio sdk.Dec) {
	bz, err := bondedRatio.Marshal()
	if err != nil {
		panic(fmt.Errorf("unable to marshal amount value: %w", err))
	}

	store := ctx.KVStore(k.storeKey)
	store.Set(types.KeyPrefixBondedRatio, bz)
}
