// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package v8

import (
	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/evmos/evmos/v19/x/evm/types"
)

// MigrateStore migrates the x/evm module state from the consensus version 7 to
// version 8. Specifically, it adds the evm denom decimals to the module params.
func MigrateStore(
	ctx sdk.Context,
	storeKey storetypes.StoreKey,
	cdc codec.BinaryCodec,
) error {
	var params types.Params

	store := ctx.KVStore(storeKey)

	bz := store.Get(types.KeyPrefixParams)
	cdc.MustUnmarshal(bz, &params)

	params.DenomDecimals = types.DefaultDenomDecimals

	if err := params.Validate(); err != nil {
		return err
	}

	bz = cdc.MustMarshal(&params)

	store.Set(types.KeyPrefixParams, bz)

	return nil
}
