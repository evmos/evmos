package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/tharsis/evmos/v5/x/inflation/types"
)

// GetEpochMintProvision gets the current EpochMintProvision
func (k Keeper) GetEpochMintProvision(ctx sdk.Context) (sdk.Dec, bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.KeyPrefixEpochMintProvision)
	if len(bz) == 0 {
		return sdk.ZeroDec(), false
	}

	var epochMintProvision sdk.Dec
	err := epochMintProvision.Unmarshal(bz)
	if err != nil {
		panic(fmt.Errorf("unable to unmarshal epochMintProvision value: %w", err))
	}

	return epochMintProvision, true
}

// SetEpochMintProvision sets the current EpochMintProvision
func (k Keeper) SetEpochMintProvision(ctx sdk.Context, epochMintProvision sdk.Dec) {
	bz, err := epochMintProvision.Marshal()
	if err != nil {
		panic(fmt.Errorf("unable to marshal amount value: %w", err))
	}

	store := ctx.KVStore(k.storeKey)
	store.Set(types.KeyPrefixEpochMintProvision, bz)
}
