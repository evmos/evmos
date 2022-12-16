package v2

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	v2types "github.com/evmos/evmos/v10/x/inflation/migrations/v2/types"
	"github.com/evmos/evmos/v10/x/inflation/types"
)

// MigrateStore migrates the x/inflation module state from the consensus version 1 to
// version 2. Specifically, it takes the parameters that are currently stored
// and managed by the Cosmos SDK params module and stores them directly into the x/inflation module state.
func MigrateStore(
	ctx sdk.Context,
	store sdk.KVStore,
	legacySubspace types.Subspace,
	cdc codec.BinaryCodec,
) error {
	var params v2types.Params
	legacySubspace.GetParamSetIfExists(ctx, &params)

	if err := params.Validate(); err != nil {
		return err
	}

	bz, err := cdc.Marshal(&params)
	if err != nil {
		return err
	}

	store.Set(v2types.ParamsKey, bz)

	return nil
}
