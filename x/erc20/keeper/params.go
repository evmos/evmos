// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/evmos/evmos/v18/x/erc20/types"
)

var isTrue = []byte("0x01")

const addressLength = 42

// GetParams returns the total set of erc20 parameters.
func (k Keeper) GetParams(ctx sdk.Context) (params types.Params) {
	enableErc20 := k.IsERC20Enabled(ctx)
	return types.NewParams(enableErc20)
}

// SetParams sets the erc20 parameters to the param space.
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) error {
	if err := params.Validate(); err != nil {
		return err
	}

	k.setERC20Enabled(ctx, params.EnableErc20)
	return nil
}

// IsERC20Enabled returns true if the module logic is enabled
func (k Keeper) IsERC20Enabled(ctx sdk.Context) bool {
	store := ctx.KVStore(k.storeKey)
	return store.Has(types.ParamStoreKeyEnableErc20)
}

// setERC20Enabled sets the EnableERC20 param in the store
func (k Keeper) setERC20Enabled(ctx sdk.Context, enable bool) {
	store := ctx.KVStore(k.storeKey)
	if enable {
		store.Set(types.ParamStoreKeyEnableErc20, isTrue)
		return
	}
	store.Delete(types.ParamStoreKeyEnableErc20)
}
