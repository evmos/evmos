// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package v3

import (
	"github.com/cosmos/cosmos-sdk/codec"

	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	v3types "github.com/evmos/evmos/v14/x/claims/migrations/v3/types"
	"github.com/evmos/evmos/v14/x/claims/types"
)

// MigrateStore migrates the x/claims module state from the consensus version 2 to
// version 3. Specifically, it takes the parameters that are currently stored
// and managed by the Cosmos SDK params module and stores them directly into the x/claims module state.
func MigrateStore(
	ctx sdk.Context,
	storeKey storetypes.StoreKey,
	legacySubspace types.Subspace,
	cdc codec.BinaryCodec,
) error {
	store := ctx.KVStore(storeKey)
	var params v3types.V3Params

	legacySubspace = legacySubspace.WithKeyTable(v3types.ParamKeyTable())
	legacySubspace.GetParamSetIfExists(ctx, &params)

	if err := params.Validate(); err != nil {
		return err
	}

	bz, err := cdc.Marshal(&params)
	if err != nil {
		return err
	}

	store.Set(types.ParamsKey, bz)

	return nil
}
