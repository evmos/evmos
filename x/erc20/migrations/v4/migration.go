// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package v4

import (
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// MigrateStore migrates the x/erc20 module state from the consensus version 2 to
// version 3. Specifically, it takes the parameters that are currently stored
// and managed by the Cosmos SDK params module and stores them directly into the x/erc20 module state.
func MigrateStore(
	ctx sdk.Context,
	storeKey storetypes.StoreKey,
) error {
	// store := ctx.KVStore(storeKey)
	// store.Delete(types.ParamStoreKeyEnableEVMHook)

	return nil
}
