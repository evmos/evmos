package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	gogotypes "github.com/gogo/protobuf/types"

	"github.com/evmos/evmos/v10/x/erc20/types"
)

// GetParams returns the total set of erc20 parameters.
func (k Keeper) GetParams(ctx sdk.Context) (params types.Params) {
	enableErc20 := k.IsERC20Enabled(ctx)
	enableEvmHook := k.GetEnableEVMHook(ctx)

	return types.NewParams(enableErc20, enableEvmHook)
}

// SetParams sets the erc20 parameters to the param space.
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) error {
	if err := params.Validate(); err != nil {
		return err
	}

	store := ctx.KVStore(k.storeKey)

	enableEvmHookBz := k.cdc.MustMarshal(&gogotypes.BoolValue{Value: params.EnableEVMHook})
	enableErc20Bz := k.cdc.MustMarshal(&gogotypes.BoolValue{Value: params.EnableErc20})

	store.Set(types.ParamStoreKeyEnableEVMHook, enableEvmHookBz)
	store.Set(types.ParamStoreKeyEnableErc20, enableErc20Bz)

	return nil
}

// IsERC20Enabled returns true if the module logic is enabled
func (k Keeper) IsERC20Enabled(ctx sdk.Context) bool {
	var enabled gogotypes.BoolValue
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.ParamStoreKeyEnableErc20)
	if bz == nil {
		return enabled.Value
	}
	k.cdc.MustUnmarshal(bz, &enabled)
	return enabled.Value
}

// GetEnableEVMHook returns true if the EVM hooks are enabled
func (k Keeper) GetEnableEVMHook(ctx sdk.Context) bool {
	var enabled gogotypes.BoolValue
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.ParamStoreKeyEnableEVMHook)
	if bz == nil {
		return enabled.Value
	}
	k.cdc.MustUnmarshal(bz, &enabled)
	return enabled.Value
}
