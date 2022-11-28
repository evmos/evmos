package v3

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/evmos/evmos/v10/x/erc20/types"
)

// MigrateStore migrates the x/erc20 module state from the consensus version 2 to
// version 3. Specifically, it takes the parameters that are currently stored
// and managed by the Cosmos SDK params module and stores them directly into the x/erc20 module state.
func MigrateStore(ctx sdk.Context,
	store sdk.KVStore,
	legacySubspace types.Subspace,
	cdc codec.BinaryCodec,
) error {
	var params types.Params
	legacySubspace.GetParamSet(ctx, &params)

	if err := params.Validate(); err != nil {
		return err
	}

	enableErc20Bz := []byte("0x00")
	if params.EnableErc20 {
		enableErc20Bz = []byte("0x01")
	}

	enableEvmHookBz := []byte("0x00")
	if params.EnableEVMHook {
		enableEvmHookBz = []byte("0x01")
	}

	store.Set(types.ParamStoreKeyEnableErc20, enableErc20Bz)
	store.Set(types.ParamStoreKeyEnableEVMHook, enableEvmHookBz)

	return nil
}
