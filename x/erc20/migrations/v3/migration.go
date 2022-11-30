package v3

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/evmos/evmos/v10/x/erc20/types"
)

// MigrateStore migrates the x/erc20 module state from the consensus version 2 to
// version 3. Specifically, it takes the parameters that are currently stored
// and managed by the Cosmos SDK params module and stores them directly into the x/erc20 module state.
func MigrateStore(ctx sdk.Context,
	store sdk.KVStore,
	legacySubspace types.Subspace,
) error {
	var params types.Params
	legacySubspace.GetParamSet(ctx, &params)

	if err := params.Validate(); err != nil {
		return err
	}

	if params.EnableErc20 {
		store.Set(types.ParamStoreKeyEnableErc20, []byte("0x01"))
	}

	if params.EnableEVMHook {
		store.Set(types.ParamStoreKeyEnableEVMHook, []byte("0x01"))
	}

	return nil
}
