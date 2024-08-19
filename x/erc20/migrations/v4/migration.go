// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package v4

import (
	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	v3types "github.com/evmos/evmos/v19/x/erc20/migrations/v3/types"
	"github.com/evmos/evmos/v19/x/erc20/types"
)

const AddressLength = 42

var isTrue = []byte{0x01}

// MigrateStore migrates the x/erc20 module state from the consensus version 3 to
// version 4. Specifically, it takes the parameters that are currently stored
// and managed by the Cosmos SDK params module and stores them directly into the x/erc20 module state.
func MigrateStore(
	ctx sdk.Context,
	storeKey storetypes.StoreKey,
) error {
	store := ctx.KVStore(storeKey)

	store.Delete(v3types.ParamStoreKeyEnableEVMHook)

	store.Get(v3types.ParamStoreKeyEnableErc20)
	store.Set(types.ParamStoreKeyEnableErc20, isTrue)

	params := types.DefaultParams()
	bz := make([]byte, 0, AddressLength*len(params.NativePrecompiles))
	for _, str := range params.NativePrecompiles {
		bz = append(bz, []byte(str)...)
	}
	store.Set(types.ParamStoreKeyNativePrecompiles, bz)

	bz = make([]byte, 0)
	store.Set(types.ParamStoreKeyDynamicPrecompiles, bz)

	return nil
}
