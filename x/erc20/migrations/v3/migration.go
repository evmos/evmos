package v3

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	v3types "github.com/evmos/evmos/v10/x/erc20/migrations/v3/types"
	"github.com/evmos/evmos/v10/x/erc20/types"
	gogotypes "github.com/gogo/protobuf/types"
)

// MigrateStore migrates the x/erc20 module state from the consensus version 2 to
// version 3. Specifically, it takes the parameters that are currently stored
// and managed by the Cosmos SDK params module and stores them directly into the x/erc20 module state.
func MigrateStore(ctx sdk.Context,
	store sdk.KVStore,
	legacySubspace types.Subspace,
	cdc codec.BinaryCodec) error {
	var params types.Params
	legacySubspace.GetParamSet(ctx, &params)

	if err := params.Validate(); err != nil {
		return err
	}

	enableErc20Bz := cdc.MustMarshal(&gogotypes.BoolValue{Value: params.EnableErc20})
	enableEvmHookBz := cdc.MustMarshal(&gogotypes.BoolValue{Value: params.EnableEVMHook})

	store.Set(v3types.ParamStoreKeyEnableErc20, enableErc20Bz)
	store.Set(v3types.ParamStoreKeyEnableEVMHook, enableEvmHookBz)

	return nil
}
