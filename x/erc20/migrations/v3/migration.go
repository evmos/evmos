// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package v3

import (
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	v3types "github.com/evmos/evmos/v18/x/erc20/migrations/v3/types"
	"github.com/evmos/evmos/v18/x/erc20/types"
)

var isTrue = []byte{0x01}

// MigrateStore migrates the x/erc20 module state from the consensus version 2 to
// version 3. Specifically, it takes the parameters that are currently stored
// and managed by the Cosmos SDK params module and stores them directly into the x/erc20 module state.
func MigrateStore(
	ctx sdk.Context,
	storeKey storetypes.StoreKey,
	legacySubspace types.Subspace,
) error {
	store := ctx.KVStore(storeKey)
	var params v3types.V3Params

	legacySubspace = legacySubspace.WithKeyTable(v3types.ParamKeyTable())
	legacySubspace.GetParamSet(ctx, &params)

	if err := params.Validate(); err != nil {
		return err
	}

	if params.EnableErc20 {
		store.Set(v3types.ParamStoreKeyEnableErc20, isTrue)
	}

	if params.EnableEVMHook {
		store.Set(v3types.ParamStoreKeyEnableEVMHook, isTrue)
	}

	return nil
}
