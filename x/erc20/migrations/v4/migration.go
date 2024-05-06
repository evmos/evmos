// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package v4

import (
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	v4types "github.com/evmos/evmos/v18/x/erc20/migrations/v4/types"
	"github.com/evmos/evmos/v18/x/erc20/types"
)

// MigrateStore migrates the x/erc20 module state from the consensus version 3 to
// version 4. Specifically, it deletes old parameters and adds the new ones into the x/erc20 module state.
func MigrateStore(
	ctx sdk.Context,
	storeKey storetypes.StoreKey,
) error {
	store := ctx.KVStore(storeKey)
	store.Delete(v4types.ParamStoreKeyEnableEVMHook)

	// empty bytes
	bz := []byte{}
	// Set both arrays as empty
	store.Set(types.ParamStoreKeyDynamicPrecompiles, bz)

	store.Set(types.ParamStoreKeyNativePrecompiles, bz)
	return nil
}
